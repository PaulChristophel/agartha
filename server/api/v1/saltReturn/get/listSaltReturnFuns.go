package saltReturn

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListSaltReturnFuns func get all SaltReturns
//
//	@Summary		List all unique instances of commands used (paginated)
//	@Description	List all unique instances of commands used
//	@Tags			SaltReturn
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltReturnFunPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_return/fun [get]
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			per_page	query	int		false	"restrict to X results"
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Security		Bearer
func ListSaltReturnFuns(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltReturns []string

	since := c.Query("since")
	until := c.Query("until")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if limit > 1000 {
		limit = 1000
	}

	log.Debug("Received request to get salt returns",
		zap.String("since", since),
		zap.String("until", until),
		zap.Int("page", page),
		zap.Int("per_page", limit))

	// Define selection fields
	selection := []string{"fun"}

	// Construct the base query with filters
	filterQuery := db.Distinct(selection)

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			log.Debug("Invalid 'since' date format", zap.String("since", since), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time >= ?", fromTime)
	} else {
		// Default to the last 24*7 hours if 'since' is not specified
		defaultSince := time.Now().Add(-24 * time.Hour * 7)
		filterQuery = filterQuery.Where("alter_time >= ?", defaultSince)
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			log.Debug("Invalid 'until' date format", zap.String("until", until), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time <= ?", toTime)
	}

	// First, get the total count for pagination metadata
	var totalCount int64
	filterQuery.Count(&totalCount)

	// Apply pagination to the query for fetching results
	resultsQuery := filterQuery.Offset((page - 1) * limit).Limit(limit)
	resultsQuery.Find(&saltReturns)

	log.Debug("Query executed successfully", zap.Int("results_count", len(saltReturns)))

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
	if int64((page-1)*limit+len(saltReturns)) < totalCount {
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

	response := dto.SaltReturnFunPageResponse{
		Results: saltReturns,
		Paging:  paging,
	}

	// If no results are present, return an error
	if len(saltReturns) == 0 {
		log.Debug("No salt_returns present")
		httputil.NewError(c, http.StatusNotFound, "No salt_returns present.")
		return
	}

	// Return the paginated salt returns
	c.JSON(http.StatusOK, response)
	log.Debug("Returned salt return data successfully")
}
