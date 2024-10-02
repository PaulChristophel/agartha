package highState

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt/view"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetHighState func get most recent HighState for a specific id
//
//	@Summary		Get most recent HighState for a specific minion id.
//	@Description	Get most recent HighState for a specific minion id.
//	@Tags			HighState
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.HighState
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/high_state/{id} [get]
//	@Param			id	path	string	true	"id of the salt return item to retrieve"
//	@Security		Bearer
func GetHighState(c *gin.Context) {
	db := db.DB
	log := logger.GetLogger()
	var highStates []model.HighState

	id := c.Param("id")
	log.Debug("Received request to get high state", zap.String("id", id))

	// find all HighStates in the database with the specified id
	if err := db.Where("id = ?", id).Find(&highStates).Error; err != nil {
		log.Error("Failed to fetch high state data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch high state data.")
		return
	}

	// If no HighState is present return an error
	if len(highStates) == 0 {
		log.Debug("No high_states present", zap.String("id", id))
		httputil.NewError(c, http.StatusNotFound, "No high_states present.")
		return
	}

	log.Debug("Successfully retrieved high state", zap.String("id", id), zap.Int("record_count", len(highStates)))
	// Else return HighStates
	c.JSON(http.StatusOK, highStates[0])
}
