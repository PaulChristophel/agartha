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

// GetSaltReturnJID func get SaltReturns for a specific jid
//
//	@Summary		Get all SaltReturns for a specific jid.
//	@Description	Get all SaltReturns for a specific jid.
//	@Tags			SaltReturn
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.SaltReturn
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_return/{jid} [get]
//	@Param			jid				path	string	true	"jid of the salt return item to retrieve"
//	@Param			load_return		query	bool	false	"Load the return field. This defaults to false for performance reasons"
//	@Param			load_full_ret	query	bool	false	"Load the full_ret field. This defaults to false for performance reasons"
//	@Security		Bearer
func GetSaltReturnJID(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltReturns []model.SaltReturn

	jid := c.Param("jid")
	loadReturn := c.Query("load_return")
	loadFullRet := c.Query("load_full_ret")

	log.Debug("Received request to get salt returns by JID",
		zap.String("jid", jid),
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

	// find all saltReturns in the database with the specified jid
	db.Select(selection).Where("jid = ?", jid).Find(&saltReturns)

	// If no saltReturn is present return an error
	if len(saltReturns) == 0 {
		log.Debug("No salt_returns present", zap.String("jid", jid))
		httputil.NewError(c, http.StatusNotFound, "No salt_returns present.")
		return
	}

	log.Debug("Returning salt returns by JID", zap.Int("count", len(saltReturns)))
	// Else return saltReturns
	c.JSON(http.StatusOK, saltReturns)
}
