package routes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PaulChristophel/agartha/server/api/v1/conformity"
	"github.com/PaulChristophel/agartha/server/api/v1/highState"
	"github.com/PaulChristophel/agartha/server/api/v1/jid"
	"github.com/PaulChristophel/agartha/server/api/v1/netapi"
	"github.com/PaulChristophel/agartha/server/api/v1/saltCache"
	"github.com/PaulChristophel/agartha/server/api/v1/saltEvent"
	"github.com/PaulChristophel/agartha/server/api/v1/saltKeys"
	"github.com/PaulChristophel/agartha/server/api/v1/saltMinion"
	"github.com/PaulChristophel/agartha/server/api/v1/saltReturn"
	"github.com/PaulChristophel/agartha/server/api/v1/secure/authUser"
	"github.com/PaulChristophel/agartha/server/api/v1/secure/userSettings"
	"github.com/PaulChristophel/agartha/server/api/v1/validate"

	v2SaltCache "github.com/PaulChristophel/agartha/server/api/v2/saltCache"

	// saltCachev2 "github.com/PaulChristophel/agartha/server/api/v2/saltCache"
	"github.com/PaulChristophel/agartha/server/auth"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	docsV1 "github.com/PaulChristophel/agartha/server/docs/v1"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/PaulChristophel/agartha/server/middleware"
	gormsessions "github.com/gin-contrib/sessions/gorm"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	router       *gin.Engine
	saltDBTables config.SaltDBTables
	options      config.HTTPOptions
	ldapOptions  config.LDAPOptions
	casOptions   config.CASOptions
	saltOptions  config.SaltOptions
	authMethods  []string
	log          *zap.Logger
)

func Router(frontend embed.FS, agarthaOptions config.Config) error {
	options = agarthaOptions.HTTP
	ldapOptions = agarthaOptions.LDAP
	casOptions = agarthaOptions.CAS
	saltOptions = agarthaOptions.Salt
	saltDBTables = agarthaOptions.DB.Tables
	var err error
	authMethods, err = agarthaOptions.EffectiveAuthMethods()
	if err != nil {
		return err
	}

	docsV1.SwaggerInfo.BasePath = "/"

	// If this is a pre-release, always run in debug mode
	isPre := Version == "" || strings.Contains(Version, "-")
	mode := gin.ReleaseMode
	if os.Getenv("GIN_MODE") == "debug" || isPre {
		mode = gin.DebugMode
	}

	gin.SetMode(mode)
	router = gin.New()
	log, err = logger.InitLogger(mode)
	if err != nil {
		return err
	}
	defer func() {
		if err := log.Sync(); err != nil {
			log.Error("Failed to sync logger", zap.Error(err))
		}
	}()

	trustedProxies := options.TrustedProxies
	if len(trustedProxies) == 0 {
		trustedProxies = nil
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		return fmt.Errorf("configure trusted proxies: %w", err)
	}

	router.Use(securityHeaders(), gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(param gin.LogFormatterParams) string {
			log.Info("request",
				zap.String("client_ip", param.ClientIP),
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.Int("status", param.StatusCode),
				zap.String("latency", param.Latency.String()),
				zap.String("user-agent", param.Request.UserAgent()),
			)
			return ""
		},
	}), gin.Recovery())

	router.UseRawPath = true
	router.UnescapePathValues = false

	// This is the server backend API
	store := gormsessions.NewStore(db.DB, true, []byte(options.Secret))
	router.Use(sessions.Sessions("agarthaAuthSession", store))
	addServerRoutes(router)
	addStaticRoutes(router, frontend)

	addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	srv := newHTTPServer(addr, router, options)

	// TLS support (server cert/key). Put full chain (server + intermediates) in the cert PEM file.
	if options.TLSCertFile != "" && options.TLSKeyFile != "" {
		tlsCfg := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		// Optional mTLS: verify client certs against a provided CA bundle (PEM).
		if options.TLSClientCAFile != "" {
			caPEM, readErr := os.ReadFile(options.TLSClientCAFile)
			if readErr != nil {
				log.Error("Failed to read TLS client CA file", zap.Error(readErr))
				return readErr
			}
			pool := x509.NewCertPool()
			if !pool.AppendCertsFromPEM(caPEM) {
				return fmt.Errorf("failed to parse TLS client CA PEM: %s", options.TLSClientCAFile)
			}
			tlsCfg.ClientCAs = pool
			tlsCfg.ClientAuth = tls.VerifyClientCertIfGiven
		}

		srv.TLSConfig = tlsCfg
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return serveHTTP(ctx, srv, options)
}

func newHTTPServer(addr string, handler http.Handler, options config.HTTPOptions) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       options.ReadTimeout,
		ReadHeaderTimeout: options.ReadHeaderTimeout,
		WriteTimeout:      options.WriteTimeout,
		IdleTimeout:       options.IdleTimeout,
	}
}

func serveHTTP(ctx context.Context, srv *http.Server, options config.HTTPOptions) error {
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", srv.Addr, err)
	}
	defer func() { _ = listener.Close() }()
	return serveHTTPListener(ctx, srv, listener, options)
}

