package routes

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const (
	routeAuthorizationSecret    = "route-authorization-test-secret"
	routeAuthorizationUsername  = "route-test-user"
	routeAuthorizationSaltToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

type authorizationRouteCase struct {
	name               string
	method             string
	path               string
	body               string
	allowedPermissions string
	deniedPermissions  string
	allowedStatus      int
	expectSaltKeys     bool
	useSaltToken       bool
}

type closeNotifyRecorder struct {
	*httptest.ResponseRecorder
}

func (recorder *closeNotifyRecorder) CloseNotify() <-chan bool {
	return make(chan bool)
}

func TestSensitiveRoutesEnforceSaltAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)
	upstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	tests := []authorizationRouteCase{
		{
			name:               "v1 Salt data read",
			method:             http.MethodGet,
			path:               "/api/v1/conformity/refresh",
			allowedPermissions: `["@jobs"]`,
			deniedPermissions:  `[]`,
			allowedStatus:      http.StatusOK,
		},
		{
			name:               "v1 Salt command execution",
			method:             http.MethodPost,
			path:               "/api/v1/salt_cache/",
			allowedPermissions: `["test.ping"]`,
			deniedPermissions:  `["@jobs"]`,
			allowedStatus:      http.StatusBadRequest,
		},
		{
			name:               "netapi Salt data read",
			method:             http.MethodGet,
			path:               "/api/v1/netapi/stats",
			allowedPermissions: `["@jobs"]`,
			deniedPermissions:  `[]`,
			allowedStatus:      http.StatusNoContent,
			useSaltToken:       true,
		},
		{
			name:               "netapi Salt command execution",
			method:             http.MethodPost,
			path:               "/api/v1/netapi/hook",
			allowedPermissions: `["test.ping"]`,
			deniedPermissions:  `["@jobs"]`,
			allowedStatus:      http.StatusNoContent,
			useSaltToken:       true,
		},
		{
			name:               "key.list_all",
			method:             http.MethodGet,
			path:               "/api/v1/salt_keys/minion_keys",
			allowedPermissions: `[{"@wheel":["key.list_all"]}]`,
			deniedPermissions:  `[{"@wheel":["key.accept"]}]`,
			allowedStatus:      http.StatusNotFound,
			expectSaltKeys:     true,
		},
		{
			name:               "key.accept",
			method:             http.MethodPost,
			path:               "/api/v1/salt_keys/minion_keys/accept",
			allowedPermissions: `[{"@wheel":["key.accept"]}]`,
			deniedPermissions:  `[{"@wheel":["key.reject"]}]`,
			allowedStatus:      http.StatusBadRequest,
		},
		{
			name:               "key.reject",
			method:             http.MethodPost,
			path:               "/api/v1/salt_keys/minion_keys/reject",
			allowedPermissions: `[{"@wheel":["key.reject"]}]`,
			deniedPermissions:  `[{"@wheel":["key.delete"]}]`,
			allowedStatus:      http.StatusBadRequest,
		},
		{
			name:               "key.delete",
			method:             http.MethodPost,
			path:               "/api/v1/salt_keys/minion_keys/delete",
			allowedPermissions: `[{"@wheel":["key.delete"]}]`,
			deniedPermissions:  `[{"@wheel":["key.list_all"]}]`,
			allowedStatus:      http.StatusBadRequest,
		},
		{
			name:               "raw salt_keys read",
			method:             http.MethodGet,
			path:               "/api/v1/salt_keys",
			allowedPermissions: `["@wheel"]`,
			deniedPermissions:  `[{"@wheel":["key.list_all"]}]`,
			allowedStatus:      http.StatusNotFound,
			expectSaltKeys:     true,
		},
		{
			name:               "raw salt_keys write",
			method:             http.MethodPost,
			path:               "/api/v1/salt_keys",
			allowedPermissions: `["@wheel"]`,
			deniedPermissions:  `[{"@wheel":["key.accept"]}]`,
			allowedStatus:      http.StatusNotFound,
			expectSaltKeys:     true,
		},
		{
			name:               "raw salt_keys delete",
			method:             http.MethodDelete,
			path:               "/api/v1/salt_keys/pki/master/key.pem",
			allowedPermissions: `["@wheel"]`,
			deniedPermissions:  `[{"@wheel":["key.delete"]}]`,
			allowedStatus:      http.StatusNotFound,
			expectSaltKeys:     true,
		},
		{
			name:               "v2 Salt cache read",
			method:             http.MethodGet,
			path:               "/api/v2/salt_cache/cache-key/",
			allowedPermissions: `["@jobs"]`,
			deniedPermissions:  `[]`,
			allowedStatus:      http.StatusBadRequest,
		},
		{
			name:               "v2 Salt cache delete",
			method:             http.MethodDelete,
			path:               "/api/v2/salt_cache/cache-key/",
			allowedPermissions: `["test.ping"]`,
			deniedPermissions:  `["@jobs"]`,
			allowedStatus:      http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			allowed := serveAuthorizedRoute(t, upstream.URL, test, test.allowedPermissions, test.expectSaltKeys)
			require.Equal(t, test.allowedStatus, allowed.Code, allowed.Body.String())

			denied := serveAuthorizedRoute(t, upstream.URL, test, test.deniedPermissions, false)
			require.Equal(t, http.StatusForbidden, denied.Code, denied.Body.String())
		})
	}
}

