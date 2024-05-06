package highState

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt/materializedView"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/high_state")

	grp.GET("", GetHighStates)
	grp.GET("/:id", GetHighStateID)
	grp.GET("/jid/:jid", GetHighState)
}

// GetHighStates func get all HighStates
//
//	@Description	Get all HighStates
//	@Tags			HighState
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.HighStatePageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/high_state [get]
//	@Param			page		query	int	false	"Page number of results to retrieve"
//	@Param			per_page	query	int	false	"restrict to X results"
//	@Security		Bearer
func GetHighStates(c *gin.Context) {
	db := db.DB
	var HighStates []model.HighState

	// Read query parameter for pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	if limit > 50 {
		limit = 50
	}

	// Initialize base query to use throughout
	baseQuery := db.Model(&model.HighState{})

	// First, get the total count for pagination metadata
	var totalCount int64
	baseQuery.Count(&totalCount)

	// Fetch the HighStates with limit and offset
	paginatedQuery := baseQuery.Offset((page - 1) * limit).Limit(limit)
	paginatedQuery.Find(&HighStates)

	// Construct pagination URLs
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	path := c.Request.URL.Path
	baseURL := fmt.Sprintf("%s://%s%s", scheme, host, path)
	nextPage, previousPage := "", ""
	if page > 1 {
		previousPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page-1, limit)
	}
	if int64((page-1)*limit+len(HighStates)) < totalCount {
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

	// Prepare the pagination response
	response := dto.HighStatePageResponse{
		Results: HighStates,
		Paging:  paging,
	}

	// If no HighState is present return an error
	if len(HighStates) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No high_states present.")
		return
	}

	// Else return HighStates with pagination
	c.JSON(http.StatusOK, response)
}

// GetHighState func get all HighStates for a specific jid
//
//	@Description	Get all HighStates for a specific jid
//	@Tags			HighState
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.HighState
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/high_state/jid/{jid} [get]
//	@Param			jid	path	string	true	"jid of the salt return item to retrieve"
//	@Security		Bearer
func GetHighState(c *gin.Context) {
	db := db.DB
	var HighStates []model.HighState

	jid := c.Param("jid")

	// find all HighStates in the database with the specified jid
	db.Where("jid = ?", jid).Find(&HighStates)

	// If no HighState is present return an error
	if len(HighStates) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No high_states present.")
		return
	}

	// Else return HighStates
	c.JSON(http.StatusOK, HighStates)
}

// GetHighState func get all HighStates for a specific id
//
//	@Description	Get all HighStates for a specific id
//	@Tags			HighState
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		model.HighState
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/high_state/{id} [get]
//	@Param			id	path	string	true	"id of the salt return item to retrieve"
//	@Security		Bearer
func GetHighStateID(c *gin.Context) {
	db := db.DB
	var HighStates []model.HighState

	id := c.Param("id")

	// find all HighStates in the database with the specified id
	db.Where("id = ?", id).Find(&HighStates)

	// If no HighState is present return an error
	if len(HighStates) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No high_states present.")
		return
	}

	// Else return HighStates
	c.JSON(http.StatusOK, HighStates)
}
