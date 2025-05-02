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

	uuid, ok := httputil.MustUnescapeParam(c, "uuid")
	if !ok {
		return
	}

	if uuid == "" {
		httputil.NewError(c, http.StatusBadRequest, "uuid is required.")
		return
	}

	// Safely delete with WHERE clause only
	tx := db.Where("id = ?", uuid).Delete(&model.SaltCache{})
	if tx.Error != nil {
		httputil.NewError(c, http.StatusInternalServerError, "Failed to delete cache item.")
		return
	}
	if tx.RowsAffected == 0 {
		httputil.NewError(c, http.StatusNotFound, fmt.Sprintf("No salt_cache item found with UUID: %s", uuid))
		return
	}

	httputil.NewError(c, http.StatusOK, fmt.Sprintf("Deleted salt_cache item %s", uuid))
}
