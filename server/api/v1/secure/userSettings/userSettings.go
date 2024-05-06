package userSettings

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/user_settings")

	// grp.GET("/", GetUserSettings)
	grp.GET("/:id", GetUserSetting)
}

// GetUserSetting func one userSetting by UserSetting
//
//	@Description	Get one userSetting by UserSetting ID
//	@Tags			UserSetting
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.UserSettings
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/secure/user_settings/{id} [get]
//	@Param			id	path	string	false	"userSettings to return"
//	@Security		Bearer
func GetUserSetting(c *gin.Context) {
	db := db.DB
	var userSetting model.UserSettings

	// Read the param userSetting
	id := c.Param("id")

	// Find the userSetting with the given Id
	db.Find(&userSetting, "user_id = ?", id)

	// If no such userSetting present return an error
	if userSetting.UserID <= 0 {
		httputil.NewError(c, http.StatusNotFound, "No user_setting data present.")
		return
	}

	// Return the userSetting with the Id
	c.JSON(http.StatusOK, userSetting)
}
