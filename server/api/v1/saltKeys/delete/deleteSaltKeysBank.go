package saltKeys

import (
	"fmt"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

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
