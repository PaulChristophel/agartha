package saltReturn

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/saltReturn/get"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_return")

	grp.GET("", get.GetSaltReturns)
	grp.GET("/fun", get.ListSaltReturnFuns)
	grp.GET("/:jid", get.GetSaltReturnJID)
	grp.GET("/:jid/:id", get.GetSaltReturnID)
}

func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
}
