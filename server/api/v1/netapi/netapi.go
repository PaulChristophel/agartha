package netapi

import (
	"bytes"
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
)

func Handler(r *gin.RouterGroup, target string) {

	headerCheck := func(c *gin.Context) {
		_, err := validate.Token(c.GetHeader("X-Auth-Token"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid X-Auth-Token"})
			c.Abort()
			return
		}
		c.Next()
	}

	// Proxy handler for exact match
	r.Any("/netapi", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// Proxy handler for exact match
	r.Any("/netapi/", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// Proxy handler for login
	r.Any("/netapi/login", DecodeTokenAndCreateCredentials(), func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// Proxy handler for logout
	r.Any("/netapi/logout", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// Proxy handler for hook
	r.Any("/netapi/hook", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// Proxy handler for hook
	r.Any("/netapi/hook/*path", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// Proxy handler for stats
	r.Any("/netapi/stats", headerCheck, func(c *gin.Context) {
		proxy(c, target, r.BasePath())
	})

	// // Proxy handler for paths
	// r.Any("/netapi/*path", headerCheck, func(c *gin.Context) {
	// 	proxy(c, target, r.BasePath())
	// })
}

func proxy(c *gin.Context, target, repl string) {
	remote, err := url.Parse(target)
	if err != nil {
		logger.GetLogger().Sugar().Fatalf("Could not parse target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)

	// Custom transport with timeout
	proxy.Transport = &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	// Custom director to modify the request
	proxy.Director = func(req *http.Request) {
		req.Header.Set("User-Agent", "Go-http-client/1.1")
		if strings.HasPrefix(req.URL.Path, repl+"/netapi") {
			req.URL.Path = strings.TrimPrefix(req.URL.Path, repl+"/netapi")
		} else if req.URL.Path == repl {
			req.URL.Path = "/"
		}
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host

		// Clear the Authorization header for endpoints other than /login
		if !strings.Contains(req.URL.Path, "/login") {
			req.Header.Del("Authorization")
		}

		if gin.Mode() == gin.DebugMode { // Log the request being forwarded
			if req.Body != nil {
				body, err := io.ReadAll(req.Body)
				if err == nil {
					logger.GetLogger().Sugar().Debugf("Forwarded Request Body: %s", string(body))
					req.Body = io.NopCloser(bytes.NewBuffer(body))
				} else {
					logger.GetLogger().Sugar().Debugf("Error reading request body: %s", err)
				}
			} else {
				logger.GetLogger().Debug("Request Body is nil")
			}
		}
	}

	// Forward the request to the proxy
	proxy.ServeHTTP(c.Writer, c.Request)
}
