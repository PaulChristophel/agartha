package conformity

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/conformity/get"
	post "github.com/PaulChristophel/agartha/server/api/v1/conformity/post"
	"github.com/gin-gonic/gin"
)

var (
	isRefreshing bool = false
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/conformity")

	post.SetOptions(&isRefreshing)
	get.SetOptions(&isRefreshing)

	grp.GET("/", get.GetConformities)
	grp.POST("/refresh", post.Refresh)
	grp.GET("/refresh", get.Refresh)
	grp.GET("/:id", get.GetConformity)
}
