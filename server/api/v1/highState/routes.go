package highState

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/highState/get"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/high_state")

	grp.GET("", get.GetHighStates)
	grp.GET("/:id", get.GetHighState)
}
