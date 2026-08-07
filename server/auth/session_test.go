package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestLogoutExpiresSecureHttpOnlySessionCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetOptions([]byte("secret"), []string{"local"}, config.LDAPOptions{}, config.CASOptions{}, true)

	store := cookie.NewStore([]byte("01234567890123456789012345678901"))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	router := gin.New()
	router.Use(sessions.Sessions("agarthaAuthSession", store))
	router.GET("/seed", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("username", "alice")
		require.NoError(t, session.Save())
		c.Status(http.StatusNoContent)
	})
	router.POST("/auth/logout", Logout)

	seedResponse := httptest.NewRecorder()
	router.ServeHTTP(seedResponse, httptest.NewRequest(http.MethodGet, "/seed", nil))
	require.NotEmpty(t, seedResponse.Result().Cookies())

	request := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	request.AddCookie(seedResponse.Result().Cookies()[0])
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusNoContent, response.Code)
	cookies := response.Result().Cookies()
	require.Len(t, cookies, 2)

	expiredByPath := make(map[string]*http.Cookie, len(cookies))
	for _, expired := range cookies {
		expiredByPath[expired.Path] = expired
		require.Less(t, expired.MaxAge, 0)
		require.True(t, expired.HttpOnly)
		require.True(t, expired.Secure)
		require.Equal(t, http.SameSiteLaxMode, expired.SameSite)
	}
	require.Contains(t, expiredByPath, "/")
	require.Contains(t, expiredByPath, legacyAuthCookiePath)
}

func TestExpireLegacyAuthCookieUsesLegacyAuthPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetOptions([]byte("secret"), []string{"local"}, config.LDAPOptions{}, config.CASOptions{}, true)

	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	expireLegacyAuthCookie(context)

	cookies := response.Result().Cookies()
	require.Len(t, cookies, 1)
	require.Equal(t, authSessionCookieName, cookies[0].Name)
	require.Equal(t, legacyAuthCookiePath, cookies[0].Path)
	require.Less(t, cookies[0].MaxAge, 0)
	require.True(t, cookies[0].HttpOnly)
	require.True(t, cookies[0].Secure)
}
