package saltMinion

import (
	"sync/atomic"

	// "strings"
	get "github.com/PaulChristophel/agartha/server/api/v1/saltMinion/get"
	post "github.com/PaulChristophel/agartha/server/api/v1/saltMinion/post"

	"github.com/gin-gonic/gin"
)

var (
	isRefreshing atomic.Bool
	refreshError atomic.Pointer[string]
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_minion")

	post.SetRefreshState(&isRefreshing, &refreshError)
	get.SetRefreshState(&isRefreshing, &refreshError)

	grp.GET("", get.GetSaltMinion)
	grp.GET("/uuid/:uuid", get.GetSaltMinionUUID)
	grp.GET("/grains_keys", get.ListSaltMinionGrainsKeys)
	grp.GET("/pillar_keys", get.ListSaltMinionPillarKeys)
	grp.GET("/keys/refresh", get.RefreshKeys)
	grp.GET("/:minion_id", get.GetSaltMinionID)
	grp.POST("/keys/refresh", post.RefreshKeys)
}
