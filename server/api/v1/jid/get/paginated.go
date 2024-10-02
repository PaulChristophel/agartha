package jid

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PaulChristophel/agartha/server/api/validate"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetJIDs retrieves paginated jids based on the provided limit and page query parameters.
//
//	@Summary		Get Job data about all jobs (paginated).
//	@Description	Get paginated jids with count and navigation links.
//	@Tags			JID
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.JIDPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/jid [get]
//	@Param			jid			query	string	false	"Filter based on JIDs starting with input string"
//	@Param			load_load	query	bool	false	"Load the Load field. This defaults to false for performance reasons"
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			per_page	query	int		false	"Number of items per page"
//	@Param			order_by	query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. jid,alter_time desc)"
//	@Security		Bearer
func GetJIDs(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var jids []model.JID

	filter := c.Query("jid")

	if filter == "*" || filter == "?" {
		filter = ""
	}

	loadLoad := c.Query("load_load")

	boolValue, err := strconv.ParseBool(loadLoad)
	if err != nil {
		log.Debug("Invalid load_load value, defaulting to false", zap.String("load_load", loadLoad))
		boolValue = false // Default or log the error
	}

	// Define selection fields
	selection := []string{"jid", "alter_time"}
	if boolValue {
		selection = append(selection, "load")
	}

	// Initialize base query to use throughout
	baseQuery := db.Select(selection).Model(&model.JID{})

	if filter != "" {
		if strings.Contains(filter, "*") {
			baseQuery = baseQuery.Where("jid LIKE ?", strings.Replace(filter, "*", "%", -1))
		} else if strings.Contains(filter, "?") {
			baseQuery = baseQuery.Where("jid LIKE ?", strings.Replace(filter, "?", "_", -1))
		} else {
			baseQuery = baseQuery.Where("jid = ?", filter)
		}
	}

	// Read and validate 'since' and 'until' query parameters
	since := c.Query("since")
	until := c.Query("until")

	// Check filter length
	filterLen := utf8.RuneCountInString(filter)

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			log.Error("Invalid 'since' date format", zap.String("since", since), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		baseQuery = baseQuery.Where("alter_time >= ?", fromTime)
	} else if filterLen <= 6 {
		// Default to the last 24*7 hours if 'since' is not specified and the filter is too short
		defaultSince := time.Now().Add(-24 * time.Hour * 7)
		baseQuery = baseQuery.Where("alter_time >= ?", defaultSince)
	}

	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			log.Error("Invalid 'until' date format", zap.String("until", until), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format.")
			return
		}
		baseQuery = baseQuery.Where("alter_time <= ?", toTime)
	}

	validColumns := []string{"jid", "alter_time", "load"}
	orderBy := c.Query("order_by")
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "", []string{})
	if err != nil {
		log.Error("Invalid order_by parameter", zap.String("order_by", orderBy), zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}
	if validatedOrderBy != "" {
		baseQuery = baseQuery.Order(validatedOrderBy)
	} else {
		baseQuery = baseQuery.Order("alter_time desc")
	}
	// Parse pagination query parameters with default values
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	if perPage > 1000 {
		perPage = 1000
	}
	if boolValue && perPage > 10 {
		log.Info("Limiting per_page to 10 due to load_load being true", zap.Int("per_page", perPage))
		perPage = 10
	}

	// First, get the total count for pagination metadata
	var totalCount int64
	baseQuery.Count(&totalCount)

	// Fetch the records with pagination
	paginatedQuery := baseQuery.Offset((page - 1) * perPage).Limit(perPage)
	paginatedQuery.Find(&jids)

	// Generate URLs for next and previous pages
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
	if int64((page-1)*perPage+len(jids)) < totalCount {
		nextPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page+1, perPage)
	}

	// Construct the pagination response
	paging := dto.PageResponse{
		PerPage:  int64(perPage),
		NumPages: int64(math.Ceil(float64(totalCount) / float64(perPage))),
		Count:    totalCount,
		Next:     nextPage,
		Previous: previousPage,
	}

	response := dto.JIDPageResponse{
		Results: jids,
		Paging:  paging,
	}

	// Handle the case when no jids are found
	if len(jids) == 0 {
		log.Debug("No jids present", zap.String("filter", filter), zap.String("since", since), zap.String("until", until))
		httputil.NewError(c, http.StatusNotFound, "No jids present.")
		return
	}

	// Log the successful retrieval of the jids
	log.Debug("Successfully retrieved jids",
		zap.String("filter", filter),
		zap.String("since", since),
		zap.String("until", until),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
		zap.Int("result_count", len(jids)),
		zap.Int64("total_count", totalCount))

	// Return the paginated jids
	c.JSON(http.StatusOK, response)
}
