package auth

import (
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
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestAuthRejectsUnsupportedMethod(t *testing.T) {
	_, err := auth(credentials{Method: "unknown"}, nil)
	require.ErrorContains(t, err, "unsupported authentication method")
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
