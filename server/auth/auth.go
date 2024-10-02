package auth

import (
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/")

	grp.POST("/token", RetrieveToken)
	grp.GET("/method", GetMethod)
}
