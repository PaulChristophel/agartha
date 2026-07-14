package saltKeys

import (
	"fmt"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

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
