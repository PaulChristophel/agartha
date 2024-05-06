package saltReturn

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_return")

	grp.GET("", GetSaltReturns)
	grp.GET("/:jid", GetSaltReturn)
}

// GetSaltReturns func get all SaltReturns
//
//	@Description	Get all SaltReturns
//	@Tags			SaltReturn
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltReturnPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_return [get]
//	@Param			minion_id	query	string	false	"Filter items by minion id (Supports wildcards * and ? for single char matches.)"
//	@Param			fun			query	string	false	"Filter items by function (Supports wildcards * and ? for single char matches.)"
//	@Param			success		query	string	false	"Filter items by success status (true/false). This can be null so it's a string in the database."
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			per_page	query	int		false	"restrict to X results"
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Security		Bearer
func GetSaltReturns(c *gin.Context) {
	db := db.DB
	var saltReturns []model.SaltReturn

	id := c.Query("minion_id")
	fun := c.Query("fun")
	success := c.Query("success")

	since := c.Query("since")
	until := c.Query("until")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if limit > 50 {
		limit = 50
	}

	// Construct the base query with filters
	filterQuery := db.Model(&model.SaltReturn{})
	if id != "" {
		if strings.Contains(id, "*") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.Replace(id, "*", "%", -1))
		} else if strings.Contains(id, "?") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.Replace(id, "?", "_", -1))
		} else {
			filterQuery = filterQuery.Where("id = ?", id)
		}
	}
	if fun != "" {
		if strings.Contains(fun, "*") {
			filterQuery = filterQuery.Where("fun LIKE ?", strings.Replace(fun, "*", "%", -1))
		} else if strings.Contains(fun, "?") {
			filterQuery = filterQuery.Where("fun LIKE ?", strings.Replace(fun, "?", "_", -1))
		} else {
			filterQuery = filterQuery.Where("fun = ?", fun)
		}
	}
	if success != "" {
		filterQuery = filterQuery.Where("success = ?", success)
	}

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time >= ?", fromTime)
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
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

	response := dto.SaltReturnPageResponse{
		Results: saltReturns,
		Paging:  paging,
	}

	// If no results are present, return an error
	if len(saltReturns) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_returns present.")
		return
	}

	// Return the paginated salt returns
	c.JSON(http.StatusOK, response)
}

// GetSaltReturn func get all SaltReturns for a specific jid
//
//	@Description	Get all SaltReturns for a specific jid
//	@Tags			SaltReturn
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.SaltReturn
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_return/{jid} [get]
//	@Param			jid	path	string	true	"jid of the salt return item to retrieve"
//	@Security		Bearer
func GetSaltReturn(c *gin.Context) {
	db := db.DB
	var saltReturns []model.SaltReturn

	jid := c.Param("jid")

	// find all saltReturns in the database with the specified jid
	db.Where("jid = ?", jid).Find(&saltReturns)

	// If no saltReturn is present return an error
	if len(saltReturns) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_returns present.")
		return
	}

	// Else return saltReturns
	c.JSON(http.StatusOK, saltReturns)
}
