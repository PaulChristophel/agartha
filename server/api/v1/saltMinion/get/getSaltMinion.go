package saltMinion

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
	model "github.com/PaulChristophel/agartha/server/model/salt/view"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GetSaltMinion func retrieves a paginated list of saltMinion items filtered by minionID and key
//
//	@Summary		Retrieve a list of salt_minion items (paginated).
//	@Description	Retrieve a list of salt_minion items filtered by minionID and key with pagination support.
//	@Tags			SaltMinion
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltMinionPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_minion [get]
//	@Param			minion_id				query	string	false	"Minion ID of the cached item to search (Supports wildcards * and ? for single char matches.)"
//	@Param			load_grains				query	bool	false	"Load the entire grains field. This defaults to false for performance reasons. (Implied false if jsonpath_grains != '')"
//	@Param			jsonpath_grains			query	string	false	"Comma separated list of grains items to return as subset of data in jsonpath syntax (e.g. os,kernel,efi)"
//	@Param			jsonpath_grains_filter	query	string	false	"Comma separated list of data items to filter on (e.g. os:RedHat::string,kernel:Linux::string,efi:true::bool)"
//	@Param			load_pillar				query	bool	false	"Load the entire pillar field. This defaults to false for performance reasons. (Implied false if jsonpath_pillar != '')"
//	@Param			jsonpath_pillar			query	string	false	"Comma separated list of grains items to return as subset of data in jsonpath syntax (e.g. os,kernel,efi)"
//	@Param			jsonpath_pillar_filter	query	string	false	"Comma separated list of data items to filter on (e.g. os:RedHat::string,kernel:Linux::string,efi:true::bool)"
//	@Param			since					query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until					query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page					query	int		false	"Page number of results to retrieve"
//	@Param			per_page				query	int		false	"Number of items per page"
//	@Param			order_by				query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. minionID,psql_key desc,alter_time asc,id)"
//	@Security		Bearer
func GetSaltMinion(c *gin.Context) {

	db := db.DB
	log := logger.GetLogger()
	var saltMinions []model.SaltMinion

	// Read query parameters
	minionID := c.Query("minion_id")
	key := c.Query("key")
	loadGrains := c.Query("load_grains")
	jsonpathGrains := c.Query("jsonpath_grains")
	jsonpathGrainsFilter := c.Query("jsonpath_grains_filter")
	loadPillar := c.Query("load_pillar")
	jsonpathPillar := c.Query("jsonpath_pillar")
	jsonpathPillarFilter := c.Query("jsonpath_pillar_filter")
	since := c.Query("since")
	until := c.Query("until")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	orderBy := c.Query("order_by")

	log.Debug("Received request to get salt minion",
		zap.String("minion_id", minionID),
		zap.String("key", key),
		zap.String("load_grains", loadGrains),
		zap.String("jsonpath_grains", jsonpathGrains),
		zap.String("jsonpath_grains_filter", jsonpathGrainsFilter),
		zap.String("load_pillar", loadPillar),
		zap.String("jsonpath_pillar", jsonpathPillar),
		zap.String("jsonpath_pillar_filter", jsonpathPillarFilter),
		zap.String("since", since),
		zap.String("until", until),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
		zap.String("order_by", orderBy))

	// Parse bool for loading data
	boolGrainsValue, err := strconv.ParseBool(loadGrains)
	if err != nil {
		boolGrainsValue = false
		log.Debug("Invalid load_grains value, defaulting to false", zap.String("load_grains", loadGrains), zap.Error(err))
	}
	if jsonpathGrains != "" {
		boolGrainsValue = false
	}

	boolPillarValue, err := strconv.ParseBool(loadPillar)
	if err != nil {
		boolPillarValue = false
		log.Debug("Invalid load_grains value, defaulting to false", zap.String("load_pillar", loadPillar), zap.Error(err))
	}
	if jsonpathPillar != "" {
		boolPillarValue = false
	}

	if perPage > 1000 {
		perPage = 1000
		log.Debug("PerPage exceeds maximum, setting to 1000")
	}
	if boolPillarValue || boolGrainsValue && perPage > 10 {
		perPage = 10
		log.Debug("PerPage exceeds maximum for detailed data, setting to 10")
	}

	// Define selection fields
	selection := []string{"minion_id", "id", "alter_time"}
	if boolGrainsValue {
		selection = append(selection, "grains")
	} else if jsonpathGrains != "" {
		// Apply subData filters if a
		subDataPaths := strings.Split(jsonpathGrains, ",")
		jsonPathQuery := jsonPathFilter.BuildJSONPathSelect(subDataPaths, "grains")
		expr := gorm.Expr(jsonPathQuery)
		selection = append(selection, expr.SQL)
	}
	if boolPillarValue {
		selection = append(selection, "pillar")
	} else if jsonpathPillar != "" {
		// Apply subData filters if a
		subDataPaths := strings.Split(jsonpathPillar, ",")
		jsonPathQuery := jsonPathFilter.BuildJSONPathSelect(subDataPaths, "pillar")
		expr := gorm.Expr(jsonPathQuery)
		selection = append(selection, expr.SQL)
	}

	// Create a base query with selected fields
	baseQuery := db.Select(selection).Model(&model.SaltMinion{})

	// Apply filters based on provided minionID and key
	if minionID != "" {
		if strings.Contains(minionID, "*") {
			baseQuery = baseQuery.Where("minion_id LIKE ?", strings.Replace(minionID, "*", "%", -1))
		} else if strings.Contains(minionID, "?") {
			baseQuery = baseQuery.Where("minion_id LIKE ?", strings.Replace(minionID, "?", "_", -1))
		} else {
			baseQuery = baseQuery.Where("minion_id = ?", minionID)
		}
		log.Debug("Applied minion_id filter", zap.String("minion_id", minionID))
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

	if jsonpathGrainsFilter != "" {
		jsonPathFilters := strings.Split(jsonpathGrainsFilter, ",")
		jsonPathFilterQuery, err := jsonPathFilter.BuildJSONPathWhere(jsonPathFilters, "grains")
		if err != nil {
			log.Debug("Invalid jsonpath_filter value", zap.String("jsonpath_grains_filter", jsonpathGrainsFilter), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "Invalid 'jsonpath_grains_filter' value")
			return
		}
		expr := gorm.Expr(jsonPathFilterQuery)
		log.Debug("Applied jsonpath_grains_filter", zap.String("expr", expr.SQL))
		baseQuery = baseQuery.Where(expr)
	}

	if jsonpathPillarFilter != "" {
		jsonPathFilters := strings.Split(jsonpathPillarFilter, ",")
		jsonPathFilterQuery, err := jsonPathFilter.BuildJSONPathWhere(jsonPathFilters, "pillar")
		if err != nil {
			log.Debug("Invalid jsonpath_filter value", zap.String("jsonpath_pillar_filter", jsonpathPillarFilter), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "Invalid 'jsonpath_pillar_filter' value")
			return
		}
		expr := gorm.Expr(jsonPathFilterQuery)
		log.Debug("Applied jsonpath_pillar_filter", zap.String("expr", expr.SQL))
		baseQuery = baseQuery.Where(expr)
	}

	var validColumns = []string{
		"minion_id",
		"id",
		"alter_time",
	}
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "grains", strings.Split(jsonpathGrains, ","))
	if err != nil {
		log.Debug("Invalid order_by value", zap.String("order_by", orderBy), zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}
	if validatedOrderBy != "" {
		baseQuery = baseQuery.Order(validatedOrderBy)
		log.Debug("Applied order_by", zap.String("order_by", validatedOrderBy))
	} else {
		baseQuery = baseQuery.Order("minion_id")
	}

	// Get total count for pagination metadata before applying pagination
	var totalCount int64
	baseQuery.Count(&totalCount)

	// Setup pagination for the query
	paginatedQuery := baseQuery.Offset((page - 1) * perPage).Limit(perPage)

	// Execute the query to find results
	paginatedQuery.Find(&saltMinions)
	log.Debug("Query executed successfully", zap.Int("results_count", len(saltMinions)))

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
	if int64((page-1)*perPage+len(saltMinions)) < totalCount {
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
	response := dto.SaltMinionPageResponse{
		Results: saltMinions,
		Paging:  paging,
	}

	// Check for empty result set
	if len(saltMinions) == 0 {
		log.Debug("No salt_minion items present")
		httputil.NewError(c, http.StatusNotFound, "No salt_minion items present.")
		return
	}

	log.Debug("Returned salt minion data successfully")
	// Return paginated results
	c.JSON(http.StatusOK, response)
}
