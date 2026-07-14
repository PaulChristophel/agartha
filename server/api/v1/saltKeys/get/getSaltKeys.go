package saltKeys

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PaulChristophel/agartha/server/api/validate"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func GetSaltKeys(c *gin.Context) {
	dbConn := db.DB.Table(table)
	log := logger.GetLogger()
	var saltKeys []model.SaltKey

	if err := ensureSaltKeysTable(dbConn); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	bank := c.Query("bank")
	key := c.Query("key")
	loadData := c.Query("load_data")
	since := c.Query("since")
	until := c.Query("until")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	orderBy := c.Query("order_by")

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 50
	}
	if perPage > 1000 {
		perPage = 1000
	}

	boolValue, err := strconv.ParseBool(loadData)
	if err != nil {
		boolValue = false
	}
	if boolValue && perPage > 10 {
		perPage = 10
	}

	selection := []string{"bank", "psql_key", "alter_time"}
	if boolValue {
		selection = append(selection, "data")
	}

	baseQuery := dbConn.Select(selection).Model(&model.SaltKey{})
	if bank != "" {
		baseQuery = applyLikeFilter(baseQuery, "bank", bank)
	}
	if key != "" {
		baseQuery = applyLikeFilter(baseQuery, "psql_key", key)
	}
	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format")
			return
		}
		baseQuery = baseQuery.Where("alter_time >= ?", fromTime)
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format")
			return
		}
		baseQuery = baseQuery.Where("alter_time <= ?", toTime)
	}

	validatedOrderBy, err := validate.OrderBy(orderBy, []string{"bank", "psql_key", "alter_time"}, "", []string{})
	if err != nil {
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}
	if validatedOrderBy != "" {
		baseQuery = baseQuery.Order(validatedOrderBy)
	} else {
		baseQuery = baseQuery.Order("bank").Order("psql_key")
	}

	var totalCount int64
	baseQuery.Count(&totalCount)
	baseQuery.Offset((page - 1) * perPage).Limit(perPage).Find(&saltKeys)

	if len(saltKeys) == 0 {
		log.Debug("No salt_keys items present", zap.String("bank", bank), zap.String("key", key))
		httputil.NewError(c, http.StatusNotFound, "No salt_keys items present.")
		return
	}

	baseURL := fmt.Sprintf("%s://%s%s", requestScheme(c), c.Request.Host, c.Request.URL.Path)
	nextPage, previousPage := "", ""
	if page > 1 {
		previousPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page-1, perPage)
	}
	if int64((page-1)*perPage+len(saltKeys)) < totalCount {
		nextPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page+1, perPage)
	}

	c.JSON(http.StatusOK, dto.SaltKeyPageResponse{
		Results: saltKeys,
		Paging: dto.PageResponse{
			PerPage:  int64(perPage),
			NumPages: int64(math.Ceil(float64(totalCount) / float64(perPage))),
			Count:    totalCount,
			Next:     nextPage,
			Previous: previousPage,
		},
	})
}

func applyLikeFilter(query *gorm.DB, column, value string) *gorm.DB {
	if strings.Contains(value, "*") {
		return query.Where(column+" LIKE ?", strings.ReplaceAll(value, "*", "%"))
	}
	if strings.Contains(value, "?") {
		return query.Where(column+" LIKE ?", strings.ReplaceAll(value, "?", "_"))
	}
	return query.Where(column+" = ?", value)
}

func requestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}
