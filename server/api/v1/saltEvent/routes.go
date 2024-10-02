package saltEvent

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/saltEvent/get"
	"github.com/PaulChristophel/agartha/server/config"

	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_event")

	grp.GET("/", get.GetSaltEvents)
	grp.GET("/:id", get.GetSaltEvent)
}

func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
}
