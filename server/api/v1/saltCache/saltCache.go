package saltCache

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	// "strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/salt_cache")

	grp.GET("", GetSaltCache)
	grp.GET("/uuid/:id", GetSaltCacheUUID)
	grp.GET("/:bank/:key", GetSaltCacheBankKey)
	grp.DELETE("/uuid/:id", DeleteSaltCacheUUID)
	grp.DELETE("/:bank/:key", DeleteSaltCacheBankKey)
}

// GetSaltCache func one saltCache by the item's UUID
//
//	@Description	Get one salt_cache item by UUID
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltCache
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/uuid/{uuid} [get]
//	@Param			uuid	path	string	true	"uuid of the salt cache item to retrieve"
//	@Security		Bearer
func GetSaltCacheUUID(c *gin.Context) {
	db := db.DB
	var saltCache model.SaltCache

	// Read the param saltCache
	uuid := c.Param("uuid")

	if uuid == "" {
		httputil.NewError(c, http.StatusNotFound, "uuid required.")
		return
	}

	// Find the saltCache with the given UUID
	db.Where("id = ?", uuid).Find(&saltCache)

	// If no such saltCache is present, return an error
	if saltCache.Bank == "" {
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	// Return the saltCache with the UUID
	c.JSON(http.StatusOK, saltCache)
}

// GetSaltCache func retrieves a paginated list of saltCache items filtered by bank and key
//
//	@Description	Retrieve a list of salt_cache items filtered by bank and key with pagination support.
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltCachePageResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache [get]
//	@Param			bank		query	string	false	"Bank of the cached item to search (Supports wildcards * and ? for single char matches.)"
//	@Param			key			query	string	false	"Key of the cached item to search (Supports wildcards * and ? for single char matches.)"
//	@Param			load_data	query	bool	false	"Load the data field. This defaults to false for performance reasons"
//	@Param			since		query	string	false	"Filter items from this date (RFC3339 format)."
//	@Param			until		query	string	false	"Filter items up to this date (RFC3339 format)."
//	@Param			page		query	int		false	"Page number of results to retrieve"
//	@Param			per_page	query	int		false	"Number of items per page"
//	@Security		Bearer
func GetSaltCache(c *gin.Context) {
	db := db.DB
	var saltCaches []model.SaltCache

	// Read query parameters
	bank := c.Query("bank")
	key := c.Query("key")
	loadData := c.Query("load_data")
	since := c.Query("since")
	until := c.Query("until")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))

	// Parse bool for loading data
	boolValue, _ := strconv.ParseBool(loadData) // Ignore error; default false

	// Define selection fields
	selection := []string{"bank", "psql_key", "id", "data_changed"}
	if boolValue {
		selection = append(selection, "data")
	}

	// Create a base query with selected fields
	baseQuery := db.Select(selection).Model(&model.SaltCache{})

	// Apply filters based on provided bank and key
	if bank != "" {
		if strings.Contains(bank, "*") {
			baseQuery = baseQuery.Where("bank LIKE ?", strings.Replace(bank, "*", "%", -1))
		} else if strings.Contains(bank, "?") {
			baseQuery = baseQuery.Where("bank LIKE ?", strings.Replace(bank, "?", "_", -1))
		} else {
			baseQuery = baseQuery.Where("bank = ?", bank)
		}
	}
	if key != "" {
		if strings.Contains(key, "*") {
			baseQuery = baseQuery.Where("psql_key LIKE ?", strings.Replace(key, "*", "%", -1))
		} else if strings.Contains(key, "?") {
			baseQuery = baseQuery.Where("psql_key LIKE ?", strings.Replace(key, "?", "_", -1))
		} else {
			baseQuery = baseQuery.Where("psql_key = ?", key)
		}
	}

	if since != "" {
		fromTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'since' date format"})
			return
		}
		baseQuery = baseQuery.Where("data_changed >= ?", fromTime)
	}
	if until != "" {
		toTime, err := time.Parse(time.RFC3339, until)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'until' date format"})
			return
		}
		baseQuery = baseQuery.Where("data_changed <= ?", toTime)
	}

	// Get total count for pagination metadata before applying pagination
	var totalCount int64
	baseQuery.Count(&totalCount)

	// Setup pagination for the query
	paginatedQuery := baseQuery.Offset((page - 1) * perPage).Limit(perPage)

	// Execute the query to find results
	paginatedQuery.Find(&saltCaches)

	// Generate pagination links
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
	if int64((page-1)*perPage+len(saltCaches)) < totalCount {
		nextPage = fmt.Sprintf("%s?page=%d&per_page=%d", baseURL, page+1, perPage)
	}

	// Prepare the pagination response
	paging := dto.PageResponse{
		PerPage:  int64(perPage),
		NumPages: int64(math.Ceil(float64(totalCount) / float64(perPage))),
		Count:    totalCount,
		Next:     nextPage,
		Previous: previousPage,
	}

	// Construct and return the structured response
	response := dto.SaltCachePageResponse{
		Results: saltCaches,
		Paging:  paging,
	}

	// Check for empty result set
	if len(saltCaches) == 0 {
		httputil.NewError(c, http.StatusNotFound, "No salt_cache items present.")
		return
	}

	// Return paginated results
	c.JSON(http.StatusOK, response)
}