func serveAuthorizedRoute(
	t *testing.T,
	upstreamURL string,
	test authorizationRouteCase,
	permissions string,
	expectSaltKeys bool,
) *httptest.ResponseRecorder {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	previousDB := db.DB
	previousOptions := options
	previousSaltOptions := saltOptions
	previousSaltTables := saltDBTables
	db.DB = database
	options.Secret = routeAuthorizationSecret
	saltOptions.URL = upstreamURL
	saltDBTables = config.SaltDBTables{SaltCache: "salt_cache", SaltKeys: "salt_keys"}
	t.Cleanup(func() {
		db.DB = previousDB
		options = previousOptions
		saltOptions = previousSaltOptions
		saltDBTables = previousSaltTables
		require.NoError(t, mock.ExpectationsWereMet())
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	expectActiveRouteUser(mock)
	expectRouteSaltPermissions(mock, permissions)
	if expectSaltKeys {
		expectMissingSaltKeysTable(mock)
	}

	engine := gin.New()
	engine.Use(sessions.Sessions("agarthaAuthSession", cookie.NewStore([]byte(routeAuthorizationSecret))))
	addServerRoutes(engine)

	request := httptest.NewRequest(test.method, test.path, bytes.NewBufferString(test.body))
	request.Header.Set("Authorization", "Bearer "+signedRouteAuthorizationToken(t))
	request.Header.Set("Content-Type", "application/json")
	if test.useSaltToken {
		request.Header.Set("X-Auth-Token", routeAuthorizationSaltToken)
	}
	response := &closeNotifyRecorder{ResponseRecorder: httptest.NewRecorder()}
	engine.ServeHTTP(response, request)
	return response.ResponseRecorder
}

func expectActiveRouteUser(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_user" WHERE id = $1 AND username = $2 AND is_active = $3 ORDER BY "auth_user"."id" LIMIT $4`)).
		WithArgs(uint(7), routeAuthorizationUsername, true, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "is_active", "is_staff", "is_superuser"}).
			AddRow(7, routeAuthorizationUsername, true, false, false))
}

func expectRouteSaltPermissions(mock sqlmock.Sqlmock, permissions string) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "salt_permissions" FROM "user_settings" WHERE user_id = $1 ORDER BY "user_settings"."user_id" LIMIT $2`)).
		WithArgs(uint(7), 1).
		WillReturnRows(sqlmock.NewRows([]string{"salt_permissions"}).AddRow(permissions))
}

func expectMissingSaltKeysTable(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2`)).
		WithArgs("salt_keys", "BASE TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

func signedRouteAuthorizationToken(t *testing.T) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": routeAuthorizationUsername,
		"user_id":  7,
		"exp":      4102444800,
	})
	signed, err := token.SignedString([]byte(routeAuthorizationSecret))
	require.NoError(t, err)
	return signed
}
