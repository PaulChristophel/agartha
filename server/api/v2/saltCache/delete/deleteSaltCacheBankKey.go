package saltCache

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

// DeleteSaltCacheBankKey func one saltCache by SaltCache
//
//	@Summary		Delete one salt_cache item by bank and key
//	@Description	Delete one salt_cache item by bank and key
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v2/salt_cache/{key}/{bank} [delete]
//	@Param			key		path	string	true	"key of the salt cache item to delete"
//	@Param			bank	path	string	true	"bank of the salt cache item to delete"
//	@Security		Bearer
func DeleteSaltCacheBankKey(c *gin.Context) {
	dbConn := db.DB.Table(table)

	rawKey := c.Param("key")
	rawBank := c.Param("bank")

	key, bank, err := splitKeyAndBank(rawKey, rawBank)
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, "bank and key are required.")
		return
	}
	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key are required.")
		return
	}

	tx := dbConn.Where("bank = ? AND psql_key = ?", bank, key).Delete(&model.SaltCache{})
	if tx.Error != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete cache item.")
		return
	}
	if tx.RowsAffected == 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item found to delete.")
		return
	}

	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted salt_cache item %s/%s", bank, key))
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
