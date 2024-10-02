package jid

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/jid/get"
	"github.com/PaulChristophel/agartha/server/config"

	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/jid")

	grp.GET("", get.GetJIDs)
	grp.GET("/:jid", get.GetJID)
	// grp.GET("/:jid/:alter_time", get.GetJIDTime)
}

func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
}
