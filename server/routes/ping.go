package routes

import (
	"context"
	"net/http"
	"time"

	"github.com/PaulChristophel/agartha/server/db"
	// So swagger can document the function
	_ "github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
)

func AddPingRoutes(rg *gin.RouterGroup) {
	rg.GET("/ping", Ping)
	rg.GET("/ready", Ready)
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

// Ready reports whether dependencies required to serve application traffic are available.
func Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := db.Ready(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
