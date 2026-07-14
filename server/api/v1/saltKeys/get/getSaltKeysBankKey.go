package saltKeys

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltKeysBankKey retrieves one salt_keys item by bank and key.
//
//	@Summary		Get one salt_keys item by bank and key.
//	@Description	Get one salt_keys item by bank and key.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltKey
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/{bank}/{key} [get]
//	@Param			bank	path	string	true	"Bank of the salt key item to retrieve."
//	@Param			key		path	string	true	"Key of the salt key item to retrieve."
//	@Security		Bearer
func GetSaltKeysBankKey(c *gin.Context) {
	dbConn := db.DB.Table(table)
	log := logger.GetLogger()
	var saltKey model.SaltKey

	if err := ensureSaltKeysTable(dbConn); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	bank, key, err := splitBankAndKey(c.Param("bank"), c.Param("key"))
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}
	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}

	if err := dbConn.Where("bank = ? AND psql_key = ?", bank, key).Find(&saltKey).Error; err != nil {
		log.Error("Failed to fetch salt key data", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt key data.")
		return
	}
	if saltKey.Bank == "" {
		httputil.NewError(c, http.StatusNotFound, "No salt_keys item present.")
		return
	}

	c.JSON(http.StatusOK, saltKey)
}
