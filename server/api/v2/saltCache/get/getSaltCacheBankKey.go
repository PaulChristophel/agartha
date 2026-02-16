package saltCache

import (
	"net/http"
	"net/url"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strings"
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
//	@router			/api/v2/salt_cache/{key}/{bank} [get]
//	@Param			key		path	string	true	"key of the salt cache item to retrieve"
//	@Param			bank	path	string	true	"bank of the salt cache item to retrieve"
//	@Security		Bearer

func GetSaltCacheBankKey(c *gin.Context) {
	dbConn := db.DB.Table(table)
	log := logger.GetLogger()
	var saltCache model.SaltCache

	rawKey := c.Param("key")
	rawBank := c.Param("bank")

	key, bank, err := splitKeyAndBank(rawKey, rawBank)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}
	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}

	log.Debug("Received parameters", zap.String("bank", bank), zap.String("key", key))

	if err := dbConn.Where("bank = ? AND psql_key = ?", bank, key).Find(&saltCache).Error; err != nil {
		log.Error("Failed to fetch salt cache data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt cache data.")
		return
	}

	if saltCache.Bank == "" {
		log.Debug("No salt cache item present", zap.String("bank", bank), zap.String("key", key))
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	c.JSON(http.StatusOK, saltCache)
}

func splitKeyAndBank(rawKey, rawBank string) (string, string, error) {
	// key is always a single segment in v2, but still unescape it
	key, err := url.PathUnescape(rawKey)
	if err != nil {
		return "", "", err
	}

	// bank may be "hello/kitty/foo" OR "hello%2Fkitty%2Ffoo"
	// If route uses /*bank, Gin gives a leading "/" for the catch-all.
	rawBank = strings.TrimPrefix(rawBank, "/")

	bank, err := url.PathUnescape(rawBank)
	if err != nil {
		return "", "", err
	}

	if key == "" || bank == "" {
		return "", "", nil
	}

	// Normalize: if caller accidentally passed "key" as a catch-all and it contains slashes,
	// treat the last segment as key and prefix the rest onto bank.
	// This keeps behavior consistent if you later use routes like /*key/:bank.
	key = strings.TrimPrefix(key, "/")
	if strings.Contains(key, "/") {
		parts := strings.Split(key, "/")
		last := parts[len(parts)-1]
		prefix := strings.Join(parts[:len(parts)-1], "/")
		if prefix != "" {
			if strings.HasSuffix(bank, "/") {
				bank = bank + prefix
			} else {
				bank = bank + "/" + prefix
			}
		}
		key = last
	}

	return key, bank, nil
}
