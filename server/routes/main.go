package routes

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/PaulChristophel/agartha/server/api/v1/conformity"
	"github.com/PaulChristophel/agartha/server/api/v1/highState"
	"github.com/PaulChristophel/agartha/server/api/v1/jid"
	"github.com/PaulChristophel/agartha/server/api/v1/netapi"
	"github.com/PaulChristophel/agartha/server/api/v1/saltCache"
	"github.com/PaulChristophel/agartha/server/api/v1/saltEvent"
	"github.com/PaulChristophel/agartha/server/api/v1/saltMinion"
	"github.com/PaulChristophel/agartha/server/api/v1/saltReturn"
	"github.com/PaulChristophel/agartha/server/api/v1/secure/authUser"
	"github.com/PaulChristophel/agartha/server/api/v1/secure/userSettings"
	"github.com/PaulChristophel/agartha/server/api/v1/validate"

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
	router       *gin.Engine = gin.New()
	saltDBTables config.SaltDBTables
	options      config.HTTPOptions
	ldapOptions  config.LDAPOptions
	casOptions   config.CASOptions
	saltOptions  config.SaltOptions
	log          *zap.Logger
)

func Router(frontend embed.FS, agarthaOptions config.Config) error {
	options = agarthaOptions.HTTP
	ldapOptions = agarthaOptions.LDAP
	casOptions = agarthaOptions.CAS
	saltOptions = agarthaOptions.Salt
	saltDBTables = agarthaOptions.DB.Tables

	docsV1.SwaggerInfo.BasePath = "/"

	// If this is a pre-release, always run in debug mode
	isPre := Version == "" || strings.Contains(Version, "-")
	mode := gin.ReleaseMode
	if os.Getenv("GIN_MODE") == "debug" || isPre {
		mode = gin.DebugMode
	}

	gin.SetMode(mode)
	log, err := logger.InitLogger(mode)
	if err != nil {
		return err
	}
	defer func() {
		if err := log.Sync(); err != nil {
			log.Error("Failed to sync logger", zap.Error(err))
		}
	}()

	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
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
	err = router.Run(fmt.Sprintf("%s:%d", options.Host, options.Port))
	if err != nil {
		log.Error("Failed to start server", zap.Error(err))
	}
	return err
}

func addServerRoutes(router *gin.Engine) {
	rootRoute := router.Group("/")
	AddPingRoutes(rootRoute)
	AddVersionRoutes(rootRoute)

	authRoute := router.Group("/auth")
	auth.SetOptions([]byte(options.Secret), ldapOptions, casOptions)
	auth.AddRoutes(authRoute)

	grpV1 := router.Group("/api/v1", middleware.AuthRequired([]byte(options.Secret)))
	netapi.Handler(grpV1, saltOptions.URL)
	conformity.AddRoutes(grpV1)
	jid.SetOptions(saltDBTables)
	jid.AddRoutes(grpV1)
	saltCache.SetOptions(saltDBTables)
	saltCache.AddRoutes(grpV1)
	saltMinion.AddRoutes(grpV1)
	saltEvent.SetOptions(saltDBTables)
	saltEvent.AddRoutes(grpV1)
	highState.AddRoutes(grpV1)
	saltReturn.SetOptions(saltDBTables)
	saltReturn.AddRoutes(grpV1)
	validate.AddRoutes(grpV1)

	grpV1secure := router.Group("/api/v1/secure", middleware.AuthRequired([]byte(options.Secret)), middleware.UniqueAuthRequired(db.DB))
	authUser.AddRoutes(grpV1secure)
	userSettings.AddRoutes(grpV1secure)

	// grpV2 := router.Group("/api/v2", middleware.AuthRequired([]byte(options.Secret)))

	// saltCachev2.SetOptions(saltDBTables)
	// saltCachev2.AddRoutes(grpV2)
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
