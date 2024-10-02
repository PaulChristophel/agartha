package authUser

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/secure/authUser/get"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/auth_user")

	// grp.GET("/", GetauthUsers)
	grp.GET("/:id", get.GetAuthUser)
}
