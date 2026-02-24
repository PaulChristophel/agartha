package saltCache

import (
	"net/http"
	"net/url"

	"strings"

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
	dbConn := db.DB.Table(table)
	log := logger.GetLogger()
	var saltCache model.SaltCache

	rawBank := c.Param("bank")
	rawRest := c.Param("key")

	bank, key, err := splitBankAndKey(rawBank, rawRest)
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

// Supports both:
//
//	/salt_cache/hello/kitty/foo/bar            -> bank=hello/kitty/foo, key=bar
//	/salt_cache/hello%2Fkitty%2Ffoo/bar        -> bank=hello/kitty/foo, key=bar
//
// With route: /:bank/*key
func splitBankAndKey(rawBank, rawRest string) (string, string, error) {
	// Gin catch-all has leading "/"
	rawRest = strings.TrimPrefix(rawRest, "/")

	bank, err := url.PathUnescape(rawBank)
	if err != nil {
		return "", "", err
	}
	rest, err := url.PathUnescape(rawRest)
	if err != nil {
		return "", "", err
	}

	if bank == "" || rest == "" {
		return "", "", nil
	}

	parts := strings.Split(rest, "/")
	if len(parts) == 0 {
		return "", "", nil
	}

	key := parts[len(parts)-1]
	if key == "" {
		return "", "", nil
	}

	if len(parts) > 1 {
		bankExtra := strings.Join(parts[:len(parts)-1], "/")
		if bankExtra != "" {
			if strings.HasSuffix(bank, "/") {
				bank = bank + bankExtra
			} else {
				bank = bank + "/" + bankExtra
			}
		}
	}

	return bank, key, nil
}
