package routes

import (
	"net/http"

	// So swagger can document the function
	_ "github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
)

func AddPingRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/ping")

	grp.GET("", Ping)
}

// Ping godoc
//
//	@Summary	ping the server
//	@Schemes
//	@Description	Pings the server to see if it is alive.
//	@Tags			Ping
//	@Accept			json
//	@Produce		json
//	@Success		200	{string}	string	"pong"
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/ping [get]
func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}
