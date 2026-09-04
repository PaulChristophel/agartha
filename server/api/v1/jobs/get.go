package jobs

import (
	"net/http"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/PaulChristophel/agartha/server/model/custom"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Detail combines a stored job and all its returns. Unknown counts remain null.
type Detail struct {
	JID             string             `json:"jid"`
	Load            custom.JSON        `json:"load"`
	Returns         []model.SaltReturn `json:"returns"`
	TargetedCount   *int               `json:"targeted_count"`
	ReturnedCount   int                `json:"returned_count"`
	SuccessfulCount int                `json:"successful_count"`
	FailedCount     int                `json:"failed_count"`
	PendingCount    *int               `json:"pending_count"`
	StartedAt       *time.Time         `json:"started_at"`
	LastReturnAt    *time.Time         `json:"last_return_at"`
}

type handler struct {
	database *gorm.DB
	tables   config.SaltDBTables
}

// getJob retrieves a job with its complete stored returns.
//
// @Summary Get a job with its returns and summary.
// @Description Returns the original load and all stored returns, ordered by minion ID. Targeted and pending counts are null unless load.Minions or load.minions contains a resolved list of minion IDs; pending counts targeted IDs without a return and does not imply they are still running. Success counts use the stored return success flag. started_at is the stored job alter_time, not a measured execution start; last_return_at is the newest return alter_time. Jobs without returns return an empty array and a null last_return_at.
// @Tags Jobs
// @Produce json
// @Param jid path string true "Stored job identifier"
// @Success 200 {object} Detail
// @Failure 401 {object} httputil.HTTPError401
// @Failure 403 {object} httputil.HTTPError403
// @Failure 404 {object} httputil.HTTPError404
// @Failure 500 {object} httputil.HTTPError500
// @Router /api/v1/jobs/{jid} [get]
// @Security Bearer
func (h handler) getJob(c *gin.Context) {
	database := h.database.WithContext(c.Request.Context())
	var job model.JID
	result := database.Table(h.tables.JIDs).Where("jid = ?", c.Param("jid")).Find(&job)
	if result.Error != nil {
		h.fail(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		httputil.NewError(c, http.StatusNotFound, "No jid present.")
		return
	}
	returns := make([]model.SaltReturn, 0)
	if err := database.Table(h.tables.SaltReturns).Where("jid = ?", job.JID).Order("id ASC").Find(&returns).Error; err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, summarize(job, returns))
}

func (h handler) fail(c *gin.Context, err error) {
	logger.GetLogger().Error("Failed to fetch job details", zap.Error(err))
	httputil.NewError(c, http.StatusInternalServerError, "Failed to fetch job details.")
}

func summarize(job model.JID, returns []model.SaltReturn) Detail {
	detail := Detail{JID: job.JID, Load: job.Load, Returns: returns, StartedAt: job.AlterTime, ReturnedCount: len(returns)}
	returnedIDs := make(map[string]bool, len(returns))
	for _, ret := range returns {
		returnedIDs[ret.ID] = true
		if ret.Success {
			detail.SuccessfulCount++
		} else {
			detail.FailedCount++
		}
		if ret.AlterTime != nil && (detail.LastReturnAt == nil || ret.AlterTime.After(*detail.LastReturnAt)) {
			detail.LastReturnAt = ret.AlterTime
		}
	}
	if targets, known := resolvedTargets(job.Load); known {
		targeted, pending := len(targets), 0
		for id := range targets {
			if !returnedIDs[id] {
				pending++
			}
		}
		detail.TargetedCount, detail.PendingCount = &targeted, &pending
	}
	return detail
}

func resolvedTargets(load custom.JSON) (map[string]bool, bool) {
	object, ok := load.Data.(map[string]any)
	if !ok {
		return nil, false
	}
	value, exists := object["Minions"]
	if !exists {
		value = object["minions"]
	}
	list, ok := value.([]any)
	if !ok {
		return nil, false
	}
	targets := make(map[string]bool, len(list))
	for _, value := range list {
		id, ok := value.(string)
		if !ok || id == "" {
			return nil, false
		}
		targets[id] = true
	}
	return targets, true
}
