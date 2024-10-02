package saltCache

import (

	// "strings"
	delete "github.com/PaulChristophel/agartha/server/api/v1/saltCache/delete"
	get "github.com/PaulChristophel/agartha/server/api/v1/saltCache/get"
	post "github.com/PaulChristophel/agartha/server/api/v1/saltCache/post"
	"github.com/PaulChristophel/agartha/server/config"

	"github.com/gin-gonic/gin"
)

var (
	isRefreshing bool = false
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_cache")

	post.SetRefreshing(&isRefreshing)
	get.SetRefreshing(&isRefreshing)

	grp.DELETE("/uuid/:uuid", delete.DeleteSaltCacheUUID)
	grp.DELETE("/:bank/:key", delete.DeleteSaltCacheBankKey)
	grp.GET("", get.GetSaltCache)
	grp.GET("/fun_keys", get.ListSaltCacheDataKeys)
	grp.GET("/fun_keys/refresh", get.RefreshKeys)
	grp.GET("/uuid/:uuid", get.GetSaltCacheUUID)
	grp.GET("/:bank/:key", get.GetSaltCacheBankKey)
	grp.POST("/", post.CreateSaltCache)
	grp.POST("/fun_keys/refresh", post.RefreshKeys)
}

func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
	post.SetOptions(saltTables)
}
