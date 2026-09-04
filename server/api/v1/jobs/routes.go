package jobs

import (
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AddRoutes registers job details with the caller's Salt authorization middleware.
func AddRoutes(rg *gin.RouterGroup, database *gorm.DB, tables config.SaltDBTables) {
	h := handler{database: database, tables: tables}
	rg.GET("/jobs/:jid", h.getJob)
}
