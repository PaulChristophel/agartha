package saltKeys

import (
	"fmt"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

// DeleteSaltKeysBankKey deletes one salt_keys item by bank and key.
//
//	@Summary		Delete one salt_keys item by bank and key.
//	@Description	Delete one salt_keys row identified by bank and key.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/{bank}/{key} [delete]
//	@Param			bank	path	string	true	"Bank of the salt key item to delete."
//	@Param			key		path	string	true	"Key of the salt key item to delete."
//	@Security		Bearer
func DeleteSaltKeysBankKey(c *gin.Context) {
	dbConn := db.DB.Table(table)

	if err := ensureSaltKeysTable(dbConn); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	bank, key, err := splitBankAndKey(c.Param("bank"), c.Param("key"))
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, "bank and key are required.")
		return
	}
	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key are required.")
		return
	}

	tx := dbConn.Where("bank = ? AND psql_key = ?", bank, key).Delete(&model.SaltKey{})
	if tx.Error != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete salt_keys item.")
		return
	}
	if tx.RowsAffected == 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_keys item found to delete.")
		return
	}

	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted salt_keys item %s/%s", bank, key))
}
