package jid

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetJID func one jid by JID
//
//	@Summary		Get Job data about a specific Job.
//	@Description	Get one jid by JID
//	@Tags			JID
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.JID
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/jid/{jid} [get]
//	@Param			jid	path	string	false	"jid to return"
//	@Security		Bearer
func GetJID(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var jids model.JID

	// Read the param jid
	id := c.Param("jid")
	log.Debug("Received request to get jid", zap.String("jid", id))

	// find all HighStates in the database with the specified id
	if err := db.Where("jid = ?", id).Find(&jids).Error; err != nil {
		log.Error("Failed to fetch jid data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch jid data.")
		return
	}

	// If no such jid present return an error
	if jids.JID != id {
		log.Debug("No jids present", zap.String("jid", id))
		httputil.NewError(c, http.StatusNotFound, "No jid present.")
		return
	}

	log.Debug("Successfully retrieved jid", zap.String("jid", id))
	// Return the jid with the Id
	c.JSON(http.StatusOK, jids)
}
