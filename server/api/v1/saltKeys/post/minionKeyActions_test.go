package saltKeys

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestAcceptAndRejectMinionKeyHandlers(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		state   string
		handler gin.HandlerFunc
	}{
		{name: "accept", path: "/accept", state: "accepted", handler: AcceptMinionKeys},
		{name: "reject", path: "/reject", state: "rejected", handler: RejectMinionKeys},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := installPostMockDatabase(t)
			expectPostSaltKeysTable(mock, 1)
			mock.ExpectBegin()
			mock.ExpectExec(`UPDATE "salt_keys".*SET data = jsonb_set.*WHERE bank = \$1 AND psql_key IN \(\$2,\$3\)`).
				WithArgs(keysBank, "minion-1", "minion-2").
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(`INSERT INTO "salt_keys".*FROM "salt_keys".*WHERE bank = \$3 AND psql_key IN \(\$4,\$5\).*ON CONFLICT`).
				WithArgs(keysBank, tt.state, deniedBank, "minion-1", "minion-2").
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectExec(`DELETE FROM "salt_keys".*WHERE bank = \$1 AND psql_key IN \(\$2,\$3\)`).
				WithArgs(deniedBank, "minion-1", "minion-2").
				WillReturnResult(sqlmock.NewResult(0, 1))
			mock.ExpectCommit()

			response := servePostRequest(tt.path, `{"match":{"minions":["minion-1"],"minions_denied":["minion-2","minion-1"]}}`, tt.handler)

			require.Equal(t, http.StatusOK, response.Code)
			require.JSONEq(t, `{"minions":["minion-1","minion-2"]}`, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteMinionKeysHandlerPreservesStateBuckets(t *testing.T) {
	mock := installPostMockDatabase(t)
	expectPostSaltKeysTable(mock, 1)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "salt_keys".*WHERE bank = \$1 AND psql_key IN \(\$2,\$3,\$4\)`).
		WithArgs(keysBank, "accepted", "pending", "rejected").
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec(`DELETE FROM "salt_keys".*WHERE bank = \$1 AND psql_key IN \(\$2\)`).
		WithArgs(deniedBank, "denied").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	response := servePostRequest("/delete", `{
		"match": {
			"minions": ["accepted"],
			"minions_denied": ["denied"],
			"minions_pre": ["pending"],
			"minions_rejected": ["rejected"]
		}
	}`, DeleteMinionKeys)

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"minions":["accepted","denied","pending","rejected"]}`, response.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestMinionKeyActionHandlersReturnErrors(t *testing.T) {
	t.Run("malformed request", func(t *testing.T) {
		response := servePostRequest("/accept", `{`, AcceptMinionKeys)
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("table unavailable", func(t *testing.T) {
		mock := installPostMockDatabase(t)
		expectPostSaltKeysTable(mock, 0)
		response := servePostRequest("/accept", `{"match":{"minions":["minion-1"]}}`, AcceptMinionKeys)

		require.Equal(t, http.StatusNotFound, response.Code)
		require.JSONEq(t, `{"code":404,"message":"salt_keys table is unavailable"}`, response.Body.String())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("transition failure rolls back", func(t *testing.T) {
		mock := installPostMockDatabase(t)
		expectPostSaltKeysTable(mock, 1)
		mock.ExpectBegin()
		mock.ExpectExec(`UPDATE "salt_keys"`).WillReturnError(errors.New("update failed"))
		mock.ExpectRollback()

		response := servePostRequest("/reject", `{"match":{"minions":["minion-1"]}}`, RejectMinionKeys)

		require.Equal(t, http.StatusInternalServerError, response.Code)
		require.JSONEq(t, `{"code":500,"message":"update failed"}`, response.Body.String())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUniqueStrings(t *testing.T) {
	require.Equal(t, []string{"one", "two"}, uniqueStrings([]string{"one", "", "two", "one"}))
}

func installPostMockDatabase(t *testing.T) sqlmock.Sqlmock {
	t.Helper()
	gin.SetMode(gin.TestMode)
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	require.NoError(t, err)

	previousDB := db.DB
	db.DB = gormDB
	SetOptions(config.SaltDBTables{SaltKeys: "salt_keys"})
	t.Cleanup(func() {
		db.DB = previousDB
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})
	return mock
}

func expectPostSaltKeysTable(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2`)).
		WithArgs("salt_keys", "BASE TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

func servePostRequest(path, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	router := gin.New()
	router.POST(path, handler)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
