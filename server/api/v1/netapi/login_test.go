package netapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDecodeTokenAndCreateCredentialsUsesServerSessionToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)
	router := gin.New()
	router.Use(sessions.Sessions(
		"agarthaAuthSession",
		cookie.NewStore([]byte("01234567890123456789012345678901")),
	))
	router.GET("/seed", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("auth_token", "Bearer server-side-token")
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	router.POST(
		"/login",
		func(c *gin.Context) {
			c.Set("username", "alice")
			c.Next()
		},
		DecodeTokenAndCreateCredentials(),
		func(c *gin.Context) {
			var credentials Credentials
			require.NoError(t, c.ShouldBindJSON(&credentials))
			c.JSON(http.StatusOK, credentials)
		},
	)

	seedResponse := httptest.NewRecorder()
	router.ServeHTTP(seedResponse, httptest.NewRequest(http.MethodGet, "/seed", nil))
	require.NotEmpty(t, seedResponse.Result().Cookies())

	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{}`))
	request.AddCookie(seedResponse.Result().Cookies()[0])
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"username":"alice","password":"Bearer server-side-token","eauth":"agartha"}`, response.Body.String())
}
