package netapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/PaulChristophel/agartha/server/api/validate"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/PaulChristophel/agartha/server/middleware"
	"github.com/gin-contrib/sessions"
	"gorm.io/gorm"
)

func Handler(r *gin.RouterGroup, target string, database *gorm.DB) {

	headerCheck := func(c *gin.Context) {
		token := c.GetHeader("X-Auth-Token")
		if token == "" {
			token, _ = sessions.Default(c).Get("salt_token").(string)
		}
		_, err := validate.Token(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid X-Auth-Token"})
			c.Abort()
			return
		}
		c.Request.Header.Set("X-Auth-Token", token)
		c.Next()
	}

	// Proxy handler for exact match
	r.Any("/netapi", middleware.SaltPermissionForMethodRequired(database), headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath(), nil)
	})

	// Proxy handler for exact match
	r.Any("/netapi/", middleware.SaltPermissionForMethodRequired(database), headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath(), nil)
	})

	// Proxy handler for login
	r.Any("/netapi/login", DecodeTokenAndCreateCredentials(), func(c *gin.Context) {
		proxy(c, target, r.BasePath(), cacheSaltPermissions(c, database))
	})

	// Proxy handler for logout
	r.Any("/netapi/logout", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath(), nil)
	})

	// Proxy handler for hook
	r.Any("/netapi/hook", middleware.SaltPermissionRequired(database, middleware.ExecuteSaltCommand), headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath(), nil)
	})

	// Proxy handler for hook
	r.Any("/netapi/hook/*path", middleware.SaltPermissionRequired(database, middleware.ExecuteSaltCommand), headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath(), nil)
	})

	// Proxy handler for stats
	r.Any("/netapi/stats", middleware.SaltPermissionRequired(database, middleware.ReadSaltData), headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath(), nil)
	})

	// // Proxy handler for paths
	// r.Any("/netapi/*path", headerCheck, func(c *gin.Context) {
	// 	proxy(c, target, r.BasePath())
	// })
}

func proxy(c *gin.Context, target, repl string, modifyResponse func(*http.Response) error) {
	remote, err := url.Parse(target)
	if err != nil {
		logger.GetLogger().Sugar().Fatalf("Could not parse target URL: %v", err)
	}

	// Do NOT use NewSingleHostReverseProxy (it sets Director, which triggers SA1019 and conflicts with Rewrite in Go 1.26).
	proxy := &httputil.ReverseProxy{}
	proxy.ModifyResponse = modifyResponse

	// Custom transport with timeout
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	proxy.Rewrite = func(pr *httputil.ProxyRequest) {
		// Standard reverse-proxy rewrite (replaces Director)
		pr.SetURL(remote)

		// capture incoming path  query from the inbound request
		p := pr.In.URL.Path
		if after, ok := strings.CutPrefix(p, repl+"/netapi"); ok {
			p = after
			if p == "" {
				p = "/"
			}
		}
		q := pr.In.URL.RawQuery

		pr.Out.Header.Set("User-Agent", "Go-http-client/1.1")

		// join upstream base path (e.g. /pepper/) with rewritten path
		base := remote.Path
		if base == "" {
			base = "/"
		}
		pr.Out.URL.Path, _ = url.JoinPath(base, p)
		pr.Out.URL.RawQuery = q

		// Clear Authorization header for endpoints other than /login
		if !strings.Contains(pr.Out.URL.Path, "/login") {
			pr.Out.Header.Del("Authorization")
		}

		if gin.Mode() == gin.DebugMode && pr.Out.Body != nil {
			body, err := io.ReadAll(pr.Out.Body)
			if err == nil {
				logger.GetLogger().Sugar().Debugf("Forwarded Request Body: %s", string(body))
				pr.Out.Body = io.NopCloser(bytes.NewBuffer(body))
			} else {
				logger.GetLogger().Sugar().Debugf("Error reading request body: %s", err)
			}
		}
	}

	// Forward the request to the proxy
	proxy.ServeHTTP(c.Writer, c.Request)
}

func cacheSaltPermissions(c *gin.Context, database *gorm.DB) func(*http.Response) error {
	usernameValue, usernameOK := c.Get("username")
	userIDValue, userIDOK := c.Get("user_id")
	username, usernameTypeOK := usernameValue.(string)
	userID, userIDTypeOK := userIDValue.(uint)

	return func(response *http.Response) error {
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return nil
		}
		if !usernameOK || !userIDOK || !usernameTypeOK || !userIDTypeOK {
			return fmt.Errorf("validated user context is missing")
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("read Salt login response: %w", err)
		}
		response.Body = io.NopCloser(bytes.NewReader(body))
		response.ContentLength = int64(len(body))

		var login struct {
			Return []struct {
				User  string          `json:"user"`
				Token string          `json:"token"`
				Perms json.RawMessage `json:"perms"`
			} `json:"return"`
		}
		if err := json.Unmarshal(body, &login); err != nil || len(login.Return) != 1 {
			return fmt.Errorf("invalid Salt login response")
		}
		if login.Return[0].User != username {
			return fmt.Errorf("salt login identity does not match authenticated user")
		}
		if _, err := validate.Token(login.Return[0].Token); err != nil {
			return fmt.Errorf("invalid Salt token in login response")
		}
		session := sessions.Default(c)
		session.Set("salt_token", login.Return[0].Token)
		if err := session.Save(); err != nil {
			return fmt.Errorf("save Salt token in session: %w", err)
		}
		permissions := login.Return[0].Perms
		if len(permissions) == 0 {
			permissions = json.RawMessage(`[]`)
		}
		var decoded any
		if err := json.Unmarshal(permissions, &decoded); err != nil {
			return fmt.Errorf("invalid Salt permissions in login response")
		}

		result := database.Model(&struct {
			UserID uint `gorm:"column:user_id"`
		}{}).Table("user_settings").Where("user_id = ?", userID).Update("salt_permissions", string(permissions))
		if result.Error != nil {
			return fmt.Errorf("cache Salt permissions: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("cache Salt permissions: user settings row not found")
		}
		return nil
	}
}
