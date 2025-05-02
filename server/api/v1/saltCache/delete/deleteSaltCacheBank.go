package saltCache

import (
	"fmt"
	"net/http"

	// "strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

// DeleteSaltCacheBankKey func one saltCache by SaltCache
//
//	@Summary		Delete a salt_cache bank
//	@Description	Deletes a salt_cache bank and all associated keys
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/bank/{bank} [delete]
//	@Param			bank	path	string	true	"bank of the salt cache item to delete"
//	@Param			key		path	string	true	"key of the salt cache item to delete"
//	@Security		Bearer
func DeleteSaltCacheBank(c *gin.Context) {
	db := db.DB.Table(table)

	bank, ok := httputil.MustUnescapeParam(c, "bank")
	if !ok {
		return
	}

	if bank == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank is required.")
		return
	}

	// Safely delete with explicit WHERE clause only
	tx := db.Where("bank = ?", bank).Delete(&model.SaltCache{})
	if tx.Error != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete salt_cache bank.")
		return
	}
	if tx.RowsAffected == 0 {
		httputil.NewError(c, http.StatusNotFound, fmt.Sprintf("No salt_cache items found for bank: %s", bank))
		return
	}

	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted all salt_cache items in bank: %s", bank))
}
