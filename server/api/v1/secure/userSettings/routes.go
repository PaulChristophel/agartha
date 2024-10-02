package userSettings

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/secure/userSettings/get"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/user_settings")

	// grp.GET("/", GetUserSettings)
	grp.GET("/:id", get.GetUserSettings)
}
