package saltEvent

import (
	"net/http"
	"strconv"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltEvent func get all SaltEvents
//
//	@Summary		Get a salt event by ID.
//	@Description	Get a salt event by ID
//	@Tags			SaltEvent
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltEvent
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_event/{id} [get]
//	@Param			id	path	int	false	"return salt event with given id"
//	@Security		Bearer
func GetSaltEvent(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltEvent model.SaltEvent

	// Read the param id
	id := c.Param("id")
	log.Debug("Received request to get event", zap.String("id", id))

	// Convert id to int64
	idInt64, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		log.Debug("Invalid id value", zap.String("id", id))
		httputil.NewError(c, http.StatusBadRequest, "Invalid event id: "+id+".")
		return
	}

	// Find the event with the given id
	if err := db.Find(&saltEvent, "id = ?", idInt64).Error; err != nil {
		log.Error("Failed to fetch event data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch event data.")
		return
	}

	// If no event is present, return an error
	if saltEvent.ID <= 0 {
		log.Debug("No salt_events present for id", zap.Int64("id", idInt64))
		httputil.NewError(c, http.StatusNotFound, "No salt_events present.")
		return
	}

	// Log the successful retrieval of the event
	log.Info("Successfully retrieved event", zap.Int64("id", idInt64))

	// Return the event data
	c.JSON(http.StatusOK, saltEvent)
}
