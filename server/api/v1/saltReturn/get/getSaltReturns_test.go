package saltReturn

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

func TestGetSaltReturnsValidatesReturnFilter(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		useJSONB   bool
		wantBody   string
		wantStatus int
	}{
		{
			name:       "match without filter",
			query:      "return_match=key_exists",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"'return_filter' is required for key_exists"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "filter without match",
			query:      "return_filter=chrony",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"'return_match' is required when return filter parameters are provided"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown match",
			query:      "return_match=contains&return_filter=chrony",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"invalid 'return_match' value: must be key_exists, string_equals, or key_field_equals"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty object key",
			query:      "return_match=key_exists&return_filter=",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"'return_filter' cannot be empty for key_exists"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "legacy text tables",
			query:      "return_match=string_equals&return_filter=failed",
			useJSONB:   false,
			wantBody:   `{"code":400,"message":"return filtering requires JSONB database tables"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "nested field missing",
			query:      "return_match=key_field_equals&return_key=state&return_value=false&return_value_type=bool",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"'return_field' is required for key_field_equals"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid nested boolean",
			query:      "return_match=key_field_equals&return_key=state&return_field=result&return_value=nope&return_value_type=bool",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"invalid boolean 'return_value'"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installSaltReturnMockDatabase(t, tt.useJSONB)

			response := serveSaltReturnRequest("/salt_return?" + tt.query)

			require.Equal(t, tt.wantStatus, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetSaltReturnsAppliesTypedJSONBFilters(t *testing.T) {
	tests := []struct {
		name           string
		fun            string
		match          string
		filter         string
		key            string
		field          string
		value          string
		valueType      string
		expectedArg    string
		countQuery     string
		selectionQuery string
	}{
		{
			name:        "top-level object key exists",
			fun:         "state.apply",
			match:       returnMatchKeyExists,
			filter:      "pkg_|-chrony_|-chrony_|-installed",
			expectedArg: "pkg_|-chrony_|-chrony_|-installed",
			countQuery: `SELECT count\(\*\) FROM "salt_returns" WHERE \(jsonb_typeof\("return"\) = 'object' ` +
				`AND jsonb_exists\("return", \$1\)\) AND fun = \$2 AND alter_time >= \$3`,
			selectionQuery: `SELECT "fun","jid","id","success","alter_time" FROM "salt_returns" WHERE ` +
				`\(jsonb_typeof\("return"\) = 'object' AND jsonb_exists\("return", \$1\)\) ` +
				`AND fun = \$2 AND alter_time >= \$3 ORDER BY alter_time desc LIMIT \$4`,
		},
		{
			name:        "root string equals",
			fun:         "event.fire",
			match:       returnMatchStringEquals,
			filter:      "Unhandled exception running event.fire",
			expectedArg: "Unhandled exception running event.fire",
			countQuery: `SELECT count\(\*\) FROM "salt_returns" WHERE \(jsonb_typeof\("return"\) = 'string' ` +
				`AND "return" = to_jsonb\(CAST\(\$1 AS text\)\)\) AND fun = \$2 AND alter_time >= \$3`,
			selectionQuery: `SELECT "fun","jid","id","success","alter_time" FROM "salt_returns" WHERE ` +
				`\(jsonb_typeof\("return"\) = 'string' AND "return" = to_jsonb\(CAST\(\$1 AS text\)\)\) ` +
				`AND fun = \$2 AND alter_time >= \$3 ORDER BY alter_time desc LIMIT \$4`,
		},
		{
			name:        "field below top-level key equals boolean",
			fun:         "state.highstate",
			match:       returnMatchKeyFieldEquals,
			key:         "file_|-Install root Bashrc_|-/root/.bashrc_|-managed",
			field:       "result",
			value:       "false",
			valueType:   "bool",
			expectedArg: `{"file_|-Install root Bashrc_|-/root/.bashrc_|-managed":{"result":false}}`,
			countQuery: `SELECT count\(\*\) FROM "salt_returns" WHERE \(jsonb_typeof\("return"\) = 'object' ` +
				`AND "return" @> CAST\(\$1 AS jsonb\)\) AND fun = \$2 AND alter_time >= \$3`,
			selectionQuery: `SELECT "fun","jid","id","success","alter_time" FROM "salt_returns" WHERE ` +
				`\(jsonb_typeof\("return"\) = 'object' AND "return" @> CAST\(\$1 AS jsonb\)\) ` +
				`AND fun = \$2 AND alter_time >= \$3 ORDER BY alter_time desc LIMIT \$4`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installSaltReturnMockDatabase(t, true)
			alterTime := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)

			mock.ExpectQuery(tt.countQuery).
				WithArgs(tt.expectedArg, tt.fun, sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery(tt.selectionQuery).
				WithArgs(tt.expectedArg, tt.fun, sqlmock.AnyArg(), 50).
				WillReturnRows(sqlmock.NewRows([]string{"fun", "jid", "id", "success", "alter_time"}).
					AddRow(tt.fun, "20260828120000000000", "minion.example.com", "false", alterTime))

			params := url.Values{}
			params.Set("fun", tt.fun)
			params.Set("return_match", tt.match)
			if tt.filter != "" {
				params.Set("return_filter", tt.filter)
			}
			if tt.key != "" {
				params.Set("return_key", tt.key)
				params.Set("return_field", tt.field)
				params.Set("return_value", tt.value)
				params.Set("return_value_type", tt.valueType)
			}
			params.Set("since", "2026-08-28T00:00:00Z")
			response := serveSaltReturnRequest("/salt_return?" + params.Encode())

			require.Equal(t, http.StatusOK, response.Code)
			require.JSONEq(t, `{
				"paging": {
					"per_page": 50,
					"num_pages": 1,
					"count": 1,
					"next": "",
					"previous": ""
				},
				"results": [{
					"fun": "`+tt.fun+`",
					"jid": "20260828120000000000",
					"return": null,
					"full_ret": null,
					"id": "minion.example.com",
					"success": false,
					"alter_time": "2026-08-28T12:00:00Z"
				}]
			}`, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBuildReturnQueryClause(t *testing.T) {
	tests := []struct {
		name          string
		clause        returnQueryClause
		wantPredicate string
		wantArgs      []any
	}{
		{
			name: "specific object key numeric greater than",
			clause: returnQueryClause{
				Scope: "key", Key: "file_|-Install root Bashrc_|-/root/.bashrc_|-managed",
				Path: []string{"duration"}, Operator: "gt", Value: "10", ValueType: "float",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path("return", ?, ?) AS value) AS return_target ` +
				`WHERE jsonb_typeof(return_target.value) = 'number' AND (return_target.value #>> '{}')::numeric > CAST(? AS numeric))`,
			wantArgs: []any{"file_|-Install root Bashrc_|-/root/.bashrc_|-managed", "duration", "10"},
		},
		{
			name: "matching field under any object key",
			clause: returnQueryClause{
				Scope: "any_key", Path: []string{"result"}, Operator: "eq", Value: "false", ValueType: "bool",
			},
			wantPredicate: `jsonb_typeof("return") = 'object' AND EXISTS (` +
				`SELECT 1 FROM jsonb_each("return") AS return_entry(key, value) WHERE ` +
				`EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path(return_entry.value, ?) AS value) AS return_target ` +
				`WHERE return_target.value = CAST(? AS jsonb)))`,
			wantArgs: []any{"result", "false"},
		},
		{
			name: "deep case-insensitive string contains",
			clause: returnQueryClause{
				Scope: "key", Key: "an-object-key", Path: []string{"changes", "diff"},
				Operator: "icontains", Value: "kernel", ValueType: "string",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path("return", ?, ?, ?) AS value) AS return_target ` +
				`WHERE jsonb_typeof(return_target.value) = 'string' AND ` +
				`strpos(lower((return_target.value #>> '{}')), lower(?)) > 0)`,
			wantArgs: []any{"an-object-key", "changes", "diff", "kernel"},
		},
		{
			name: "root regular expression",
			clause: returnQueryClause{
				Scope: "root", Operator: "regex", Value: "(?i)exception", ValueType: "string",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT "return" AS value) AS return_target ` +
				`WHERE jsonb_typeof(return_target.value) = 'string' AND (return_target.value #>> '{}') ~ ?)`,
			wantArgs: []any{"(?i)exception"},
		},
		{
			name: "nested field does not exist",
			clause: returnQueryClause{
				Scope: "key", Key: "an-object-key", Path: []string{"changes", "diff"}, Operator: "not_exists",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path("return", ?, ?, ?) AS value) AS return_target ` +
				`WHERE return_target.value IS NULL)`,
			wantArgs: []any{"an-object-key", "changes", "diff"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate, args, err := buildReturnQueryClause(tt.clause)
			require.NoError(t, err)
			require.Equal(t, tt.wantPredicate, predicate)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestGetSaltReturnsValidatesAdvancedReturnQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantBody string
	}{
		{
			name:     "invalid JSON",
			query:    `{`,
			wantBody: `{"code":400,"message":"invalid 'return_query' JSON"}`,
		},
		{
			name:     "empty clauses",
			query:    `{"logic":"and","clauses":[]}`,
			wantBody: `{"code":400,"message":"'return_query' must contain at least one clause"}`,
		},
		{
			name: "numeric operator with string type",
			query: `{"logic":"and","clauses":[{"scope":"key","key":"an-object-key",` +
				`"path":["duration"],"operator":"gt","value":"10","value_type":"string"}]}`,
			wantBody: `{"code":400,"message":"invalid return query clause 1: numeric operators require an int or float value type"}`,
		},
		{
			name: "invalid regular expression",
			query: `{"logic":"and","clauses":[{"scope":"root","operator":"regex",` +
				`"value":"[","value_type":"string"}]}`,
			wantBody: `{"code":400,"message":"invalid return query clause 1: invalid regular expression value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installSaltReturnMockDatabase(t, true)
			params := url.Values{}
			params.Set("return_query", tt.query)
			response := serveSaltReturnRequest("/salt_return?" + params.Encode())

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetSaltReturnsCombinesAdvancedClausesWithOr(t *testing.T) {
	_, mock := installSaltReturnMockDatabase(t, true)
	alterTime := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	query := `{"logic":"or","clauses":[` +
		`{"scope":"any_key","path":["duration"],"operator":"gt","value":"10","value_type":"float"},` +
		`{"scope":"any_key","path":["result"],"operator":"eq","value":"false","value_type":"bool"}` +
		`]}`
	countQuery := `SELECT count\(\*\) FROM "salt_returns" WHERE .*CAST\(\$2 AS numeric\).* OR ` +
		`.*CAST\(\$4 AS jsonb\).* AND jid = \$5`
	selectionQuery := `SELECT "fun","jid","id","success","alter_time" FROM "salt_returns" WHERE ` +
		`.*CAST\(\$2 AS numeric\).* OR .*CAST\(\$4 AS jsonb\).* AND jid = \$5 ` +
		`ORDER BY alter_time desc LIMIT \$6`

	mock.ExpectQuery(countQuery).
		WithArgs("duration", "10", "result", "false", "20260828120000000000").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(selectionQuery).
		WithArgs("duration", "10", "result", "false", "20260828120000000000", 50).
		WillReturnRows(sqlmock.NewRows([]string{"fun", "jid", "id", "success", "alter_time"}).
			AddRow("state.highstate", "20260828120000000000", "minion.example.com", false, alterTime))

	params := url.Values{}
	params.Set("jid", "20260828120000000000")
	params.Set("return_query", query)
	response := serveSaltReturnRequest("/salt_return?" + params.Encode())

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func installSaltReturnMockDatabase(t *testing.T, jsonbEnabled bool) (*gorm.DB, sqlmock.Sqlmock) {
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
	SetOptions(config.SaltDBTables{SaltReturns: "salt_returns", UseJSONB: jsonbEnabled})
	t.Cleanup(func() {
		db.DB = previousDB
		table = previousTable
		useJSONB = previousUseJSONB
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})
	return gormDB, mock
}

func serveSaltReturnRequest(requestURL string) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/salt_return", GetSaltReturns)
	request := httptest.NewRequest(http.MethodGet, requestURL, nil)
	request.Host = "example.com"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
