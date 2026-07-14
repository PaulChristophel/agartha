package saltKeys

import (
	"errors"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	custom "github.com/PaulChristophel/agartha/server/model/custom"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SaltKey struct {
	Bank    string      `json:"bank" example:"pki/master/keys"`
	PSQLKey string      `json:"psql_key" example:"server.example.com"`
	Data    custom.JSON `json:"data"`
}

func CreateSaltKey(c *gin.Context) {
	dbConn := db.DB.Table(table)
	var input SaltKey
	var query model.SaltKey

	if err := ensureSaltKeysTable(dbConn); err != nil {
		writeSaltKeysError(c, err)
		return
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.NewError(c, http.StatusBadRequest, "Invalid input.")
		return
	}
	if input.Bank == "" || input.PSQLKey == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and psql_key are required.")
		return
	}

	result := dbConn.Where("bank = ? AND psql_key = ?", input.Bank, input.PSQLKey).First(&query)
	err := result.Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to check existing salt key entry.")
		return
	}
	if err == nil {
		query.Data = input.Data
		if err := dbConn.Save(&query).Error; err != nil {
			httputil.NewError(c, http.StatusInternalServerError, "Failed to update existing salt key entry.")
			return
		}
		c.JSON(http.StatusOK, query)
		return
	}

	saltKey := model.SaltKey{
		Bank:    input.Bank,
		PSQLKey: input.PSQLKey,
		Data:    input.Data,
	}
	if err := dbConn.Create(&saltKey).Error; err != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to create salt key entry.")
		return
	}

	c.JSON(http.StatusCreated, saltKey)
}
