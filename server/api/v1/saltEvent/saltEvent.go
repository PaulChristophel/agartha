package saltEvent

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
	grp := rg.Group("/salt_event")

	grp.GET("/", GetSaltEvents)
	grp.GET("/:id", GetSaltEvent)
}

// GetSaltEvents func get all SaltEvents
//
//	@Description	Get all SaltEvents
//	@Tags			SaltEvent
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltEventPageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_event [get]
//	@Param			tag			query	string	false	"tag of the event sent to the master (Supports wildcards * and ? for single char matches.)"
//	@Param			master_id	query	string	false	"id of the master that received the event"
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			per_page	query	int		false	"restrict to X results"
//	@Security		Bearer
func GetSaltEvents(c *gin.Context) {
	db := db.DB
	var saltEvents []model.SaltEvent
	tag := c.Query("tag")
	masterID := c.Query("master_id")
	since := c.Query("since")
	until := c.Query("until")

	// Read query parameters for pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	// Construct the base query with filters
	filterQuery := db.Model(&model.SaltEvent{})
	if tag != "" {
		if strings.Contains(tag, "*") {
			filterQuery = filterQuery.Where("tag LIKE ?", strings.Replace(tag, "*", "%", -1))
		} else if strings.Contains(tag, "?") {
			filterQuery = filterQuery.Where("tag LIKE ?", strings.Replace(tag, "?", "_", -1))
		} else {
			filterQuery = filterQuery.Where("id = ?", tag)
		}
	}
	if masterID != "" {
		filterQuery = filterQuery.Where("master_id = ?", masterID)
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
	resultsQuery.Find(&saltEvents)

	// URL base for pagination links
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
	if int64((page-1)*limit+len(saltEvents)) < totalCount {
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

	response := dto.SaltEventPageResponse{
		Results: saltEvents,
		Paging:  paging,
	}

	// If no event is present, return an error
	if len(saltEvents) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_events present.")
		return
	}

	// Return saltEvents with pagination
	c.JSON(http.StatusOK, response)
}

// GetSaltEvents func get all SaltEvents
//
//	@Description	Get all SaltEvents
//	@Tags			SaltEvent
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltEvent
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_event/{id} [get]
//	@Param			id	path	int	false	"return salt event with given id"
//	@Security		Bearer
func GetSaltEvent(c *gin.Context) {
	db := db.DB
	var saltEvent model.SaltEvent

	// Read the param jid
	id := c.Param("id")

	// Find the jid with the given Id
	idInt64, _ := strconv.ParseInt(id, 10, 32)
	db.Find(&saltEvent, "id = ?", idInt64)

	// If no event is present throw an error
	if saltEvent.ID <= 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_events present.")
		return
	}

	// Else return events
	c.JSON(http.StatusOK, saltEvent)
}
