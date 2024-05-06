package routes

import (
	"net/http"

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
//	@Success		200	{string}	pong "pong"
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		403	{object}	httputil.HTTPError403
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/ping [get]
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, "pong")
}
