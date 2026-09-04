package saltMinion

import (
	"net/http"
	"sync/atomic"

	_ "github.com/PaulChristophel/agartha/server/httputil"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	isRefreshing *atomic.Bool
	refreshError *atomic.Pointer[string]
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
//	@Tags			SaltMinion
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	RefreshStatus
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/api/v1/salt_minion/keys/refresh [post]
//	@Security		Bearer
func RefreshKeys(c *gin.Context) {
	log := logger.GetLogger()
	sugar := log.Sugar()
	if !isRefreshing.CompareAndSwap(false, true) {
		status := RefreshStatus{
			Status:  "pending",
			Message: "Materialized view refresh is already in progress",
		}
		log.Debug("Returning refresh status", zap.Object("refresh", status))
		c.JSON(http.StatusOK, status)
		return
	}
	refreshError.Store(nil)
	go func() {
		defer func() {
			isRefreshing.Store(false)
		}()

		result := db.DB.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY mat_salt_minions_grains_keys;")
		if result.Error != nil {
			message := "Failed to refresh the grains dropdown list"
			refreshError.Store(&message)
			sugar.Errorf("%s: %s", message, result.Error)
			return
		}
		result = db.DB.Exec("REFRESH MATERIALIZED VIEW CONCURRENTLY mat_salt_minions_pillar_keys;")
		if result.Error != nil {
			message := "Failed to refresh the pillar dropdown list"
			refreshError.Store(&message)
			sugar.Errorf("%s: %s", message, result.Error)
			return
		}

		log.Info("Materialized views refreshed successfully")
	}()

	status := RefreshStatus{
		Status:  "success",
		Message: "Materialized view refresh initiated",
	}
	log.Debug("Returning refresh status", zap.Object("refresh", status))
	c.JSON(http.StatusOK, status)
}

func SetRefreshState(refreshing *atomic.Bool, lastError *atomic.Pointer[string]) {
	isRefreshing = refreshing
	refreshError = lastError
}
