package saltKeys

import (
	delete "github.com/PaulChristophel/agartha/server/api/v1/saltKeys/delete"
	get "github.com/PaulChristophel/agartha/server/api/v1/saltKeys/get"
	post "github.com/PaulChristophel/agartha/server/api/v1/saltKeys/post"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AddRoutes registers v1 salt_keys API routes.
func AddRoutes(rg *gin.RouterGroup, database *gorm.DB) {
	grp := rg.Group("/salt_keys")
	rawKeyAdmin := middleware.SaltWheelAdministrationRequired(database)

	grp.GET("", rawKeyAdmin, get.GetSaltKeys)
	grp.GET("/minion_keys", middleware.SaltWheelPermissionRequired(database, "key.list_all"), get.GetMinionKeys)
	grp.GET("/:bank/*key", rawKeyAdmin, get.GetSaltKeysBankKey)
	grp.POST("", rawKeyAdmin, post.CreateSaltKey)
	grp.POST("/minion_keys/accept", middleware.SaltWheelPermissionRequired(database, "key.accept"), post.AcceptMinionKeys)
	grp.POST("/minion_keys/reject", middleware.SaltWheelPermissionRequired(database, "key.reject"), post.RejectMinionKeys)
	grp.POST("/minion_keys/delete", middleware.SaltWheelPermissionRequired(database, "key.delete"), post.DeleteMinionKeys)
	grp.DELETE("/bank/:bank", rawKeyAdmin, delete.DeleteSaltKeysBank)
	grp.DELETE("/:bank/*key", rawKeyAdmin, delete.DeleteSaltKeysBankKey)
}

// SetOptions configures the database table used by the salt_keys API.
func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
	post.SetOptions(saltTables)
	delete.SetOptions(saltTables)
}
