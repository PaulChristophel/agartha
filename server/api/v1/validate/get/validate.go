package validate

import (
	"net/http"

	// So swagger can document the function
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
)

// Validate godoc
//
//	@Summary	Validate auth token
//	@Schemes
//	@Description	Authenticate with the auth token to see if it is still valid
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	httputil.HTTPError200
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/api/v1/validate [get]
//	@Security		Bearer
func Validate(c *gin.Context) {
	httputil.NewError(c, http.StatusOK, "Success")
}
