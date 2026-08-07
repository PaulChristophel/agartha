package auth

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/")

	grp.POST("/token", RetrieveToken)
	grp.POST("/logout", Logout)
	grp.GET("/method", GetMethod)
}

func AddSessionRoutes(rg *gin.RouterGroup) {
	rg.GET("/session", GetSession)
}
