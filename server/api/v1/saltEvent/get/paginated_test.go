package saltEvent

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
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

func TestGetSaltEventsValidatesPaginationAndFilters(t *testing.T) {
	tests := []struct {
		name string
		url  string
		body string
	}{
		{name: "invalid page", url: "/salt_event?page=zero", body: `{"code":400,"message":"invalid page parameter"}`},
		{name: "zero page", url: "/salt_event?page=0", body: `{"code":400,"message":"invalid page parameter"}`},
		{name: "negative per page", url: "/salt_event?per_page=-1", body: `{"code":400,"message":"invalid per_page parameter"}`},
		{name: "invalid since", url: "/salt_event?since=yesterday", body: `{"code":400,"message":"invalid 'since' date format"}`},
		{name: "invalid until", url: "/salt_event?until=tomorrow", body: `{"code":400,"message":"invalid 'until' date format"}`},
		{name: "invalid order", url: "/salt_event?order_by=data", body: `{"code":400,"message":"invalid column name 'data'. Valid columns: [id tag alter_time master_id]"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installMockDatabase(t)
			response := serveSaltEventRequest(tt.url)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.JSONEq(t, tt.body, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetSaltEventsBuildsFilteredPaginatedQuery(t *testing.T) {
	_, mock := installMockDatabase(t)
	since := "2026-08-01T00:00:00Z"
	until := "2026-08-02T00:00:00Z"
	alterTime, err := time.Parse(time.RFC3339, "2026-08-01T12:00:00Z")
	require.NoError(t, err)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "salt_events" WHERE tag LIKE $1 AND master_id LIKE $2 AND alter_time >= $3 AND alter_time <= $4`)).
		WithArgs("salt/%", `master\_%`, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(`SELECT .* FROM "salt_events" WHERE tag LIKE \$1 AND master_id LIKE \$2 AND alter_time >= \$3 AND alter_time <= \$4 ORDER BY master_id desc LIMIT \$5 OFFSET \$6`).
		WithArgs("salt/%", `master\_%`, sqlmock.AnyArg(), sqlmock.AnyArg(), 2, 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "tag", "alter_time", "master_id"}).
			AddRow(3, "salt/job/3", alterTime, "master_1"))

	response := serveSaltEventRequest("/salt_event?tag=salt/*&master_id=master_*&since=" + since + "&until=" + until + "&page=2&per_page=2&order_by=master_id%20desc")

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{
		"paging": {
			"per_page": 2,
			"num_pages": 2,
			"count": 3,
			"next": "",
			"previous": "http://example.com/salt_event?page=1&per_page=2"
		},
		"results": [{
			"id": 3,
			"tag": "salt/job/3",
			"data": null,
			"alter_time": "2026-08-01T12:00:00Z",
			"master_id": "master_1"
		}]
	}`, response.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetSaltEventsReturnsDatabaseErrors(t *testing.T) {
	tests := []struct {
		name       string
		expect     func(sqlmock.Sqlmock)
		wantBody   string
		wantStatus int
	}{
		{
			name: "count failure",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "salt_events"`).WillReturnError(errors.New("count failed"))
			},
			wantBody:   `{"code":500,"message":"Failed to count salt events."}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "select failure",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT count\(\*\) FROM "salt_events"`).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				mock.ExpectQuery(`SELECT .* FROM "salt_events"`).WillReturnError(errors.New("select failed"))
			},
			wantBody:   `{"code":500,"message":"Failed to retrieve salt events."}`,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, mock := installMockDatabase(t)
			tt.expect(mock)

			response := serveSaltEventRequest("/salt_event?since=2026-08-01T00:00:00Z")

			require.Equal(t, tt.wantStatus, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func installMockDatabase(t *testing.T, jsonbEnabled ...bool) (*gorm.DB, sqlmock.Sqlmock) {
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
	useJSONBForTest := true
	if len(jsonbEnabled) > 0 {
		useJSONBForTest = jsonbEnabled[0]
	}
	SetOptions(config.SaltDBTables{SaltEvents: "salt_events", UseJSONB: useJSONBForTest})
	t.Cleanup(func() {
		db.DB = previousDB
		table = previousTable
		useJSONB = previousUseJSONB
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})
	return gormDB, mock
}

func serveSaltEventRequest(url string) *httptest.ResponseRecorder {
	router := gin.New()
	router.GET("/salt_event", GetSaltEvents)
	request := httptest.NewRequest(http.MethodGet, url, nil)
	request.Host = "example.com"
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