func serveHTTPListener(ctx context.Context, srv *http.Server, listener net.Listener, options config.HTTPOptions) error {
	serveErr := make(chan error, 1)
	go func() {
		if options.TLSCertFile != "" && options.TLSKeyFile != "" {
			serveErr <- srv.ServeTLS(listener, options.TLSCertFile, options.TLSKeyFile)
			return
		}
		serveErr <- srv.Serve(listener)
	}()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		log.Info("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("gracefully shut down HTTP server: %w", err)
	}
	if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	log.Info("HTTP server stopped")
	return nil
}

func addServerRoutes(router *gin.Engine) {
	rootRoute := router.Group("/")
	AddPingRoutes(rootRoute)
	AddVersionRoutes(rootRoute)

	authRoute := router.Group("/auth")
	auth.SetOptions([]byte(options.Secret), authMethods, ldapOptions, casOptions)
	auth.AddRoutes(authRoute)

	grpV1 := router.Group(
		"/api/v1",
		middleware.AuthRequired([]byte(options.Secret)),
		middleware.ActiveUserRequired(db.DB),
	)
	netapi.Handler(grpV1, saltOptions.URL)
	conformity.AddRoutes(grpV1)
	jid.SetOptions(saltDBTables)
	jid.AddRoutes(grpV1)
	saltCache.SetOptions(saltDBTables)
	saltCache.AddRoutes(grpV1)
	saltKeys.SetOptions(saltDBTables)
	saltKeys.AddRoutes(grpV1)
	saltMinion.AddRoutes(grpV1)
	saltEvent.SetOptions(saltDBTables)
	saltEvent.AddRoutes(grpV1)
	highState.AddRoutes(grpV1)
	saltReturn.SetOptions(saltDBTables)
	saltReturn.AddRoutes(grpV1)
	validate.AddRoutes(grpV1)

	grpV1secure := grpV1.Group("/secure", middleware.UniqueAuthRequired())
	authUser.AddRoutes(grpV1secure)
	userSettings.AddRoutes(grpV1secure)

	grpV2 := router.Group(
		"/api/v2",
		middleware.AuthRequired([]byte(options.Secret)),
		middleware.ActiveUserRequired(db.DB),
	)
	v2SaltCache.SetOptions(saltDBTables)
	v2SaltCache.AddRoutes(grpV2)

}

func addDocRoutes(router *gin.Engine, frontend embed.FS) {
	router.GET("/docs/index.html", func(c *gin.Context) {
		file, err := frontend.ReadFile("web/dist/swagger.html")
		if err != nil {
			log.Error("failed to read swagger.html", zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", file)
	})
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	router.GET("/docs/doc.json", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs/swagger-ui.css", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs/swagger-ui.css.map", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs/swagger-ui-bundle.js", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs/swagger-ui-bundle.js.map", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs/swagger-ui-standalone-preset.js", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs/swagger-ui-standalone-preset.js.map", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

func addStaticRoutes(router *gin.Engine, frontend embed.FS) {
	addDocRoutes(router, frontend)

	assetsFS, err := fs.Sub(frontend, "web/dist/assets")
	if err != nil {
		log.Error("failed to get assetsFS", zap.Error(err))
		return
	}
	router.StaticFS("/assets", http.FS(assetsFS))

	staticFS, err := fs.Sub(frontend, "web/dist/static")
	if err != nil {
		log.Error("failed to get staticFS", zap.Error(err))
		return
	}
	router.StaticFS("/static", http.FS(staticFS))

	publicFS, err := fs.Sub(frontend, "web/dist/public")
	if err != nil {
		log.Error("failed to get publicFS", zap.Error(err))
		return
	}
	router.StaticFS("/public", http.FS(publicFS))

	faviconFS, err := fs.Sub(frontend, "web/dist/favicon")
	if err != nil {
		log.Error("failed to get faviconFS", zap.Error(err))
		return
	}
	router.StaticFS("/favicon", http.FS(faviconFS))

	router.GET("/manifest.json", func(c *gin.Context) {
		file, err := frontend.ReadFile("web/dist/manifest.json")
		if err != nil {
			log.Error("failed to read manifest.json", zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "application/json", file)
	})

	router.NoRoute(func(c *gin.Context) {
		config := struct {
			SaltAPIEndpoint   string
			Version           string
			GetStartedURL     string
			ForgotPasswordURL string
			CASServiceURL     string
		}{
			SaltAPIEndpoint:   saltOptions.ExternalURL,
			Version:           Version,
			GetStartedURL:     options.GetStartedURL,
			ForgotPasswordURL: options.ForgotPasswordURL,
			CASServiceURL:     casOptions.ServiceURL,
		}

		tmpl, err := template.ParseFS(frontend, "web/dist/index.html")
		if err != nil {
			log.Error("failed to parse index.html", zap.Error(err))
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(c.Writer, config); err != nil {
			log.Error("failed to execute template", zap.Error(err))
			c.Status(http.StatusInternalServerError)
		}
	})
}
