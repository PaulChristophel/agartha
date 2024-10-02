package authUser

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetauthUser func one authUser by AuthUser
//
//	@Description	Get one authUser by AuthUser ID
//	@Tags			AuthUser
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.AuthUser
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/secure/auth_user/{id} [get]
//	@Param			id	path	string	false	"authUsers to return"
//	@Security		Bearer
func GetAuthUser(c *gin.Context) {
	db := db.DB
	log := logger.GetLogger()
	var authUser model.AuthUser

	// Read the param authUser
	id := c.Param("id")
	log.Debug("Received request to get auth user", zap.String("id", id))

	// Find the authUser with the given Id
	db.Find(&authUser, "id = ?", id)

	// If no such authUser present return an error
	if authUser.ID <= 0 {
		log.Debug("No auth_user data present", zap.String("id", id))
		httputil.NewError(c, http.StatusNotFound, "No auth_user data present.")
		return
	}

	authUser.Password = "" // Scrub the password field.

	log.Debug("Returning auth user data", zap.String("id", id))
	// Return the authUser with the Id
	c.JSON(http.StatusOK, authUser)
}
