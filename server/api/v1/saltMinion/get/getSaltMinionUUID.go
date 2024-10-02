package saltMinion

import (
	"net/http"

	// "strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt/view"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltMinion func one saltMinion by the item's UUID
//
//	@Summary		Get one salt_minion item by UUID
//	@Description	Get one salt_minion item by UUID
//	@Tags			SaltMinion
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltMinion
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_minion/uuid/{uuid} [get]
//	@Param			uuid	path	string	true	"uuid of the salt minion item to retrieve"
//	@Security		Bearer
func GetSaltMinionUUID(c *gin.Context) {
	db := db.DB
	log := logger.GetLogger()
	var saltMinion model.SaltMinion

	// Read the param saltMinion
	uuid := c.Param("uuid")

	if uuid == "" {
		log.Error("No uuid provided", zap.String("uuid", uuid))
		httputil.NewError(c, http.StatusNotFound, "uuid required.")
		return
	}

	// Find the saltMinion with the given UUID
	if err := db.Where("id = ?", uuid).Find(&saltMinion).Find(&saltMinion).Error; err != nil {
		log.Error("Failed to fetch salt minion data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt minion data.")
		return
	}

	// If no such saltMinion is present, return an error
	if saltMinion.MinionID == "" {
		log.Debug("No salt minion item present", zap.String("uuid", uuid))
		httputil.NewError(c, http.StatusNotFound, "No salt_minion item present.")
		return
	}

	log.Debug("Successfully retrieved salt minion", zap.String("uuid", uuid))
	// Return the saltMinion with the UUID
	c.JSON(http.StatusOK, saltMinion)
}
