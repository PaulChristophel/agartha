package saltCache

import (
	"net/http"
	"net/url"

	// "strings"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltCacheBankKey func one saltCache by SaltCache
//
//	@Summary		Get one salt_cache item by bank and key.
//	@Description	Get one salt_cache item by bank and key.
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltCache
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/{bank}/{key} [get]
//	@Param			bank	path	string	true	"bank of the salt cache item to retrieve"
//	@Param			key		path	string	true	"key of the salt cache item to retrieve"
//	@Security		Bearer
func GetSaltCacheBankKey(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltCache model.SaltCache

	// Read the parameters
	rawBank := c.Param("bank")
	rawKey := c.Param("key")

	// Log the raw parameters
	log.Debug("Raw parameters", zap.String("bank", rawBank), zap.String("key", rawKey))

	// Read the param saltCache
	bank, err := url.PathUnescape(rawBank)
	if err != nil {
		log.Error("Error unescaping path", zap.String("bank", bank))
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}
	key, err := url.PathUnescape(rawKey)
	if err != nil {
		log.Error("Error unescaping key", zap.String("key", key))
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}

	log.Debug("Received parameters", zap.String("bank", bank), zap.String("key", key))

	if bank == "" || key == "" {
		log.Error("No bank and/or key provided", zap.String("bank", bank), zap.String("key", key))
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}

	// Find the saltCache with the given bank
	if err := db.Where("bank = ? AND psql_key = ?", bank, key).Find(&saltCache).Error; err != nil {
		log.Error("Failed to fetch salt cache data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt cache data.")
		return
	}

	// If no such saltCache present return an error
	if saltCache.Bank == "" {
		log.Debug("No salt cache item present", zap.String("bank", bank), zap.String("key", key))
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	log.Debug("Successfully retrieved salt cache", zap.String("bank", bank), zap.String("key", key))
	// Return the saltCache with the Id
	c.JSON(http.StatusOK, saltCache)
}
