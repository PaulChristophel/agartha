package middleware

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestAuthRequiredSetsValidatedClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := []byte("test-secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": "alice",
		"user_id":  7,
		"exp":      4102444800,
	})
	signedToken, err := token.SignedString(secret)
	require.NoError(t, err)

	router := gin.New()
	router.GET("/protected", AuthRequired(secret), func(c *gin.Context) {
		username, _ := c.Get("username")
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"username": username, "user_id": userID})
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer "+signedToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"username":"alice","user_id":7}`, response.Body.String())
}

func TestActiveUserRequired(t *testing.T) {
	tests := []struct {
		name       string
		rows       *sqlmock.Rows
		wantStatus int
	}{
		{
			name: "active user",
			rows: sqlmock.NewRows([]string{"id", "username", "is_active", "is_staff", "is_superuser"}).
				AddRow(7, "alice", true, false, false),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "inactive or missing user",
			rows:       sqlmock.NewRows([]string{"id", "username", "is_active", "is_staff", "is_superuser"}),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "auth_user" WHERE id = $1 AND username = $2 AND is_active = $3 ORDER BY "auth_user"."id" LIMIT $4`)).
				WithArgs(uint(7), "alice", true, 1).
				WillReturnRows(tt.rows)

			router := gin.New()
			router.GET("/protected", func(c *gin.Context) {
				c.Set("user_id", uint(7))
				c.Set("username", "alice")
				c.Next()
			}, ActiveUserRequired(database), noContent)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/protected", nil))
			require.Equal(t, tt.wantStatus, response.Code)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUniqueAuthRequiredChecksCurrentUser(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		user       model.AuthUser
		wantStatus int
	}{
		{name: "owner", path: "/users/7", user: model.AuthUser{ID: 7}, wantStatus: http.StatusNoContent},
		{name: "different user", path: "/users/8", user: model.AuthUser{ID: 7}, wantStatus: http.StatusForbidden},
		{name: "superuser", path: "/users/8", user: model.AuthUser{ID: 7, IsSuperuser: true}, wantStatus: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/users/:id", func(c *gin.Context) {
				c.Set(authUserContextKey, tt.user)
				c.Next()
			}, UniqueAuthRequired(), noContent)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))
			require.Equal(t, tt.wantStatus, response.Code)
		})
	}
}

func noContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}
