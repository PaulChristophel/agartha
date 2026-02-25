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

	// Do NOT use NewSingleHostReverseProxy (it sets Director, which triggers SA1019 and conflicts with Rewrite in Go 1.26).
	proxy := &httputil.ReverseProxy{}

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
