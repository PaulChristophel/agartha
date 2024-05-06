package routes

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/PaulChristophel/agartha/server/api/v1/conformity"
	"github.com/PaulChristophel/agartha/server/api/v1/highState"
	"github.com/PaulChristophel/agartha/server/api/v1/jid"
	"github.com/PaulChristophel/agartha/server/api/v1/saltCache"
	"github.com/PaulChristophel/agartha/server/api/v1/saltEvent"
	"github.com/PaulChristophel/agartha/server/api/v1/saltReturn"
	"github.com/PaulChristophel/agartha/server/api/v1/secure/authUser"
	"github.com/PaulChristophel/agartha/server/api/v1/secure/userSettings"
	"github.com/PaulChristophel/agartha/server/auth"
	"github.com/PaulChristophel/agartha/server/db"
	docsV1 "github.com/PaulChristophel/agartha/server/docs/v1"
	"github.com/PaulChristophel/agartha/server/middleware"

	gormsessions "github.com/gin-contrib/sessions/gorm"

	swaggerfiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

var router *gin.Engine = gin.New()

func Router(f embed.FS) {
	docsV1.SwaggerInfo.BasePath = "/"

	if os.Getenv("GIN_MODE") == "debug" {
		gin.SetMode(gin.DebugMode)
		router.Use(gin.Logger(), gin.Recovery())
		// logger, _ := zap.NewProduction()
		// // Add a ginzap middleware, which:
		// //   - Logs all requests, like a combined access and error log.
		// //   - Logs to stdout.
		// //   - RFC3339 with UTC time format.
		// router.Use(ginzap.Ginzap(logger, time.RFC3339, true))

		// // Logs all panic to error log
		// //   - stack means whether output the stack info.
		// router.Use(ginzap.RecoveryWithZap(logger, true))
	} else {
		gin.SetMode(gin.ReleaseMode)
		router.Use(gin.Logger(), gin.Recovery())
		// logger, _ := zap.NewProduction()
		// // Add a ginzap middleware, which:
		// //   - Logs all requests, like a combined access and error log.
		// //   - Logs to stdout.
		// //   - RFC3339 with UTC time format.
		// router.Use(ginzap.Ginzap(logger, time.RFC3339, true))

		// // Logs all panic to error log
		// //   - stack means whether output the stack info.
		// router.Use(ginzap.RecoveryWithZap(logger, true))
	}

	// This is the server backend API
	configureSessionStore(router)
	addServerRoutes(router)
	addStaticRoutes(router, f)
	router.Run(":" + viper.GetString("port"))
}

func configureSessionStore(router *gin.Engine) {
	store := gormsessions.NewStore(db.DB, true, []byte(viper.GetString("secret")))
	router.Use(sessions.Sessions("agarthaAuthSession", store))
}

func addServerRoutes(router *gin.Engine) {
	rootRoute := router.Group("/")
	AddPingRoutes(rootRoute)

	authRoute := router.Group("/auth")
	auth.AddRoutes(authRoute)

	v1 := router.Group("/api/v1", middleware.AuthRequired())
	conformity.AddRoutes(v1)
	jid.AddRoutes(v1)
	saltCache.AddRoutes(v1)
	saltEvent.AddRoutes(v1)
	highState.AddRoutes(v1)
	saltReturn.AddRoutes(v1)

	secure := router.Group("/api/v1/secure", middleware.AuthRequired(), middleware.UniqueAuthRequired(db.DB))
	userSettings.AddRoutes(secure)
	authUser.AddRoutes(secure)

}

func addStaticRoutes(router *gin.Engine, f embed.FS) {

	// Configure static routes for documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	router.GET("/api", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/docs/index.html")
	})
	router.GET("/api/v1", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/docs/index.html")
	})

	// Configure static routes for the vue frontend
	assetsFS, err := fs.Sub(f, "web/dist/assets")
	if err != nil {
		log.Print(err)
	}
	staticFS, err := fs.Sub(f, "web/dist/static")
	if err != nil {
		log.Print(err)
	}
	router.StaticFS("/assets", http.FS(assetsFS))
	router.StaticFS("/static", http.FS(staticFS))
	router.GET("/favicon.ico", func(c *gin.Context) {
		file, _ := f.ReadFile("web/dist/favicon.ico")
		c.Data(
			http.StatusOK,
			"image/x-icon",
			file,
		)
	})
	// router.GET("/logo.png", func(c *gin.Context) {
	// 	file, _ := f.ReadFile("web/dist/logo.png")
	// 	c.Data(
	// 		http.StatusOK,
	// 		"image/png",
	// 		file,
	// 	)
	// })
	// router.GET("/_app.config.js", func(c *gin.Context) {
	// 	file, _ := f.ReadFile("web/dist/_app.config.js")
	// 	c.Data(
	// 		http.StatusOK,
	// 		"text/javascript",
	// 		file,
	// 	)
	// })
	// Everything that doesn't match can be routed to the frontend.
	router.NoRoute(func(c *gin.Context) {
		file, _ := f.ReadFile("web/dist/index.html")
		c.Data(
			http.StatusOK,
			"text/html; charset=utf-8",
			file,
		)
	})
}
