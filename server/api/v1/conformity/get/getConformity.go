package conformity

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt/view"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetConformity func get all GetConformity
//
//	@Summary		Get conformity data about a specific minion.
//	@Description	Get all GetConformity
//	@Tags			Conformity
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.Conformity
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/conformity/{id} [get]
//	@Param			id	path	string	false	"conformity object to get"
//	@Security		Bearer
func GetConformity(c *gin.Context) {
	log := logger.GetLogger()
	db := db.DB
	var conformity model.Conformity

	id := c.Param("id")

	// find all jids in the database
	if err := db.Where("id = ?", id).Find(&conformity).Error; err != nil {
		log.Error("Failed to fetch conformity data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch conformity data.")
		return
	}

	// If no jid is present return an error
	if conformity.ID != id {
		log.Debug("No conformity item present", zap.String("id", id))
		httputil.NewError(c, http.StatusNotFound, "No conformity item present.")
		return
	}

	log.Debug("Conformity data fetched successfully", zap.String("id", id))
	// Else return jids
	c.JSON(http.StatusOK, conformity)
}
