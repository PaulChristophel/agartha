package saltCache

import (
	"net/http"

	// "strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltCache func one saltCache by the item's UUID
//
//	@Summary		Get one salt_cache item by UUID
//	@Description	Get one salt_cache item by UUID
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltCache
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/uuid/{uuid} [get]
//	@Param			uuid	path	string	true	"uuid of the salt cache item to retrieve"
//	@Security		Bearer
func GetSaltCacheUUID(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltCache model.SaltCache

	// Read the param saltCache
	uuid := c.Param("uuid")

	if uuid == "" {
		log.Error("No uuid provided", zap.String("uuid", uuid))
		httputil.NewError(c, http.StatusNotFound, "uuid required.")
		return
	}

	// Find the saltCache with the given UUID
	if err := db.Where("id = ?", uuid).Find(&saltCache).Find(&saltCache).Error; err != nil {
		log.Error("Failed to fetch salt cache data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt cache data.")
		return
	}

	// If no such saltCache is present, return an error
	if saltCache.Bank == "" {
		log.Debug("No salt cache item present", zap.String("uuid", uuid))
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	log.Debug("Successfully retrieved salt cache", zap.String("uuid", uuid))
	// Return the saltCache with the UUID
	c.JSON(http.StatusOK, saltCache)
}
