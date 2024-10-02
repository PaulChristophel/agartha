package conformity

import (
	"net/http"

	_ "github.com/PaulChristophel/agartha/server/httputil"

	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	isRefreshing *bool
)

type RefreshStatus struct {
	Status  string `json:"status" example:"pending"`
	Message string `json:"message" example:"Materialized view refresh is already in progress"`
}

// MarshalLogObject implements zapcore.ObjectMarshaler to allow logging of RefreshStatus
func (r RefreshStatus) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("status", r.Status)
	enc.AddString("message", r.Message)
	return nil
}

// Refresh retrieves the status of the conformity table.
//
//	@Summary		Get the status of the refresh operation.
//	@Description	Retrieve the status of the conformity materialized view refresh.
//	@Tags			Conformity
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	RefreshStatus
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/api/v1/conformity/refresh [get]
//	@Security		Bearer
func Refresh(c *gin.Context) {
	var status RefreshStatus
	log := logger.GetLogger()

	if *isRefreshing {
		status = RefreshStatus{
			Status:  "pending",
			Message: "Materialized view refresh is already in progress",
		}
	} else {
		status = RefreshStatus{
			Status:  "available",
			Message: "Materialized view refresh complete",
		}
	}
	log.Debug("Returning refresh status", zap.Object("refresh", status))

	c.JSON(http.StatusOK, status)
}

func SetOptions(refreshing *bool) {
	isRefreshing = refreshing
}
