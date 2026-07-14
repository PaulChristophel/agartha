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

type MinionKeysResponse struct {
	Minions         []string `json:"minions"`
	MinionsPre      []string `json:"minions_pre"`
	MinionsRejected []string `json:"minions_rejected"`
	MinionsDenied   []string `json:"minions_denied"`
	Local           []string `json:"local"`
}

func GetMinionKeys(c *gin.Context) {
	keyList, err := listMinionKeys(db.DB.Table(table))
	if err != nil {
		writeSaltKeysError(c, err)
		return
	}

	c.JSON(http.StatusOK, keyList)
}

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
