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

// GetSaltKeys retrieves a paginated list of salt_keys rows.
//
//	@Summary		Retrieve a list of salt_keys items (paginated).
//	@Description	Retrieve salt_keys rows filtered by bank and key with pagination support.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltKeyPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys [get]
//	@Param			bank		query	string	false	"Bank of the salt key item to search (supports wildcards * and ?)."
//	@Param			key			query	string	false	"Key of the salt key item to search (supports wildcards * and ?)."
//	@Param			load_data	query	bool	false	"Load the data field. This defaults to false for performance reasons."
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page		query	int		false	"Page number of results to retrieve."
//	@Param			per_page	query	int		false	"Number of items per page."
//	@Param			order_by	query	string	false	"Order by columns (e.g. bank,psql_key desc,alter_time asc)."
//	@Security		Bearer
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

// applyLikeFilter applies exact or wildcard filtering to a GORM query.
func applyLikeFilter(query *gorm.DB, column, value string) *gorm.DB {
	if strings.Contains(value, "*") {
		return query.Where(column+" LIKE ?", strings.ReplaceAll(value, "*", "%"))
	}
	if strings.Contains(value, "?") {
		return query.Where(column+" LIKE ?", strings.ReplaceAll(value, "?", "_"))
	}
	return query.Where(column+" = ?", value)
}

// requestScheme returns the request scheme used to construct pagination links.
func requestScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}
