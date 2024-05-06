package authUser

import (
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	model "github.com/PaulChristophel/agartha/server/model/agartha"
	"github.com/gin-gonic/gin"
)

func AddRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/auth_user")

	// grp.GET("/", GetauthUsers)
	grp.GET("/:id", GetAuthUser)
}

// GetauthUser func one authUser by AuthUser
//
//	@Description	Get one authUser by AuthUser ID
//	@Tags			AuthUser
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.AuthUser
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/secure/auth_user/{id} [get]
//	@Param			id	path	string	false	"authUsers to return"
//	@Security		Bearer
func GetAuthUser(c *gin.Context) {
	db := db.DB
	var authUser model.AuthUser

	// Read the param authUser
	id := c.Param("id")

	// Find the authUser with the given Id
	db.Find(&authUser, "id = ?", id)

	// If no such authUser present return an error
	if authUser.ID <= 0 {
		httputil.NewError(c, http.StatusNotFound, "No auth_user data present.")
		return
	}

	// Return the authUser with the Id
	c.JSON(http.StatusOK, authUser)
}
