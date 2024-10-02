package saltMinion

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/dto"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListSaltMinionGrainsKeys func retrieves the list of JSON keys from the materialized view with optional filtering and pagination
//
//	@Summary		Retrieve the list of JSON keys from the salt_minions grains with optional filtering and pagination
//	@Description	Retrieve the list of JSON keys from the salt_minions grains with optional filtering and pagination
//	@Tags			SaltMinion
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dto.SaltMinionGrainsKeyResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_minion/grains_keys [get]
//	@Param			exact_includes	query	string	false	"Comma separated list of exact patterns to include"
//	@Param			like_includes	query	string	false	"Comma separated list of like patterns to include"
//	@Param			exact_excludes	query	string	false	"Comma separated list of exact patterns to exclude"
//	@Param			like_excludes	query	string	false	"Comma separated list of like patterns to exclude"
//	@Param			page			query	int		false	"Page number of results to retrieve"
//	@Param			per_page		query	int		false	"Number of items per page"
//	@Security		Bearer
func ListSaltMinionGrainsKeys(c *gin.Context) {
	log := logger.GetLogger()
	var grainsKeys []string

	// Read query parameters
	exactIncludes := c.Query("exact_includes")
	likeIncludes := c.Query("like_includes")
	exactExcludes := c.Query("exact_excludes")
	likeExcludes := c.Query("like_excludes")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	log.Info("Received request to get salt minions grains keys",
		zap.String("exact_includes", exactIncludes),
		zap.String("like_includes", likeIncludes),
		zap.String("exact_excludes", exactExcludes),
		zap.String("like_excludes", likeExcludes),
		zap.Int("page", page),
		zap.Int("per_page", perPage))

	// Create the base query
	query := db.DB.Table("mat_salt_minions_grains_keys")

	// Apply exact includes if provided
	if exactIncludes != "" {
		exactIncludePatterns := strings.Split(exactIncludes, ",")
		for _, pattern := range exactIncludePatterns {
			query = query.Where("path = ?", pattern)
		}
		log.Debug("Applied exact includes", zap.Strings("exact_includes", exactIncludePatterns))
	}

	// Apply like includes if provided
	if likeIncludes != "" {
		likeIncludePatterns := strings.Split(likeIncludes, ",")
		for _, pattern := range likeIncludePatterns {
			query = query.Where("path LIKE ?", "%"+pattern+"%")
		}
		log.Debug("Applied like includes", zap.Strings("like_includes", likeIncludePatterns))
	}

	// Apply exact excludes if provided
	if exactExcludes != "" {
		exactExcludePatterns := strings.Split(exactExcludes, ",")
		for _, pattern := range exactExcludePatterns {
			query = query.Where("path != ?", pattern)
		}
		log.Debug("Applied exact excludes", zap.Strings("exact_excludes", exactExcludePatterns))
	}

	// Apply like excludes if provided
	if likeExcludes != "" {
		likeExcludePatterns := strings.Split(likeExcludes, ",")
		for _, pattern := range likeExcludePatterns {
			query = query.Where("path NOT LIKE ?", "%"+pattern+"%")
		}
		log.Debug("Applied like excludes", zap.Strings("like_excludes", likeExcludePatterns))
	}

	// Get the total count for pagination
	var totalCount int64
	query.Count(&totalCount)

	// Get the raw SQL query and log it
	sql := query.Statement.SQL.String()
	args := query.Statement.Vars
	log.Debug("Generated SQL query", zap.String("sql", sql), zap.Any("args", args))

	// Apply pagination
	query = query.Offset((page - 1) * perPage).Limit(perPage)

	// Execute the query to fetch grains keys
	if err := query.Find(&grainsKeys).Error; err != nil {
		log.Error("Failed to fetch salt minions grains keys", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch salt minions grains keys.")
		return
	}

	// If no grains keys are present, return an error
	if len(grainsKeys) == 0 {
		log.Debug("No grains keys present in the materialized view")
		httputil.NewError(c, http.StatusNotFound, "No grains keys present in the materialized view.")
		return
	}

	// Prepare the pagination response
	paging := dto.PageResponse{
		PerPage:  int64(perPage),
		NumPages: int64(math.Ceil(float64(totalCount) / float64(perPage))),
		Count:    totalCount,
	}

	// Construct and return the structured response
	response := dto.SaltMinionGrainsKeyResponse{
		Paging:  paging,
		Results: grainsKeys,
	}

	log.Debug("Successfully retrieved salt minions grains keys", zap.Int("count", len(grainsKeys)))
	// Return the paginated list of data keys
	c.JSON(http.StatusOK, response)
}
