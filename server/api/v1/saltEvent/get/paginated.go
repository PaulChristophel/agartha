package saltEvent

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
)

// GetSaltEvents func get all SaltEvents
//
//	@Summary		Get a list of all salt events (paginated).
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
//	@Param			load_data	query	bool	false	"Load the data field. This defaults to false for performance reasons"
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			per_page	query	int		false	"restrict to X results"
//	@Param			order_by	query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. tag,master_id desc)"
//	@Security		Bearer
func GetSaltEvents(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltEvents []model.SaltEvent

	tag := c.Query("tag")
	masterID := c.Query("master_id")
	loadData := c.Query("load_data")
	since := c.Query("since")
	until := c.Query("until")

	log.Debug("Received request to get salt events", zap.String("tag", tag), zap.String("master_id", masterID), zap.String("load_data", loadData), zap.String("since", since), zap.String("until", until))

	boolValue, err := strconv.ParseBool(loadData)
	if err != nil {
		boolValue = false
		log.Debug("Invalid load_data value, defaulting to false", zap.Error(err))
	} else {
		log.Debug("Parsed load_data successfully", zap.Bool("load_data", boolValue))
	}

	selection := []string{"id", "tag", "alter_time", "master_id"}
	if boolValue {
		selection = append(selection, "data")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	if limit > 1000 {
		limit = 1000
		log.Debug("Limit exceeds maximum, setting to 1000")
	}
	if boolValue && limit > 10 {
		limit = 10
		log.Debug("Limit exceeds maximum for detailed data, setting to 10")
	}

	filterQuery := db.Select(selection).Model(&model.SaltEvent{})
	if tag != "" {
		if strings.Contains(tag, "*") {
			filterQuery = filterQuery.Where("tag LIKE ?", strings.Replace(tag, "*", "%", -1))
		} else if strings.Contains(tag, "?") {
			filterQuery = filterQuery.Where("tag LIKE ?", strings.Replace(tag, "?", "_", -1))
		} else {
			filterQuery = filterQuery.Where("tag = ?", tag)
		}
		log.Debug("Applied tag filter", zap.String("tag", tag))
	}
	if masterID != "" {
		masterID = strings.Replace(masterID, "_", "\\_", -1)
		if strings.Contains(masterID, "*") {
			filterQuery = filterQuery.Where("master_id LIKE ?", strings.Replace(masterID, "*", "%", -1))
		} else if strings.Contains(masterID, "?") {
			filterQuery = filterQuery.Where("master_id LIKE ?", strings.Replace(masterID, "?", "_", -1))
		} else {
			filterQuery = filterQuery.Where("master_id = ?", masterID)
		}
		log.Debug("Applied master_id filter", zap.String("master_id", masterID))
	}

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			log.Debug("Invalid since date format", zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time >= ?", fromTime)
	} else {
		defaultSince := time.Now().Add(-24 * time.Hour * 7)
		filterQuery = filterQuery.Where("alter_time >= ?", defaultSince)
		log.Debug("Applied default since date", zap.Time("default_since", defaultSince))
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			log.Debug("Invalid until date format", zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'until' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time <= ?", toTime)
		log.Debug("Applied until date", zap.Time("until", toTime))
	}

	validColumns := []string{
		"id",
		"tag",
		"alter_time",
		"master_id",
	}
	orderBy := c.Query("order_by")
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "", []string{})
	if err != nil {
		log.Debug("Invalid order_by value", zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}
	if validatedOrderBy != "" {
		filterQuery = filterQuery.Order(validatedOrderBy)
	} else {
		filterQuery.Order("id desc")
	}

	var totalCount int64
	filterQuery.Count(&totalCount)

	resultsQuery := filterQuery.Offset((page - 1) * limit).Limit(limit)
	resultsQuery.Find(&saltEvents)

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

	if len(saltEvents) == 0 {
		log.Debug("No salt_events found")
		httputil.NewError(c, http.StatusNotFound, "No salt_events present.")
		return
	}

	log.Debug("Returning salt_events", zap.Int("count", len(saltEvents)))
	c.JSON(http.StatusOK, response)
}
