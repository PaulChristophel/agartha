package conformity

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
	matModel "github.com/PaulChristophel/agartha/server/model/salt/materializedView"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetConformities retrieves paginated conformity items based on the provided limit and page query parameters.
//
//	@Summary		Get conformity data about all minions (paginated).
//	@Description	Get paginated conformity items with count and navigation links.
//	@Tags			Conformity
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.ConformityPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/conformity [get]
//	@Param			id			query	string	false	"Filter items by minion id (Supports wildcards * and ? for single char matches.)"
//	@Param			success		query	boolean	false	"Filter items by success status (true/false). This can be null so it's a string in the database."
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			limit		query	int		false	"Number of items per page"
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			order_by	query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. id,true_count desc,success asc)"
//	@Security		Bearer
func GetConformities(c *gin.Context) {
	var log = logger.GetLogger()
	var conformities []matModel.Conformity
	db := db.DB

	id := c.Query("id")
	success := c.Query("success")

	since := c.Query("since")
	until := c.Query("until")

	// Read query parameters for pagination
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		log.Error("invalid page parameter", zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, "invalid page parameter.")
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if err != nil {
		log.Error("invalid limit parameter", zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, "invalid limit parameter.")
		return
	}

	if limit > 1000 {
		limit = 1000
	}

	// Initialize base query to use throughout
	filterQuery := db.Model(&matModel.Conformity{})

	if id != "" {
		if strings.Contains(id, "*") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.ReplaceAll(id, "*", "%"))
		} else if strings.Contains(id, "?") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.ReplaceAll(id, "?", "_"))
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
			log.Error("invalid 'since' date format", zap.String("since", since), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time >= ?", fromTime)
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			log.Error("invalid 'until' date format", zap.String("until", until), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time <= ?", toTime)
	}

	var validColumns = []string{
		"id",
		"alter_time",
		"success",
		"true_count",
		"false_count",
		"changed_count",
		"unchanged_count",
	}
	orderBy := c.Query("order_by")
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "", []string{})
	if err != nil {
		log.Error("invalid order_by parameter", zap.String("order_by", orderBy), zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}

	if validatedOrderBy != "" {
		filterQuery = filterQuery.Order(validatedOrderBy)
	}

	// First, get the total count for pagination metadata
	var totalCount int64
	filterQuery.Count(&totalCount)

	// Calculate the offset for pagination
	offset := (page - 1) * limit

	// Fetch the conformities with limit and offset
	paginatedQuery := filterQuery.Offset(offset).Limit(limit)
	paginatedQuery.Find(&conformities)

	log.Info("conformities fetched", zap.Int("count", len(conformities)), zap.Int64("total", totalCount))

	// Determine next and previous page URLs
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	path := c.Request.URL.Path
	baseURL := fmt.Sprintf("%s://%s%s", scheme, host, path)
	nextPage, previousPage := "", ""
	if page > 1 {
		previousPage = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page-1, limit)
	}
	if int64((page-1)*limit+len(conformities)) < totalCount {
		nextPage = fmt.Sprintf("%s?page=%d&limit=%d", baseURL, page+1, limit)
	}

	// Prepare the pagination response
	paging := dto.PageResponse{
		PerPage:  int64(limit),
		NumPages: int64(math.Ceil(float64(totalCount) / float64(limit))),
		Count:    totalCount,
		Next:     nextPage,
		Previous: previousPage,
	}

	// Prepare the pagination response
	response := dto.ConformityPageResponse{
		Results: conformities,
		Paging:  paging,
	}

	// If no conformity is present return an error
	if len(conformities) == 0 {
		log.Debug("No conformities present", zap.String("id", id))
		httputil.NewError(c, http.StatusNotFound, "No conformities present.")
		return
	}

	log.Debug("Returning paginated conformities", zap.Int("page", page), zap.Int("limit", limit))
	// Return paginated conformities
	c.JSON(http.StatusOK, response)
}
