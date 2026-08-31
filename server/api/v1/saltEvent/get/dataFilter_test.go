package saltEvent

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestGetSaltEventsValidatesDataFilter(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		useJSONB   bool
		wantBody   string
		wantStatus int
	}{
		{
			name:       "match without filter",
			query:      "data_match=key_exists",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"'data_filter' is required for key_exists"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "filter without match",
			query:      "data_filter=minion_id",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"'data_match' is required when data filter parameters are provided"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unknown match",
			query:      "data_match=contains&data_filter=failed",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"invalid 'data_match' value: must be key_exists, string_equals, or key_field_equals"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "legacy text tables",
			query:      "data_match=string_equals&data_filter=failed",
			useJSONB:   false,
			wantBody:   `{"code":400,"message":"data filtering requires JSONB database tables"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid nested boolean",
			query:      "data_match=key_field_equals&data_key=return&data_field=result&data_value=nope&data_value_type=bool",
			useJSONB:   true,
			wantBody:   `{"code":400,"message":"invalid boolean 'data_value'"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installMockDatabase(t, tt.useJSONB)
			response := serveSaltEventRequest("/salt_event?" + tt.query)

			require.Equal(t, tt.wantStatus, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetSaltEventsAppliesTypedJSONBFilters(t *testing.T) {
	tests := []struct {
		name           string
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
			match:       dataMatchKeyExists,
			filter:      "minion_id",
			expectedArg: "minion_id",
			countQuery: `SELECT count\(\*\) FROM "salt_events" WHERE \(jsonb_typeof\("data"\) = 'object' ` +
				`AND jsonb_exists\("data", \$1\)\) AND tag = \$2 AND alter_time >= \$3`,
			selectionQuery: `SELECT "id","tag","alter_time","master_id" FROM "salt_events" WHERE ` +
				`\(jsonb_typeof\("data"\) = 'object' AND jsonb_exists\("data", \$1\)\) ` +
				`AND tag = \$2 AND alter_time >= \$3 ORDER BY id desc LIMIT \$4`,
		},
		{
			name:        "root string equals",
			match:       dataMatchStringEquals,
			filter:      "Unhandled exception",
			expectedArg: "Unhandled exception",
			countQuery: `SELECT count\(\*\) FROM "salt_events" WHERE \(jsonb_typeof\("data"\) = 'string' ` +
				`AND "data" = to_jsonb\(CAST\(\$1 AS text\)\)\) AND tag = \$2 AND alter_time >= \$3`,
			selectionQuery: `SELECT "id","tag","alter_time","master_id" FROM "salt_events" WHERE ` +
				`\(jsonb_typeof\("data"\) = 'string' AND "data" = to_jsonb\(CAST\(\$1 AS text\)\)\) ` +
				`AND tag = \$2 AND alter_time >= \$3 ORDER BY id desc LIMIT \$4`,
		},
		{
			name:        "field below top-level key equals boolean",
			match:       dataMatchKeyFieldEquals,
			key:         "return",
			field:       "result",
			value:       "false",
			valueType:   "bool",
			expectedArg: `{"return":{"result":false}}`,
			countQuery: `SELECT count\(\*\) FROM "salt_events" WHERE \(jsonb_typeof\("data"\) = 'object' ` +
				`AND "data" @> CAST\(\$1 AS jsonb\)\) AND tag = \$2 AND alter_time >= \$3`,
			selectionQuery: `SELECT "id","tag","alter_time","master_id" FROM "salt_events" WHERE ` +
				`\(jsonb_typeof\("data"\) = 'object' AND "data" @> CAST\(\$1 AS jsonb\)\) ` +
				`AND tag = \$2 AND alter_time >= \$3 ORDER BY id desc LIMIT \$4`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installMockDatabase(t, true)
			alterTime := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)

			mock.ExpectQuery(tt.countQuery).
				WithArgs(tt.expectedArg, "salt/job/complete", sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			mock.ExpectQuery(tt.selectionQuery).
				WithArgs(tt.expectedArg, "salt/job/complete", sqlmock.AnyArg(), 50).
				WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "alter_time", "master_id"}).
					AddRow(42, "salt/job/complete", alterTime, "master_1"))

			params := url.Values{}
			params.Set("tag", "salt/job/complete")
			params.Set("data_match", tt.match)
			if tt.filter != "" {
				params.Set("data_filter", tt.filter)
			}
			if tt.key != "" {
				params.Set("data_key", tt.key)
				params.Set("data_field", tt.field)
				params.Set("data_value", tt.value)
				params.Set("data_value_type", tt.valueType)
			}
			params.Set("since", "2026-08-31T00:00:00Z")
			response := serveSaltEventRequest("/salt_event?" + params.Encode())

			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.JSONEq(t, `{
				"paging":{"per_page":50,"num_pages":1,"count":1,"next":"","previous":""},
				"results":[{"id":42,"tag":"salt/job/complete","data":null,
				"alter_time":"2026-08-31T12:00:00Z","master_id":"master_1"}]
			}`, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestBuildDataQueryClause(t *testing.T) {
	tests := []struct {
		name          string
		clause        dataQueryClause
		wantPredicate string
		wantArgs      []any
	}{
		{
			name: "specific key numeric greater than",
			clause: dataQueryClause{
				Scope: "key", Key: "payload", Path: []string{"duration"},
				Operator: "gt", Value: "10", ValueType: "float",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path("data", ?, ?) AS value) AS data_target ` +
				`WHERE jsonb_typeof(data_target.value) = 'number' AND (data_target.value #>> '{}')::numeric > CAST(? AS numeric))`,
			wantArgs: []any{"payload", "duration", "10"},
		},
		{
			name: "failed state under highstate return object",
			clause: dataQueryClause{
				Scope: "any_key", ContainerPath: []string{"return"}, Path: []string{"result"},
				Operator: "eq", Value: "false", ValueType: "bool",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path("data", ?) AS value) AS data_container ` +
				`CROSS JOIN LATERAL jsonb_each(CASE WHEN jsonb_typeof(data_container.value) = 'object' ` +
				`THEN data_container.value ELSE '{}'::jsonb END) AS data_entry(key, value) WHERE ` +
				`EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path(data_entry.value, ?) AS value) AS data_target ` +
				`WHERE data_target.value = CAST(? AS jsonb)))`,
			wantArgs: []any{"return", "result", "false"},
		},
		{
			name: "root regular expression",
			clause: dataQueryClause{
				Scope: "root", Path: []string{"message"}, Operator: "regex", Value: "(?i)exception", ValueType: "string",
			},
			wantPredicate: `EXISTS (SELECT 1 FROM (SELECT jsonb_extract_path("data", ?) AS value) AS data_target ` +
				`WHERE jsonb_typeof(data_target.value) = 'string' AND (data_target.value #>> '{}') ~ ?)`,
			wantArgs: []any{"message", "(?i)exception"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			predicate, args, err := buildDataQueryClause(tt.clause)
			require.NoError(t, err)
			require.Equal(t, tt.wantPredicate, predicate)
			require.Equal(t, tt.wantArgs, args)
		})
	}
}

func TestGetSaltEventsValidatesAdvancedDataQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		wantBody string
	}{
		{name: "invalid JSON", query: `{`, wantBody: `{"code":400,"message":"invalid 'data_query' JSON"}`},
		{name: "empty clauses", query: `{"logic":"and","clauses":[]}`, wantBody: `{"code":400,"message":"'data_query' must contain at least one clause"}`},
		{
			name: "numeric operator with string type",
			query: `{"logic":"and","clauses":[{"scope":"root","path":["duration"],` +
				`"operator":"gt","value":"10","value_type":"string"}]}`,
			wantBody: `{"code":400,"message":"invalid data query clause 1: numeric operators require an int or float value type"}`,
		},
		{
			name: "invalid regular expression",
			query: `{"logic":"and","clauses":[{"scope":"root","operator":"regex",` +
				`"value":"[","value_type":"string"}]}`,
			wantBody: `{"code":400,"message":"invalid data query clause 1: invalid regular expression value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installMockDatabase(t, true)
			params := url.Values{}
			params.Set("data_query", tt.query)
			response := serveSaltEventRequest("/salt_event?" + params.Encode())

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetSaltEventsCombinesAdvancedClausesWithOr(t *testing.T) {
	_, mock := installMockDatabase(t, true)
	alterTime := time.Date(2026, time.August, 31, 12, 0, 0, 0, time.UTC)
	query := `{"logic":"or","clauses":[` +
		`{"scope":"root","path":["duration"],"operator":"gt","value":"10","value_type":"float"},` +
		`{"scope":"root","path":["result"],"operator":"eq","value":"false","value_type":"bool"}` +
		`]}`
	countQuery := `SELECT count\(\*\) FROM "salt_events" WHERE .*CAST\(\$2 AS numeric\).* OR ` +
		`.*CAST\(\$4 AS jsonb\).* AND master_id = \$5 AND alter_time >= \$6`
	selectionQuery := `SELECT "id","tag","alter_time","master_id" FROM "salt_events" WHERE ` +
		`.*CAST\(\$2 AS numeric\).* OR .*CAST\(\$4 AS jsonb\).* AND master_id = \$5 ` +
		`AND alter_time >= \$6 ORDER BY id desc LIMIT \$7`

	mock.ExpectQuery(countQuery).
		WithArgs("duration", "10", "result", "false", "master_1", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(selectionQuery).
		WithArgs("duration", "10", "result", "false", "master_1", sqlmock.AnyArg(), 50).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "alter_time", "master_id"}).
			AddRow(42, "salt/job/complete", alterTime, "master_1"))

	params := url.Values{}
	params.Set("master_id", "master_1")
	params.Set("data_query", query)
	params.Set("since", "2026-08-31T00:00:00Z")
	response := serveSaltEventRequest("/salt_event?" + params.Encode())

	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}
