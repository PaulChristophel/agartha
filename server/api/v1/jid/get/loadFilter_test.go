package jid

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestGetJIDsValidatesLoadQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		useJSONB bool
		wantBody string
	}{
		{
			name:     "invalid JSON",
			query:    `{`,
			useJSONB: true,
			wantBody: `{"code":400,"message":"invalid 'load_query' JSON"}`,
		},
		{
			name:     "empty clauses",
			query:    `{"logic":"and","clauses":[]}`,
			useJSONB: true,
			wantBody: `{"code":400,"message":"'load_query' must contain at least one clause"}`,
		},
		{
			name:     "legacy text tables",
			query:    `{"logic":"and","clauses":[{"scope":"root","path":["fun"],"operator":"exists"}]}`,
			useJSONB: false,
			wantBody: `{"code":400,"message":"load filtering requires JSONB database tables"}`,
		},
		{
			name:     "numeric operator with string type",
			query:    `{"logic":"and","clauses":[{"scope":"root","path":["retcode"],"operator":"gt","value":"0","value_type":"string"}]}`,
			useJSONB: true,
			wantBody: `{"code":400,"message":"invalid load query clause 1: numeric operators require an int or float value type"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installJIDMockDatabase(t, tt.useJSONB)
			params := url.Values{}
			params.Set("load_query", tt.query)
			response := serveJIDRequest("/jid?" + params.Encode())

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetJIDsAppliesLoadQueryBeforeCountAndResults(t *testing.T) {
	_, mock := installJIDMockDatabase(t, true)
	alterTime := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	query := `{"logic":"and","clauses":[{"scope":"root","path":["fun"],` +
		`"operator":"eq","value":"state.sls","value_type":"string"}]}`
	countQuery := `SELECT count\(\*\) FROM "jids" WHERE \(\(EXISTS \(SELECT 1 FROM ` +
		`\(SELECT jsonb_extract_path\("load", \$1\) AS value\) AS load_target WHERE ` +
		`load_target.value = CAST\(\$2 AS jsonb\)\)\)\) AND alter_time >= \$3`
	selectionQuery := `SELECT "jid","alter_time" FROM "jids" WHERE \(\(EXISTS \(SELECT 1 FROM ` +
		`\(SELECT jsonb_extract_path\("load", \$1\) AS value\) AS load_target WHERE ` +
		`load_target.value = CAST\(\$2 AS jsonb\)\)\)\) AND alter_time >= \$3 ` +
		`ORDER BY alter_time desc LIMIT \$4`

	mock.ExpectQuery(countQuery).
		WithArgs("fun", `"state.sls"`, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(selectionQuery).
		WithArgs("fun", `"state.sls"`, sqlmock.AnyArg(), 50).
		WillReturnRows(sqlmock.NewRows([]string{"jid", "alter_time"}).
			AddRow("20260831120000000000", alterTime))

	params := url.Values{}
	params.Set("load_query", query)
	params.Set("since", "2026-08-31T00:00:00Z")
	response := serveJIDRequest("/jid?" + params.Encode())

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.JSONEq(t, `{
		"paging":{"per_page":50,"num_pages":1,"count":1,"next":"","previous":""},
		"results":[{"jid":"20260831120000000000","load":null,
		"alter_time":"2026-08-31T12:00:00Z"}]
	}`, response.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestBuildLoadQueryClauseSupportsNestedAndArbitraryEntries(t *testing.T) {
	predicate, args, err := buildLoadQueryClause(loadQueryClause{
		Scope: "any_key", ContainerPath: []string{"return"}, Path: []string{"result"},
		Operator: "eq", Value: "false", ValueType: "bool",
	})

	require.NoError(t, err)
	require.Contains(t, predicate, `jsonb_extract_path("load", ?)`)
	require.Contains(t, predicate, "CROSS JOIN LATERAL jsonb_each")
	require.Equal(t, []any{"return", "result", "false"}, args)
}

func installJIDMockDatabase(t *testing.T, jsonbEnabled bool) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	_, err := logger.InitLogger(gin.TestMode)
	require.NoError(t, err)

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	previousDB := db.DB
	previousTable := table
	previousUseJSONB := useJSONB
	db.DB = gormDB
	SetOptions(config.SaltDBTables{JIDs: "jids", UseJSONB: jsonbEnabled})
	t.Cleanup(func() {
		db.DB = previousDB
		table = previousTable
		useJSONB = previousUseJSONB
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})
	return gormDB, mock
}

func serveJIDRequest(requestURL string) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/jid", GetJIDs)
	request := httptest.NewRequest(http.MethodGet, requestURL, nil)
	request.Host = "example.com"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
