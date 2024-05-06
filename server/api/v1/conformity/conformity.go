package conformity

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
	grp := rg.Group("/conformity")

	grp.GET("/", GetConformities)
	grp.GET("/:id", GetConformity)
}

// GetConformities retrieves paginated conformity items based on the provided limit and page query parameters.
//
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
//	@Param			limit	query	int	false	"Number of items per page"
//	@Param			page	query	int	false	"Page number of results to retrieve"
//	@Security		Bearer
func GetConformities(c *gin.Context) {
	db := db.DB
	var conformities []model.Conformity

	// Read query parameters for pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if limit > 50 {
		limit = 50 // Ensure limit does not exceed 50 items per page for performance reasons
	}

	// Initialize base query to use throughout
	baseQuery := db.Model(&model.Conformity{})

	// First, get the total count for pagination metadata
	var totalCount int64
	baseQuery.Count(&totalCount)

	// Calculate the offset for pagination
	offset := (page - 1) * limit

	// Fetch the conformities with limit and offset
	paginatedQuery := baseQuery.Offset(offset).Limit(limit)
	paginatedQuery.Find(&conformities)

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
		httputil.NewError(c, http.StatusNotFound, "No conformities present.")
		return
	}

	// Return paginated conformities
	c.JSON(http.StatusOK, response)
}

// GetConformity func get all GetConformity
//
//	@Description	Get all GetConformity
//	@Tags			Conformity
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.Conformity
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/conformity/{id} [get]
//	@Param			id	path	string	false	"conformity object to get"
//	@Security		Bearer
func GetConformity(c *gin.Context) {
	db := db.DB
	var conformity []model.Conformity

	id := c.Param("id")

	// find all jids in the database
	db.Where("id = ?", id).Find(&conformity)

	// If no jid is present return an error
	if len(conformity) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No conformity item present.")
		return
	}

	// Else return jids
	c.JSON(http.StatusOK, conformity)
}
