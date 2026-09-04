package jobs

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/PaulChristophel/agartha/server/model/custom"
	model "github.com/PaulChristophel/agartha/server/model/salt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestJobDetail(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)
	for _, scenario := range []string{"legacy", "jsonb", "empty", "missing", "job error", "return error"} {
		t.Run(scenario, func(t *testing.T) {
			sqlDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			t.Cleanup(func() { mock.ExpectClose(); require.NoError(t, sqlDB.Close()) })
			database, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
			require.NoError(t, err)
			start := time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC)
			load := `{"Minions":["a","b","c","a"],"fun":"test.ping"}`
			query := mock.ExpectQuery(`SELECT \* FROM "custom_jobs" WHERE jid = \$1`).WithArgs("20260904120000000000")
			status := 200
			switch scenario {
			case "job error":
				query.WillReturnError(errors.New("private database error"))
				status = 500
			case "missing":
				query.WillReturnRows(sqlmock.NewRows([]string{"jid", "load", "alter_time"}))
				status = 404
			default:
				query.WillReturnRows(sqlmock.NewRows([]string{"jid", "load", "alter_time"}).AddRow("20260904120000000000", load, start))
				returnsQuery := mock.ExpectQuery(`SELECT \* FROM "custom_returns" WHERE jid = \$1 ORDER BY id ASC`).WithArgs("20260904120000000000")
				rows := sqlmock.NewRows([]string{"jid", "id", "fun", "return", "full_ret", "success", "alter_time"})
				switch scenario {
				case "return error":
					returnsQuery.WillReturnError(errors.New("private database error"))
					status = 500
				case "empty":
					returnsQuery.WillReturnRows(rows)
				default:
					var yes, no any = "true", "false"
					if scenario == "jsonb" {
						yes, no = true, false
					}
					returnsQuery.WillReturnRows(rows.AddRow("20260904120000000000", "a", "test.ping", `true`, `{}`, yes, start.Add(time.Minute)).AddRow("20260904120000000000", "b", "test.ping", `false`, `{}`, no, start.Add(2*time.Minute)))
				}
			}
			router := gin.New()
			AddRoutes(router.Group("/api/v1"), database, config.SaltDBTables{JIDs: "custom_jobs", SaltReturns: "custom_returns", UseJSONB: scenario == "jsonb"})
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/jobs/20260904120000000000", nil))
			require.Equal(t, status, response.Code, response.Body.String())
			require.NotContains(t, response.Body.String(), "private database error")
			if status == 200 {
				var detail Detail
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &detail))
				require.Equal(t, 3, *detail.TargetedCount)
				require.Equal(t, start, *detail.StartedAt)
				require.JSONEq(t, load, mustJSON(t, detail.Load))
				if scenario == "empty" {
					require.NotNil(t, detail.Returns)
					require.Empty(t, detail.Returns)
					require.Nil(t, detail.LastReturnAt)
					require.Equal(t, 3, *detail.PendingCount)
				} else {
					require.Len(t, detail.Returns, 2)
					require.Equal(t, 2, detail.ReturnedCount)
					require.Equal(t, 1, detail.SuccessfulCount)
					require.Equal(t, 1, detail.FailedCount)
					require.Equal(t, 1, *detail.PendingCount)
					require.Equal(t, start.Add(2*time.Minute), *detail.LastReturnAt)
					require.Equal(t, true, detail.Returns[0].Return.Data)
				}
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return string(data)
}

func TestSummaryTargets(t *testing.T) {
	for _, test := range []struct {
		load              string
		targeted, pending *int
	}{
		{`{"tgt":"*"}`, nil, nil},
		{`{"Minions":null}`, nil, nil},
		{`{"Minions":["a",1]}`, nil, nil},
		{`{"Minions":[""]}`, nil, nil},
		{`"invalid load"`, nil, nil},
		{`{"minions":[]}`, new(0), new(0)},
		{`{"minions":["a","a","b"]}`, new(2), new(1)},
	} {
		t.Run(test.load, func(t *testing.T) {
			var load custom.JSON
			require.NoError(t, json.Unmarshal([]byte(test.load), &load))
			result := summarize(model.JID{Load: load}, []model.SaltReturn{{ID: "a", Success: true}, {ID: "unexpected", Success: true}})
			require.Equal(t, test.targeted, result.TargetedCount)
			require.Equal(t, test.pending, result.PendingCount)
			require.Nil(t, result.StartedAt)
			require.Nil(t, result.LastReturnAt)
		})
	}
}
