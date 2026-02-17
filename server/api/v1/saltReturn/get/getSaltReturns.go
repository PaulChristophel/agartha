package saltReturn

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

// GetSaltReturns func get all SaltReturns
//
//	@Summary		Get all SaltReturns (paginated)
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
//	@Param			id				query	string	false	"Filter items by minion id (Supports wildcards * and ? for single char matches.)"
//	@Param			jid				query	string	false	"Filter items by jid (Supports wildcards * and ? for single char matches.)"
//	@Param			fun				query	string	false	"Filter items by function (Supports wildcards * and ? for single char matches.)"
//	@Param			success			query	bool	false	"Filter items by success status (true/false)."
//	@Param			load_return		query	bool	false	"Load the return field. This defaults to false for performance reasons"
//	@Param			load_full_ret	query	bool	false	"Load the full_ret field. This defaults to false for performance reasons"
//	@Param			since			query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until			query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			per_page		query	int		false	"restrict to X results"
//	@Param			page			query	int		false	"Page number of results to retrieve"
//	@Param			order_by		query	string	false	"Order by column(s). Comma separated list of columns to order by (e.g. id,fun desc)"
//	@Security		Bearer
func GetSaltReturns(c *gin.Context) {
	db := db.DB.Table(table)
	log := logger.GetLogger()
	var saltReturns []model.SaltReturn

	id := c.Query("id")
	jid := c.Query("jid")
	fun := c.Query("fun")
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
	orderBy := c.Query("order_by")

	log.Debug("Received request to get salt returns",
		zap.String("id", id),
		zap.String("jid", jid),
		zap.String("fun", fun),
		zap.String("success", success),
		zap.String("load_return", loadReturn),
		zap.String("load_full_ret", loadFullRet),
		zap.String("since", since),
		zap.String("until", until),
		zap.Int("page", page),
		zap.Int("per_page", limit),
		zap.String("order_by", orderBy),
	)

	// Parse bool for loading data
	boolLoadReturn, err := strconv.ParseBool(loadReturn)
	if err != nil {
		boolLoadReturn = false
		log.Debug("Invalid load_return value, defaulting to false", zap.String("load_return", loadReturn), zap.Error(err))
	}
	boolFullRet, err := strconv.ParseBool(loadFullRet)
	if err != nil {
		boolFullRet = false
		log.Debug("Invalid load_full_ret value, defaulting to false", zap.String("load_full_ret", loadFullRet), zap.Error(err))
	}

	if (boolLoadReturn || boolFullRet) && limit > 10 {
		limit = 10
		log.Debug("PerPage exceeds maximum for detailed data, setting to 10")
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
	filterQuery := db.Select(selection).Model(&model.SaltReturn{})

	if id != "" {
		if strings.Contains(id, "*") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.ReplaceAll(id, "*", "%"))
		} else if strings.Contains(id, "?") {
			filterQuery = filterQuery.Where("id LIKE ?", strings.ReplaceAll(id, "?", "_"))
		} else {
			filterQuery = filterQuery.Where("id = ?", id)
		}
	}
	if jid != "" {
		if strings.Contains(jid, "*") {
			filterQuery = filterQuery.Where("jid LIKE ?", strings.ReplaceAll(jid, "*", "%"))
		} else if strings.Contains(jid, "?") {
			filterQuery = filterQuery.Where("jid LIKE ?", strings.ReplaceAll(jid, "?", "_"))
		} else {
			filterQuery = filterQuery.Where("jid = ?", jid)
		}
	}
	if fun != "" {
		funs := strings.Split(fun, ",")
		orConditions := make([]string, 0, len(funs))
		args := make([]any, 0, len(funs))

		for _, element := range funs {
			if strings.Contains(element, "*") {
				orConditions = append(orConditions, "fun LIKE ?")
				args = append(args, strings.ReplaceAll(element, "*", "%"))
			} else if strings.Contains(element, "?") {
				orConditions = append(orConditions, "fun LIKE ?")
				args = append(args, strings.ReplaceAll(element, "?", "_"))
			} else {
				orConditions = append(orConditions, "fun = ?")
				args = append(args, element)
			}
		}

		if len(orConditions) > 0 {
			filterQuery = filterQuery.Where(strings.Join(orConditions, " OR "), args...)
		}
	}
	if success != "" {
		filterQuery = filterQuery.Where("success = ?", success)
	}

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			log.Debug("Invalid 'since' date format", zap.String("since", since), zap.Error(err))
			httputil.NewError(c, http.StatusBadRequest, "invalid 'since' date format.")
			return
		}
		filterQuery = filterQuery.Where("alter_time >= ?", fromTime)
	} else if jid == "" {
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
	var validColumns = []string{
		"fun",
		"jid",
		"id",
		"success",
		"alter_time",
	}
	validatedOrderBy, err := validate.OrderBy(orderBy, validColumns, "", []string{})
	if err != nil {
		log.Debug("Invalid order_by value", zap.String("order_by", orderBy), zap.Error(err))
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}

	if validatedOrderBy != "" {
		filterQuery = filterQuery.Order(validatedOrderBy)
	} else {
		filterQuery = filterQuery.Order("alter_time desc")
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

	response := dto.SaltReturnPageResponse{
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
