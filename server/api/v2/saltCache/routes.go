package saltCache

import (

	// "strings"
	delete "github.com/PaulChristophel/agartha/server/api/v2/saltCache/delete"
	get "github.com/PaulChristophel/agartha/server/api/v2/saltCache/get"
	"github.com/PaulChristophel/agartha/server/config"

	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_cache")

	grp.DELETE("/:key/*bank", delete.DeleteSaltCacheBankKey)
	grp.GET("/:key/*bank", get.GetSaltCacheBankKey)
}

func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
}
