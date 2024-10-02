package saltCache

import (
	"errors"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	custom "github.com/PaulChristophel/agartha/server/model/custom"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SaltCache struct {
	Bank    string      `json:"bank" example:"minions/server.example.com"`
	PSQLKey string      `json:"psql_key" example:"data"`
	Data    custom.JSON `json:"data"`
}

// CreateSaltCache func to create or update a saltCache entry.
//
//	@Summary		Create or update a salt_cache item.
//	@Description	Create or update a salt_cache item.
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltCache
//	@Success		201	{object}	model.SaltCache
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache [post]
//	@Param			req	body	SaltCache	true	"SaltCache item to create or update"
//	@Security		Bearer
func CreateSaltCache(c *gin.Context) {
	dbConn := db.DB.Table(table)
	log := logger.GetLogger()
	var input SaltCache
	var query model.SaltCache

	// Bind JSON to SaltCacheInput struct
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Error("Invalid input", zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, "Invalid input.")
		return
	}

	// Check if the entry already exists
	result := dbConn.Where("bank = ? AND psql_key = ?", input.Bank, input.PSQLKey).First(&query)
	err := result.Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error("Failed to check existing salt cache entry", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to check existing salt cache entry.")
		return
	}
	// Recreate the DB connection to reset the query
	dbConn = db.DB.Table(table)
	// If entry exists, update it
	if err == nil {
		query.Data = input.Data
		if err := dbConn.Save(&query).Error; err != nil {
			log.Error("Failed to update existing salt cache entry", zap.Error(err))
			httputil.NewError(c, http.StatusInternalServerError, "Failed to update existing salt cache entry.")
			return
		}
		log.Debug("Successfully updated existing salt cache", zap.String("bank/key", query.Bank+"/"+query.PSQLKey))
		c.JSON(http.StatusOK, query)
		return
	}

	saltCache := model.SaltCache{
		Bank:    input.Bank,
		PSQLKey: input.PSQLKey,
		Data:    input.Data,
	}

	// If entry does not exist, create a new one
	if err := dbConn.Create(&saltCache).Error; err != nil {
		log.Error("Failed to create salt cache entry", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to create salt cache entry.")
		return
	}

	log.Debug("Successfully created salt cache", zap.String("bank/key", saltCache.Bank+"/"+saltCache.PSQLKey))
	c.JSON(http.StatusCreated, saltCache)
}
