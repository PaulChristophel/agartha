package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const saltPermissionsQuery = `SELECT "salt_permissions" FROM "user_settings" WHERE user_id = $1 ORDER BY "user_settings"."user_id" LIMIT $2`

func TestSaltPermissionRequiredFailureModes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing authentication context", func(t *testing.T) {
		status := authorizationStatus(t, nil, SaltPermissionRequired(nil, ReadSaltData), http.MethodGet)
		require.Equal(t, http.StatusUnauthorized, status)
	})

	tests := []struct {
		name       string
		expect     func(sqlmock.Sqlmock)
		wantStatus int
	}{
		{
			name: "no user settings row",
			expect: func(mock sqlmock.Sqlmock) {
				expectSaltPermissions(mock).
					WillReturnRows(sqlmock.NewRows([]string{"salt_permissions"}))
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "empty Salt permissions",
			expect: func(mock sqlmock.Sqlmock) {
				expectSaltPermissions(mock).
					WillReturnRows(sqlmock.NewRows([]string{"salt_permissions"}).AddRow(`[]`))
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "malformed cached permission JSON",
			expect: func(mock sqlmock.Sqlmock) {
				expectSaltPermissions(mock).
					WillReturnRows(sqlmock.NewRows([]string{"salt_permissions"}).AddRow(`{"@wheel":`))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "database failure",
			expect: func(mock sqlmock.Sqlmock) {
				expectSaltPermissions(mock).WillReturnError(errors.New("database unavailable"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			tt.expect(mock)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				SaltPermissionRequired(database, ReadSaltData),
				http.MethodGet,
			)

			require.Equal(t, tt.wantStatus, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestReadOnlyJobsPermissionsAllowReadsButDenyCommands(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		method     string
		wantStatus int
	}{
		{name: "read", method: http.MethodGet, wantStatus: http.StatusNoContent},
		{name: "command", method: http.MethodPost, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			expectPermissionJSON(mock, `["@jobs"]`)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				SaltPermissionForMethodRequired(database),
				tt.method,
			)

			require.Equal(t, tt.wantStatus, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestExecutableSaltPermissionsAllowCommands(t *testing.T) {
	gin.SetMode(gin.TestMode)
	database, mock := authorizationTestDatabase(t)
	expectPermissionJSON(mock, `["test.*"]`)

	status := authorizationStatus(
		t,
		&model.AuthUser{ID: 7, Username: "alice"},
		SaltPermissionRequired(database, ExecuteSaltCommand),
		http.MethodPost,
	)

	require.Equal(t, http.StatusNoContent, status)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestScopedWheelPermissionsOnlyAllowMatchingFunction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		function   string
		wantStatus int
	}{
		{name: "matching function", function: "key.accept", wantStatus: http.StatusNoContent},
		{name: "different function", function: "key.delete", wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			expectPermissionJSON(mock, `[{"@wheel":["key.accept"]}]`)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				SaltWheelPermissionRequired(database, tt.function),
				http.MethodPost,
			)

			require.Equal(t, tt.wantStatus, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBareWheelPermissionAllowsAllWheelAndRawKeyActions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	wheelFunctions := []string{"key.list_all", "key.accept", "key.reject", "key.delete"}
	for _, function := range wheelFunctions {
		t.Run(function, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			expectPermissionJSON(mock, `["@wheel"]`)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				SaltWheelPermissionRequired(database, function),
				http.MethodPost,
			)

			require.Equal(t, http.StatusNoContent, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		t.Run("raw key "+method, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			expectPermissionJSON(mock, `["@wheel"]`)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				SaltWheelAdministrationRequired(database),
				method,
			)

			require.Equal(t, http.StatusNoContent, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestScopedWheelPermissionCannotAccessRawKeyMaterial(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			expectPermissionJSON(mock, `[{"@wheel":["key.*"]}]`)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				SaltWheelAdministrationRequired(database),
				method,
			)

			require.Equal(t, http.StatusForbidden, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestStaffAndSuperusersReceiveOperationalOverrides(t *testing.T) {
	gin.SetMode(gin.TestMode)

	users := []struct {
		name string
		user model.AuthUser
	}{
		{name: "staff", user: model.AuthUser{ID: 7, Username: "staff", IsStaff: true}},
		{name: "superuser", user: model.AuthUser{ID: 8, Username: "root", IsSuperuser: true}},
	}

	for _, tt := range users {
		t.Run(tt.name, func(t *testing.T) {
			tests := []struct {
				name       string
				middleware gin.HandlerFunc
			}{
				{name: "command execution", middleware: SaltPermissionRequired(nil, ExecuteSaltCommand)},
				{name: "wheel function", middleware: SaltWheelPermissionRequired(nil, "key.delete")},
				{name: "raw key administration", middleware: SaltWheelAdministrationRequired(nil)},
			}

			for _, capability := range tests {
				t.Run(capability.name, func(t *testing.T) {
					status := authorizationStatus(t, &tt.user, capability.middleware, http.MethodPost)
					require.Equal(t, http.StatusNoContent, status)
				})
			}
		})
	}
}

func TestOrdinaryAuthenticatedUsersCannotUseAdministrativeCapabilities(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		middleware func(*gorm.DB) gin.HandlerFunc
	}{
		{
			name: "wheel function",
			middleware: func(database *gorm.DB) gin.HandlerFunc {
				return SaltWheelPermissionRequired(database, "key.delete")
			},
		},
		{
			name:       "raw key administration",
			middleware: SaltWheelAdministrationRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			database, mock := authorizationTestDatabase(t)
			expectPermissionJSON(mock, `["test.*"]`)

			status := authorizationStatus(
				t,
				&model.AuthUser{ID: 7, Username: "alice"},
				tt.middleware(database),
				http.MethodPost,
			)

			require.Equal(t, http.StatusForbidden, status)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func authorizationTestDatabase(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})

	database, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)
	return database, mock
}

func expectSaltPermissions(mock sqlmock.Sqlmock) *sqlmock.ExpectedQuery {
	return mock.ExpectQuery(regexp.QuoteMeta(saltPermissionsQuery)).WithArgs(uint(7), 1)
}

func expectPermissionJSON(mock sqlmock.Sqlmock, permissions string) {
	expectSaltPermissions(mock).
		WillReturnRows(sqlmock.NewRows([]string{"salt_permissions"}).AddRow(permissions))
}

func authorizationStatus(
	t *testing.T,
	user *model.AuthUser,
	authorization gin.HandlerFunc,
	method string,
) int {
	t.Helper()

	router := gin.New()
	router.Handle(method, "/protected", func(c *gin.Context) {
		if user != nil {
			c.Set(authUserContextKey, *user)
		}
		c.Next()
	}, authorization, noContent)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(method, "/protected", nil))
	return response.Code
}