// GetSaltCache func one saltCache by SaltCache
//
//	@Description	Get one salt_cache item by bank and key
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.SaltCache
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/{bank}/{key} [get]
//	@Param			bank	path	string	true	"bank of the salt cache item to retrieve"
//	@Param			key		path	string	true	"key of the salt cache item to retrieve"
//	@Security		Bearer
func GetSaltCacheBankKey(c *gin.Context) {
	db := db.DB
	var saltCache model.SaltCache

	// Read the param saltCache
	bank, _ := url.PathUnescape(c.Param("bank"))
	key, _ := url.PathUnescape(c.Param("key"))

	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")

		return
	}

	// Find the saltCache with the given bank
	db.Where("bank = ? AND psql_key = ?", bank, key).Find(&saltCache)

	// If no such saltCache present return an error
	if saltCache.Bank == "" {
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	// Return the saltCache with the Id
	c.JSON(http.StatusOK, saltCache)
}

// DeleteSaltCache func one saltCache by SaltCache
//
//	@Description	Delete one salt_cache item by bank and key
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/{bank}/{key} [delete]
//	@Param			bank	path	string	true	"bank of the salt cache item to delete"
//	@Param			key		path	string	true	"key of the salt cache item to delete"
//	@Security		Bearer
func DeleteSaltCacheBankKey(c *gin.Context) {
	db := db.DB
	var saltCache model.SaltCache

	// Read the param saltCache
	bank, _ := url.PathUnescape(c.Param("bank"))
	key, _ := url.PathUnescape(c.Param("key"))

	if bank == "" || key == "" {
		httputil.NewError(c, http.StatusBadRequest, "bank and key required.")
		return
	}

	// Find the saltCache with the given bank
	db.Where("bank = ? AND psql_key = ?", bank, key).Find(&saltCache)

	// If no such saltCache present return an error
	if saltCache.Bank == "" {
		httputil.NewError(c, http.StatusNotFound, "No salt_cache item present.")
		return
	}

	err := db.Delete(&saltCache, "bank = ? AND psql_key = ?", bank, key).Error

	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete cache item.")
		return
	}

	// Return success message
	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted salt_cache item %s/%s", bank, key))
}

// DeleteSaltCache func one saltCache by the item's UUID
//
//	@Description	Delete one salt_cache item by UUID
//	@Tags			SaltCache
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_cache/uuid/{uuid} [delete]
//	@Param			uuid	path	string	true	"uuid of the salt cache item to delete"
//	@Security		Bearer
func DeleteSaltCacheUUID(c *gin.Context) {
	db := db.DB
	var saltCache model.SaltCache

	// Read the param saltCache
	uuid := c.Param("uuid")

	if uuid == "" {
		httputil.NewError(c, http.StatusBadRequest, "uuid required.")
		return
	}

	// Find the saltCache with the given bank
	db.Where("id = ?", uuid).Find(&saltCache)

	// If no such saltCache present return an error
	if saltCache.Bank == "" {
		httputil.NewError(c, http.StatusBadRequest, "No salt_cache item present.")
		return
	}

	// Delete the note and return error if encountered
	err := db.Delete(&saltCache, "id = ? ", uuid).Error

	if err != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete cache item.")
		return
	}

	// Return success message
	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted salt_cache item %s", uuid))
}
