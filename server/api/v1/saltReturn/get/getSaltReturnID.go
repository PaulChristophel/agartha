package saltReturn

import (
	"net/http"
	"strconv"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetSaltReturnID func get a SaltReturn for a specific jid and id
//
//	@Summary		Get a SaltReturn for a specific jid and id.
//	@Description	Get a SaltReturn for a specific jid and id.
//	@Tags			SaltReturn
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltReturn
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_return/{jid}/{id} [get]
//	@Param			jid				path	string	true	"jid of the salt return item to retrieve"
//	@Param			id				path	string	true	"minion_id of the salt return item to retrieve"
//	@Param			load_return		query	bool	false	"Load the return field. This defaults to false for performance reasons"
//	@Param			load_full_ret	query	bool	false	"Load the full_ret field. This defaults to false for performance reasons"
//	@Security		Bearer
func GetSaltReturnID(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltReturn model.SaltReturn

	jid := c.Param("jid")
	id := c.Param("id")
	loadReturn := c.Query("load_return")
	loadFullRet := c.Query("load_full_ret")

	log.Debug("Received request to get salt return by JID and ID",
		zap.String("jid", jid),
		zap.String("id", id),
		zap.String("load_return", loadReturn),
		zap.String("load_full_ret", loadFullRet))

	// Parse bool for loading data
	boolLoadReturn, err := strconv.ParseBool(loadReturn)
	if err != nil {
		boolLoadReturn = false
		log.Debug("Invalid load_return value, defaulting to false", zap.String("load_return", loadReturn), zap.Error(err))
	}
	boolFullRet, err := strconv.ParseBool(loadFullRet)
	if err != nil {
		boolFullRet = false
		log.Debug("Invalid load_full_ret value, defaulting to false", zap.String("load_full_ret", loadFullRet), zap.Error(err))
	}

	// Define selection fields
	selection := []string{"fun", "jid", "id", "success", "alter_time"}
	if boolLoadReturn {
		selection = append(selection, "return")
	}
	if boolFullRet {
		selection = append(selection, "full_ret")
	}

	// Find the saltReturn in the database with the specified jid and id
	result := db.Select(selection).Where(&model.SaltReturn{JID: jid, ID: id}).First(&saltReturn)

	// If no saltReturn is present return an error
	if result.RowsAffected == 0 {
		log.Debug("No salt_returns present", zap.String("jid", jid), zap.String("id", id))
		httputil.NewError(c, http.StatusNotFound, "No salt_returns present.")
		return
	}

	log.Debug("Returning salt return by JID and ID", zap.String("jid", jid), zap.String("id", id))
	// Else return saltReturn
	c.JSON(http.StatusOK, saltReturn)
}
