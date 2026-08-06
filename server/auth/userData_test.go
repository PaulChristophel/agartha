package auth

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type fakeLDAPConnection struct {
	binds         [][2]string
	closed        bool
	searchRequest *ldap.SearchRequest
	searchResult  *ldap.SearchResult
}

func (f *fakeLDAPConnection) Bind(username, password string) error {
	f.binds = append(f.binds, [2]string{username, password})
	return nil
}

func (f *fakeLDAPConnection) Close() error {
	f.closed = true
	return nil
}

func (f *fakeLDAPConnection) Search(request *ldap.SearchRequest) (*ldap.SearchResult, error) {
	f.searchRequest = request
	return f.searchResult, nil
}

func (f *fakeLDAPConnection) StartTLS(*tls.Config) error {
	return nil
}

func TestAuthRejectsUnsupportedMethod(t *testing.T) {
	for _, method := range []string{"", "unknown", "LOCAL", " ldap"} {
		t.Run(method, func(t *testing.T) {
			_, err := auth(credentials{Method: method}, nil)
			require.ErrorContains(t, err, "unsupported authentication method")
		})
	}
}

func TestAuthRejectsDisabledMethod(t *testing.T) {
	originalMethods := enabledMethods
	enabledMethods = map[string]struct{}{"local": {}}
	t.Cleanup(func() { enabledMethods = originalMethods })

	_, err := auth(credentials{Method: "ldap"}, nil)
	require.ErrorContains(t, err, "not enabled")
}

