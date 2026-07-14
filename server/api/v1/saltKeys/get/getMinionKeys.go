package saltKeys

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	keysBank   = "pki/master/keys"
	deniedBank = "pki/master/denied_keys"
)

// MinionKeysResponse is the Salt wheel-compatible minion key list response.
type MinionKeysResponse struct {
	Minions         []string `json:"minions"`
	MinionsPre      []string `json:"minions_pre"`
	MinionsRejected []string `json:"minions_rejected"`
	MinionsDenied   []string `json:"minions_denied"`
	Local           []string `json:"local"`
}

// GetMinionKeys returns salt_keys rows grouped like Salt's key.list_all wheel response.
//
//	@Summary		Retrieve minion keys grouped by state.
//	@Description	Return minion keys from salt_keys in the same shape as Salt's key.list_all wheel response.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	MinionKeysResponse
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/minion_keys [get]
//	@Security		Bearer
func GetMinionKeys(c *gin.Context) {
	keyList, err := listMinionKeys(db.DB.Table(table))
	if err != nil {
		writeSaltKeysError(c, err)
		return
	}

	c.JSON(http.StatusOK, keyList)
}

// listMinionKeys loads minion keys from the configured salt_keys table.
func listMinionKeys(dbConn *gorm.DB) (MinionKeysResponse, error) {
	if err := ensureSaltKeysTable(dbConn); err != nil {
		return MinionKeysResponse{}, err
	}

	var rows []model.SaltKey
	if err := dbConn.Where("bank = ?", keysBank).Order("psql_key").Find(&rows).Error; err != nil {
		return MinionKeysResponse{}, err
	}

	var deniedRows []model.SaltKey
	if err := dbConn.Where("bank = ?", deniedBank).Order("psql_key").Find(&deniedRows).Error; err != nil {
		return MinionKeysResponse{}, err
	}

	resp := MinionKeysResponse{
		Minions:         []string{},
		MinionsPre:      []string{},
		MinionsRejected: []string{},
		MinionsDenied:   []string{},
		Local:           []string{},
	}
	for _, row := range rows {
		switch keyState(row.Data.Data) {
		case "accepted":
			resp.Minions = append(resp.Minions, row.PSQLKey)
		case "pending":
			resp.MinionsPre = append(resp.MinionsPre, row.PSQLKey)
		case "rejected":
			resp.MinionsRejected = append(resp.MinionsRejected, row.PSQLKey)
		}
	}
	for _, row := range deniedRows {
		resp.MinionsDenied = append(resp.MinionsDenied, row.PSQLKey)
	}

	return resp, nil
}
