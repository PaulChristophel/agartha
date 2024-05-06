package jid

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/jid")

	grp.GET("", GetJIDs)
	grp.GET("/:jid", GetJID)
}

// GetJIDs retrieves paginated jids based on the provided limit and page query parameters.
//
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
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			per_page	query	int		false	"Number of items per page"
//	@Security		Bearer
func GetJIDs(c *gin.Context) {
	db := db.DB
	var jids []model.JID

	// Initialize base query to use throughout
	baseQuery := db.Model(&model.JID{})

	// Read and validate 'since' and 'until' query parameters
	since := c.Query("since")
	until := c.Query("until")
	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		baseQuery = baseQuery.Where("alter_time >= ?", fromTime)
	}

	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format.")
			return
		}
		baseQuery = baseQuery.Where("alter_time <= ?", toTime)
	}

	// Parse pagination query parameters with default values
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	if perPage > 50 {
		perPage = 50
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
		httputil.NewError(c, http.StatusNotFound, "No jids present.")
		return
	}

	// Return the paginated jids
	c.JSON(http.StatusOK, response)
}

// GetJID func one jid by JID
//
//	@Description	Get one jid by JID
//	@Tags			JID
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.JID
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/jid/{jid} [get]
//	@Param			jid	path	string	false	"jid to return"
//	@Security		Bearer
func GetJID(c *gin.Context) {
	db := db.DB
	var jid model.JID

	// Read the param jid
	id := c.Param("jid")

	// Find the jid with the given Id
	db.Find(&jid, "jid = ?", id)

	// If no such jid present return an error
	if jid.JID == "" {
		httputil.NewError(c, http.StatusNotFound, "No jid present.")
		return
	}

	// Return the jid with the Id
	c.JSON(http.StatusOK, jid)
}