func TestGetMethodReturnsOnlyEnabledMethods(t *testing.T) {
	originalMethods := enabledMethods
	enabledMethods = map[string]struct{}{"local": {}, "cas": {}}
	t.Cleanup(func() { enabledMethods = originalMethods })

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	GetMethod(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"auth_methods":["local","cas"]}`, recorder.Body.String())
}

func TestNormalizeLDAPUsername(t *testing.T) {
	tests := []struct {
		name          string
		username      string
		defaultDomain string
		wantAccount   string
		wantUPN       string
		wantError     bool
	}{
		{
			name:          "qualified username keeps account before at sign",
			username:      "user@example.com",
			defaultDomain: "fallback.example.com",
			wantAccount:   "user",
			wantUPN:       "user@example.com",
		},
		{
			name:          "unqualified username uses default domain",
			username:      "user",
			defaultDomain: "example.com",
			wantAccount:   "user",
			wantUPN:       "user@example.com",
		},
		{name: "rejects empty account", username: "@example.com", defaultDomain: "example.com", wantError: true},
		{name: "rejects empty explicit domain", username: "user@", defaultDomain: "example.com", wantError: true},
		{name: "rejects missing default domain", username: "user", wantError: true},
		{name: "rejects multiple separators", username: "user@example.com@invalid", defaultDomain: "example.com", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, upn, err := normalizeLDAPUsername(tt.username, tt.defaultDomain)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantAccount, account)
			require.Equal(t, tt.wantUPN, upn)
		})
	}
}

func TestAuthLDAPBindsAndUsesDirectoryIdentity(t *testing.T) {
	_, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)

	connection := &fakeLDAPConnection{
		searchResult: &ldap.SearchResult{Entries: []*ldap.Entry{
			ldap.NewEntry("cn=Alice,dc=example,dc=test", map[string][]string{
				"sAMAccountName": {"directory-alice"},
				"givenName":      {"Alice"},
				"sn":             {"Admin"},
				"mail":           {"alice@example.test"},
			}),
		}},
	}
	originalOptions := ldapOptions
	originalDial := ldapDialURL
	ldapOptions = config.LDAPOptions{
		Server:            "ldaps://directory.example.test:636",
		User:              "cn=service,dc=example,dc=test",
		Password:          "service-password",
		BaseDN:            "dc=example,dc=test",
		Filter:            "(&(objectClass=person)(sAMAccountName=%s))",
		LDAPDomainDefault: "example.test",
	}
	ldapDialURL = func(server string) (ldapConnection, error) {
		require.Equal(t, ldapOptions.Server, server)
		return connection, nil
	}
	t.Cleanup(func() {
		ldapOptions = originalOptions
		ldapDialURL = originalDial
	})

	user, err := authLDAP("requested-alice", "user-password")
	require.NoError(t, err)
	require.Equal(t, "directory-alice", user.Username)
	require.Equal(t, "alice@example.test", user.Email)
	require.Equal(t, [][2]string{
		{"requested-alice@example.test", "user-password"},
		{"cn=service,dc=example,dc=test", "service-password"},
	}, connection.binds)
	require.Equal(t, ldapOptions.BaseDN, connection.searchRequest.BaseDN)
	require.Equal(t, "(&(objectClass=person)(sAMAccountName=requested-alice))", connection.searchRequest.Filter)
	require.True(t, connection.closed)
}

func TestAuthLDAPRejectsEmptyPasswordBeforeConnecting(t *testing.T) {
	originalOptions := ldapOptions
	originalDial := ldapDialURL
	ldapOptions.LDAPDomainDefault = "example.test"
	ldapDialURL = func(string) (ldapConnection, error) {
		t.Fatal("LDAP connection should not be opened for an empty password")
		return nil, nil
	}
	t.Cleanup(func() {
		ldapOptions = originalOptions
		ldapDialURL = originalDial
	})

	_, err := authLDAP("alice", "")
	require.ErrorContains(t, err, "password is empty")
}

func TestAuthLocalUsesStoredPassword(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	originalDB := db.DB
	db.DB = gormDB
	t.Cleanup(func() { db.DB = originalDB })

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, username, first_name, last_name, email, is_active
		FROM auth_user
		WHERE username = $1
		  AND is_active = TRUE
		  AND password = crypt($2, password)
		LIMIT 1
	`)).
		WithArgs("alice", "correct-password").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "first_name", "last_name", "email", "is_active"}).
			AddRow(7, "alice", "Alice", "Admin", "alice@example.com", true))

	user, err := authLocal("alice", "correct-password")
	require.NoError(t, err)
	require.Equal(t, "alice", user.Username)
	require.Equal(t, "alice@example.com", user.Email)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthLocalRejectsFormerDemonstrationCredentials(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	originalDB := db.DB
	db.DB = gormDB
	t.Cleanup(func() { db.DB = originalDB })

	mock.ExpectQuery("SELECT id, username, first_name, last_name, email, is_active").
		WithArgs("localuser", "localpassword").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "first_name", "last_name", "email", "is_active"}))

	_, err = authLocal("localuser", "localpassword")
	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthCASRequiresAuthenticationSuccess(t *testing.T) {
	_, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)

	tests := []struct {
		name         string
		responseBody string
		wantUser     string
		wantError    bool
	}{
		{
			name: "accepts asserted user",
			responseBody: `<?xml version="1.0"?>
				<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
				  <cas:authenticationSuccess><cas:user>cas-user</cas:user></cas:authenticationSuccess>
				</cas:serviceResponse>`,
			wantUser: "cas-user",
		},
		{
			name: "rejects HTTP 200 failure response",
			responseBody: `<?xml version="1.0"?>
				<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
				  <cas:authenticationFailure code="INVALID_TICKET">invalid ticket</cas:authenticationFailure>
				</cas:serviceResponse>`,
			wantError: true,
		},
		{
			name:         "rejects malformed XML response",
			responseBody: `<cas:serviceResponse><cas:authenticationSuccess>`,
			wantError:    true,
		},
		{
			name: "rejects success without an asserted user",
			responseBody: `<?xml version="1.0"?>
				<cas:serviceResponse xmlns:cas="http://www.yale.edu/tp/cas">
				  <cas:authenticationSuccess><cas:user> </cas:user></cas:authenticationSuccess>
				</cas:serviceResponse>`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				require.Equal(t, "/cas/serviceValidate", r.URL.Path)
				require.Equal(t, "ticket-value", r.URL.Query().Get("ticket"))
				require.Equal(t, "https://agartha.example.com/cas", r.URL.Query().Get("service"))
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"application/xml"}},
					Body:       io.NopCloser(strings.NewReader(tt.responseBody)),
					Request:    r,
				}, nil
			})}

			originalOptions := casOptions
			originalClient := casHTTPClient
			casOptions = config.CASOptions{
				Server:       "https://cas.example.test/cas",
				ServiceURL:   "https://agartha.example.com/cas",
				ValidatePath: "/serviceValidate",
			}
			casHTTPClient = client
			t.Cleanup(func() {
				casOptions = originalOptions
				casHTTPClient = originalClient
			})

			context, _ := gin.CreateTestContext(httptest.NewRecorder())
			context.Request = httptest.NewRequest(http.MethodPost, "/auth/token?ticket=ticket-value", nil)
			user, err := authCAS("requested-user", context)
			if tt.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantUser, user.Username)
		})
	}
}
