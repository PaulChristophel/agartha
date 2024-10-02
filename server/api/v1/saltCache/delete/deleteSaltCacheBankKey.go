package saltCache

import (
	"fmt"
	"net/http"
	"net/url"

	// "strings"

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
//	@router			/api/v1/salt_cache/{bank}/{key} [delete]
//	@Param			bank	path	string	true	"bank of the salt cache item to delete"
//	@Param			key		path	string	true	"key of the salt cache item to delete"
//	@Security		Bearer
func DeleteSaltCacheBankKey(c *gin.Context) {
	db := db.DB.Table(table)
	var saltCache model.SaltCache

	// Read the param saltCache
	bank, _ := url.PathUnescape(c.Param("bank"))
	key, _ := url.PathUnescape(c.Param("key"))

	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}

	// Find the saltCache with the given bank
	db.Where("bank = ? AND psql_key = ?", bank, key).Find(&saltCache)

	// If no such saltCache present return an error
	if saltCache.Bank == "" {
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	err := db.Delete(&saltCache, "bank = ? AND psql_key = ?", bank, key).Error

	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete cache item.")
		return
	}

	// Return success message
	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted salt_cache item %s/%s", bank, key))
}
