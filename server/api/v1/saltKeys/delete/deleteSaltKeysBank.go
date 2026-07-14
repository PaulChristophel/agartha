package saltKeys

import (
	"fmt"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

// DeleteSaltKeysBank deletes all salt_keys items in a bank.
//
//	@Summary		Delete a salt_keys bank.
//	@Description	Delete all salt_keys rows associated with a bank.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/bank/{bank} [delete]
//	@Param			bank	path	string	true	"Bank of the salt_keys items to delete."
//	@Security		Bearer
func DeleteSaltKeysBank(c *gin.Context) {
	dbConn := db.DB.Table(table)

	if err := ensureSaltKeysTable(dbConn); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	bank, ok := httputil.MustUnescapeParam(c, "bank")
	if !ok {
		return
	}
	if bank == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank is required.")
		return
	}

	tx := dbConn.Where("bank = ?", bank).Delete(&model.SaltKey{})
	if tx.Error != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete salt_keys bank.")
		return
	}
	if tx.RowsAffected == 0 {
		httputil.NewError(c, http.StatusNotFound, fmt.Sprintf("No salt_keys items found for bank: %s", bank))
		return
	}

	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted all salt_keys items in bank: %s", bank))
}
