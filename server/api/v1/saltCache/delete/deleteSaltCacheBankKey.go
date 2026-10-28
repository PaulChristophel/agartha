package saltCache

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PaulChristophel/sylvite/server/db"
	"github.com/PaulChristophel/sylvite/server/httputil"
	model "github.com/PaulChristophel/sylvite/server/model/salt"
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
	dbConn := db.DB.Table(table)

	rawBank := c.Param("bank")
	rawRest := c.Param("key")

	bank, key, err := splitBankAndKey(rawBank, rawRest)
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

func splitBankAndKey(rawBank, rawRest string) (string, string, error) {
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
