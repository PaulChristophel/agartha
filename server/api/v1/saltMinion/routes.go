package saltMinion

import (

	// "strings"
	get "github.com/PaulChristophel/agartha/server/api/v1/saltMinion/get"
	post "github.com/PaulChristophel/agartha/server/api/v1/saltMinion/post"

	"github.com/gin-gonic/gin"
)

var (
	isRefreshing bool = false
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_minion")

	post.SetRefreshing(&isRefreshing)
	get.SetRefreshing(&isRefreshing)

	grp.GET("", get.GetSaltMinion)
	grp.GET("/uuid/:uuid", get.GetSaltMinionUUID)
	grp.GET("/:minion_id", get.GetSaltMinionID)
	grp.GET("/grains_keys", get.ListSaltMinionGrainsKeys)
	grp.GET("/pillar_keys", get.ListSaltMinionPillarKeys)
	grp.GET("/keys/refresh", get.RefreshKeys)
	grp.POST("/keys/refresh", post.RefreshKeys)
}
