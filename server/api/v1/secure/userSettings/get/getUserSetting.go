package userSettings

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

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
func GetUserSettings(c *gin.Context) {
	db := db.DB
	log := logger.GetLogger()
	var userSetting model.UserSettings

	// Read the param userSetting
	id := c.Param("id")
	log.Debug("Received request to get user settings", zap.String("id", id))

	// Find the userSetting with the given Id
	db.Find(&userSetting, "user_id = ?", id)

	// If no such userSetting present return an error
	if userSetting.UserID <= 0 {
		log.Debug("No user_setting data present", zap.String("user_id", id))
		httputil.NewError(c, http.StatusNotFound, "No user_setting data present.")
		return
	}

	log.Debug("Returning user settings data", zap.String("user_id", id))
	// Return the userSetting with the Id
	c.JSON(http.StatusOK, userSetting)
}
