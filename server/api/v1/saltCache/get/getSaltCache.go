package saltCache

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PaulChristophel/agartha/server/api/jsonPathFilter"
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

// GetSaltCache func retrieves a paginated list of saltCache items filtered by bank and key
//
//	@Summary		Retrieve a list of salt_cache items (paginated).
//	@Description	Retrieve a list of salt_cache items filtered by bank and key with pagination support.
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltCachePageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache [get]
//	@Param			bank			query	string	false	"Bank of the cached item to search (Supports wildcards * and ? for single char matches.)"
//	@Param			key				query	string	false	"Key of the cached item to search (Supports wildcards * and ? for single char matches.)"
//	@Param			load_data		query	bool	false	"Load the data field. This defaults to false for performance reasons. (Implied false if jsonpath != '')""
//	@Param			jsonpath		query	string	false	"Comma separated list of data items to return as subset of data in jsonpath syntax (e.g. grains.os,grains.kernel,grains.efi)"
//	@Param			jsonpath_filter	query	string	false	"Comma separated list of data items to filter on (e.g. grains.os:RedHat::string,grains.kernel:Linux::string,grains.efi:true::bool)"
//	@Param			since			query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until			query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page			query	int		false	"Page number of results to retrieve"
//	@Param			per_page		query	int		false	"Number of items per page"
//	@Param			order_by		query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. bank,psql_key desc,alter_time asc,id)"
//	@Security		Bearer
func GetSaltCache(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltCaches []model.SaltCache

	// Read query parameters
	bank := c.Query("bank")
	key := c.Query("key")
	loadData := c.Query("load_data")
	jsonpath := c.Query("jsonpath")
	jsonpathFilter := c.Query("jsonpath_filter")
	since := c.Query("since")
	until := c.Query("until")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	orderBy := c.Query("order_by")

	log.Debug("Received request to get salt cache",
		zap.String("bank", bank),
		zap.String("key", key),
		zap.String("load_data", loadData),
		zap.String("jsonpath", jsonpath),
		zap.String("jsonpath_filter", jsonpathFilter),
		zap.String("since", since),
		zap.String("until", until),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
		zap.String("order_by", orderBy))

	// Parse bool for loading data
	boolValue, err := strconv.ParseBool(loadData)
	if err != nil {
		boolValue = false
		log.Debug("Invalid load_data value, defaulting to false", zap.String("load_data", loadData), zap.Error(err))
	}
	// If we are filtering a subset of data, loading it all is redundant.
	if jsonpath != "" {
		boolValue = false
	}
	log.Debug("load_data option set", zap.Bool("boolValue", boolValue))

	if perPage > 1000 {
		perPage = 1000
		log.Debug("PerPage exceeds maximum, setting to 1000")
	}
	if boolValue && perPage > 10 {
		perPage = 10
		log.Debug("PerPage exceeds maximum for detailed data, setting to 10")
	}

	// Define selection fields
	selection := []string{"bank", "psql_key", "id", "alter_time"}
	if boolValue {
		selection = append(selection, "data")
	} else if jsonpath != "" {
		// Apply subData filters if a
		subDataPaths := strings.Split(jsonpath, ",")
		jsonPathQuery := jsonPathFilter.BuildJSONPathSelect(subDataPaths, "data")
		expr := gorm.Expr(jsonPathQuery)
		selection = append(selection, expr.SQL)
	}

	log.Debug("Using selection: ", zap.Strings("selection", selection))

	// Create a base query with selected fields
	baseQuery := db.Select(selection).Model(&model.SaltCache{})

	// Apply filters based on provided bank and key
	if bank != "" {
		if strings.Contains(bank, "*") {
			baseQuery = baseQuery.Where("bank LIKE ?", strings.Replace(bank, "*", "%", -1))
		} else if strings.Contains(bank, "?") {
			baseQuery = baseQuery.Where("bank LIKE ?", strings.Replace(bank, "?", "_", -1))
		} else {
			baseQuery = baseQuery.Where("bank = ?", bank)
		}
		log.Debug("Applied bank filter", zap.String("bank", bank))
	}
	if key != "" {
		if strings.Contains(key, "*") {
			baseQuery = baseQuery.Where("psql_key LIKE ?", strings.Replace(key, "*", "%", -1))
		} else if strings.Contains(key, "?") {
			baseQuery = baseQuery.Where("psql_key LIKE ?", strings.Replace(key, "?", "_", -1))
		} else {
			baseQuery = baseQuery.Where("psql_key = ?", key)
		}
		log.Debug("Applied key filter", zap.String("key", key))
	}

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			log.Debug("Invalid since date format", zap.String("since", since), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format")
			return
		}
		baseQuery = baseQuery.Where("alter_time >= ?", fromTime)
		log.Debug("Applied since filter", zap.Time("fromTime", fromTime))
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			log.Debug("Invalid until date format", zap.String("until", until), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format")
			return
		}
		baseQuery = baseQuery.Where("alter_time <= ?", toTime)
		log.Debug("Applied until filter", zap.Time("toTime", toTime))
	}

	if jsonpathFilter != "" {
		jsonPathFilters := strings.Split(jsonpathFilter, ",")
		jsonPathFilterQuery, err := jsonPathFilter.BuildJSONPathWhere(jsonPathFilters, "data")
		if err != nil {
			log.Debug("Invalid jsonpath_filter value", zap.String("jsonpath_filter", jsonpathFilter), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "Invalid 'jsonpath_filter' value")
			return
		}
		expr := gorm.Expr(jsonPathFilterQuery)
		log.Debug("Applied jsonpath_filter", zap.String("expr", expr.SQL))
		baseQuery = baseQuery.Where(expr)
	}

	var validColumns = []string{
		"bank",
		"psql_key",
		"id",
		"alter_time",
	}
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "", []string{})
	if err != nil {
		log.Debug("Invalid order_by value", zap.String("order_by", orderBy), zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}
	if validatedOrderBy != "" {
		baseQuery = baseQuery.Order(validatedOrderBy)
		log.Debug("Applied order_by", zap.String("order_by", validatedOrderBy))
	} else {
		baseQuery = baseQuery.Order("bank")
	}

	// Get total count for pagination metadata before applying pagination
	var totalCount int64
	baseQuery.Count(&totalCount)

	// Setup pagination for the query
	paginatedQuery := baseQuery.Offset((page - 1) * perPage).Limit(perPage)

	// Execute the query to find results
	paginatedQuery.Find(&saltCaches)
	log.Debug("Query executed successfully", zap.Int("results_count", len(saltCaches)))

	// Generate pagination links
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	path := c.Request.URL.Path
	baseURL := fmt.Sprintf("%s://%s%s", scheme, host, path)
	nextPage, previousPage := "", ""
	if page > 1 {
		previousPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page-1, perPage)
	}
	if int64((page-1)*perPage+len(saltCaches)) < totalCount {
		nextPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page+1, perPage)
	}

	// Prepare the pagination response
	paging := dto.PageResponse{
		PerPage:  int64(perPage),
		NumPages: int64(math.Ceil(float64(totalCount) / float64(perPage))),
		Count:    totalCount,
		Next:     nextPage,
		Previous: previousPage,
	}

	// Construct and return the structured response
	response := dto.SaltCachePageResponse{
		Results: saltCaches,
		Paging:  paging,
	}

	// Check for empty result set
	if len(saltCaches) == 0 {
		log.Debug("No salt_cache items present")
		httputil.NewError(c, http.StatusNotFound, "No salt_cache items present.")
		return
	}

	log.Debug("Returned salt cache data successfully")
	// Return paginated results
	c.JSON(http.StatusOK, response)
}
