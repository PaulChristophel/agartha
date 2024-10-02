package validate

import (
	get "github.com/PaulChristophel/agartha/server/api/v1/validate/get"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/validate")

	grp.GET("/", get.Validate)
}
