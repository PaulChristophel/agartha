package saltKeys

import (
	delete "github.com/PaulChristophel/agartha/server/api/v1/saltKeys/delete"
	get "github.com/PaulChristophel/agartha/server/api/v1/saltKeys/get"
	post "github.com/PaulChristophel/agartha/server/api/v1/saltKeys/post"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/gin-gonic/gin"
)

// AddRoutes registers v1 salt_keys API routes.
func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_keys")

	grp.GET("", get.GetSaltKeys)
	grp.GET("/minion_keys", get.GetMinionKeys)
	grp.GET("/:bank/*key", get.GetSaltKeysBankKey)
	grp.POST("", post.CreateSaltKey)
	grp.POST("/minion_keys/accept", post.AcceptMinionKeys)
	grp.POST("/minion_keys/reject", post.RejectMinionKeys)
	grp.POST("/minion_keys/delete", post.DeleteMinionKeys)
	grp.DELETE("/bank/:bank", delete.DeleteSaltKeysBank)
	grp.DELETE("/:bank/*key", delete.DeleteSaltKeysBankKey)
}

// SetOptions configures the database table used by the salt_keys API.
func SetOptions(saltTables config.SaltDBTables) {
	get.SetOptions(saltTables)
	post.SetOptions(saltTables)
	delete.SetOptions(saltTables)
}
