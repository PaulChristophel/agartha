package saltCache

import (
	"fmt"
	"net/http"

	// "strings"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
)

// DeleteSaltCacheUUID func one saltCache by the item's UUID
//
//	@Summary		Delete one salt_cache item by UUID
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
	db := db.DB.Table(table)
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
