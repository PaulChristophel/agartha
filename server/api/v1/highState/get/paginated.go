package highState

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
	model "github.com/PaulChristophel/agartha/server/model/salt/view"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetHighStates retrieves paginated high state items based on the provided query parameters.
//
//	@Summary		Get high state data (paginated).
//	@Description	Get paginated high state items with count and navigation links.
//	@Tags			HighState
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.HighStatePageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/api/v1/high_state [get]
//	@Param			id			query	string	false	"Filter items by minion id (Supports wildcards * and ? for single char matches.)"
//	@Param			success		query	boolean	false	"Filter items by success status (true/false). This can be null so it's a string in the database."
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			per_page	query	int		false	"Number of items per page"
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			order_by	query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. id,true_count desc,success asc)"
//	@Security		Bearer
func GetHighStates(c *gin.Context) {
	db := db.DB
	log := logger.GetLogger()
	var highStates []model.HighState

	id := c.Query("id")
	success := c.Query("success")

	loadReturn := c.Query("load_return")
	loadFullRet := c.Query("load_full_ret")

	since := c.Query("since")
	until := c.Query("until")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if limit > 1000 {
		limit = 1000
	}

	// Parse bool for loading data
	boolLoadReturn, err := strconv.ParseBool(loadReturn)
	if err != nil {
		log.Debug("Invalid load_return value, defaulting to false", zap.String("load_return", loadReturn))
		boolLoadReturn = false
	}
	boolFullRet, err := strconv.ParseBool(loadFullRet)
	if err != nil {
		log.Debug("Invalid load_full_ret value, defaulting to false", zap.String("load_full_ret", loadFullRet))
		boolFullRet = false
	}

	if (boolLoadReturn || boolFullRet) && limit > 10 {
		log.Debug("Limiting per_page to 10 due to load_return or load_full_ret being true", zap.Int("per_page", limit))
		limit = 10
	}

	// Define selection fields
	selection := []string{"fun", "jid", "id", "success", "alter_time"}
	if boolLoadReturn {
		selection = append(selection, "return")
	}
	if boolFullRet {
		selection = append(selection, "full_ret")
	}

	// Construct the base query with filters
	filterQuery := db.Select(selection).Model(&model.HighState{})

	if id != "" {
		if strings.Contains(id, "*") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.Replace(id, "*", "%", -1))
		} else if strings.Contains(id, "?") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.Replace(id, "?", "_", -1))
		} else {
			filterQuery = filterQuery.Where("id = ?", id)
		}
	}

	if success != "" {
		filterQuery = filterQuery.Where("success = ?", success)
	}

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			log.Error("Invalid 'since' date format", zap.String("since", since), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time >= ?", fromTime)
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			log.Error("Invalid 'until' date format", zap.String("until", until), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time <= ?", toTime)
	}

	// List of valid columns that can be sorted
	var validColumns = []string{
		"fun",
		"id",
		"success",
		"alter_time",
	}
	orderBy := c.Query("order_by")
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "", []string{})
	if err != nil {
		log.Error("Invalid order_by parameter", zap.String("order_by", orderBy), zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}

	if validatedOrderBy != "" {
		filterQuery = filterQuery.Order(validatedOrderBy)
	}

	// First, get the total count for pagination metadata
	var totalCount int64
	filterQuery.Count(&totalCount)

	// Apply pagination to the query for fetching results
	resultsQuery := filterQuery.Offset((page - 1) * limit).Limit(limit)
	resultsQuery.Find(&highStates)

	// Construct pagination URLs
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	path := c.Request.URL.Path
	baseURL := fmt.Sprintf("%s://%s%s", scheme, host, path)

	var nextPage, previousPage string
	if page > 1 {
		previousPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page-1, limit)
	}
	if int64((page-1)*limit+len(highStates)) < totalCount {
		nextPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page+1, limit)
	}

	// Prepare the pagination response
	paging := dto.PageResponse{
		PerPage:  int64(limit),
		NumPages: int64(math.Ceil(float64(totalCount) / float64(limit))),
		Count:    totalCount,
		Next:     nextPage,
		Previous: previousPage,
	}

	response := dto.HighStatePageResponse{
		Results: highStates,
		Paging:  paging,
	}

	// If no results are present, return an error
	if len(highStates) == 0 {
		log.Debug("No high_states present", zap.String("id", id), zap.String("success", success), zap.String("since", since), zap.String("until", until))
		httputil.NewError(c, http.StatusNotFound, "No high_states present.")
		return
	}

	// Log the successful retrieval of the high states
	log.Debug("Successfully retrieved high states",
		zap.String("id", id),
		zap.String("success", success),
		zap.String("since", since),
		zap.String("until", until),
		zap.Int("page", page),
		zap.Int("per_page", limit),
		zap.Int("result_count", len(highStates)),
		zap.Int64("total_count", totalCount))

	// Return the paginated high states
	c.JSON(http.StatusOK, response)
}
