package saltMinion

import (
	"net/http"
	"net/url"

	// "strings"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt/view"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltMinion func one saltCache by SaltMinion
//
//	@Summary		Get one salt_minion item by minion_id.
//	@Description	Get one salt_minion item by minion_id.
//	@Tags			SaltMinion
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltMinion
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_minion/{minion_id} [get]
//	@Param			minion_id	path	string	true	"minion_id of the salt minion item to retrieve"
//	@Security		Bearer
func GetSaltMinionID(c *gin.Context) {
	db := db.DB
	log := logger.GetLogger()
	var saltMinion model.SaltMinion

	// Read the param saltCache
	minionID, _ := url.PathUnescape(c.Param("minion_id"))

	if minionID == "" {
		log.Error("No minion_id provided", zap.String("minion_id", minionID))
		httputil.NewError(c, http.StatusBadRequest, "minion_id required.")
		return
	}

	// Find the saltCache with the given minionID
	if err := db.Where("minion_id = ?", minionID).Find(&saltMinion).Error; err != nil {
		log.Error("Failed to fetch salt minion data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt minion data.")
		return
	}

	// If no such saltCache present return an error
	if saltMinion.MinionID == "" {
		log.Debug("No salt minion item present", zap.String("minion_id", minionID))
		httputil.NewError(c, http.StatusNotFound, "No salt_minion item present.")
		return
	}

	log.Debug("Successfully retrieved salt minion", zap.String("minion_id", minionID))
	// Return the saltCache with the Id
	c.JSON(http.StatusOK, saltMinion)
}
