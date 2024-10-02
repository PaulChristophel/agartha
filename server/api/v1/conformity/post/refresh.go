package conformity

import (
	"net/http"
	"sync"

	_ "github.com/PaulChristophel/agartha/server/httputil"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	refreshMu    sync.Mutex
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

// Refresh retrieves paginated conformity items based on the provided limit and page query parameters.
//
//	@Summary		Refresh the conformity data for all minions.
//	@Description	Refresh the conformity materialized view.
//	@Tags			Conformity
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	RefreshStatus
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/api/v1/conformity/refresh [post]
//	@Security		Bearer
func Refresh(c *gin.Context) {
	log := logger.GetLogger()
	sugar := log.Sugar()
	if *isRefreshing {
		status := RefreshStatus{
			Status:  "pending",
			Message: "Materialized view refresh is already in progress",
		}
		log.Debug("Returning refresh status", zap.Object("refresh", status))
		c.JSON(http.StatusOK, status)
		return
	}
	log.Debug("Locking the table to block futher goroutines")
	refreshMu.Lock()
	go func() {
		*isRefreshing = true
		defer func() {
			*isRefreshing = false
			log.Debug("Unlocking the table and allowing goroutines")
			refreshMu.Unlock()
		}()

		result := db.DB.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY mat_conformity;")
		if result.Error != nil {
			sugar.Errorf("Failed to refresh materialized view: %s", result.Error)
			return
		}

		log.Info("Materialized view refreshed successfully")
	}()

	status := RefreshStatus{
		Status:  "success",
		Message: "Materialized view refresh initiated",
	}
	log.Debug("Returning refresh status", zap.Object("refresh", status))
	c.JSON(http.StatusOK, status)
}

func SetOptions(refreshing *bool) {
	isRefreshing = refreshing
}
