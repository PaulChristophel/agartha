package saltKeys

import (
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

func TestDeleteSaltKeysBankKey(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		tableCount int
		deleteErr  error
		rows       int64
		wantStatus int
		wantBody   string
	}{
		{
			name:       "deletes nested bank and key",
			path:       "/salt_keys/pki/master/keys/minion.example.com",
			tableCount: 1,
			rows:       1,
			wantStatus: http.StatusOK,
			wantBody:   `{"code":200,"message":"Deleted salt_keys item pki/master/keys/minion.example.com"}`,
		},
		{
			name:       "missing key",
			path:       "/salt_keys/pki/master/keys/missing.example.com",
			tableCount: 1,
			rows:       0,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"code":404,"message":"No salt_keys item found to delete."}`,
		},
		{
			name:       "database failure",
			path:       "/salt_keys/pki/master/keys/minion.example.com",
			tableCount: 1,
			deleteErr:  errors.New("delete failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"code":500,"message":"Failed to delete salt_keys item."}`,
		},
		{
			name:       "table unavailable",
			path:       "/salt_keys/pki/master/keys/minion.example.com",
			tableCount: 0,
			wantStatus: http.StatusNotFound,
			wantBody:   `{"code":404,"message":"salt_keys table is unavailable"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := installDeleteMockDatabase(t)
			expectSaltKeysTable(mock, tt.tableCount)
			if tt.tableCount > 0 {
				mock.ExpectBegin()
				expectation := mock.ExpectExec(`DELETE FROM "salt_keys".*bank = \$1 AND psql_key = \$2`).
					WithArgs("pki/master/keys", pathKey(tt.path))
				if tt.deleteErr != nil {
					expectation.WillReturnError(tt.deleteErr)
					mock.ExpectRollback()
				} else {
					expectation.WillReturnResult(sqlmock.NewResult(0, tt.rows))
					mock.ExpectCommit()
				}
			}

			response := serveDeleteRequest(tt.path)

			require.Equal(t, tt.wantStatus, response.Code)
			require.JSONEq(t, tt.wantBody, response.Body.String())
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteSaltKeysBank(t *testing.T) {
	mock := installDeleteMockDatabase(t)
	expectSaltKeysTable(mock, 1)
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "salt_keys".*bank = \$1`).
		WithArgs("/pki/master/keys").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	router := gin.New()
	router.DELETE("/salt_keys/bank/*bank", DeleteSaltKeysBank)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/salt_keys/bank/pki/master/keys", nil))

	require.Equal(t, http.StatusOK, response.Code)
	require.JSONEq(t, `{"code":200,"message":"Deleted all salt_keys items in bank: /pki/master/keys"}`, response.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSplitBankAndKey(t *testing.T) {
	tests := []struct {
		name     string
		bank     string
		rest     string
		wantBank string
		wantKey  string
		wantErr  bool
	}{
		{name: "simple", bank: "pki", rest: "/master", wantBank: "pki", wantKey: "master"},
		{name: "nested", bank: "pki", rest: "/master/keys/minion", wantBank: "pki/master/keys", wantKey: "minion"},
		{name: "escaped", bank: "pki%2Fmaster", rest: "/keys%2Fminion", wantBank: "pki/master/keys", wantKey: "minion"},
		{name: "empty", bank: "pki", rest: "/", wantBank: "", wantKey: ""},
		{name: "invalid escape", bank: "%zz", rest: "/key", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bank, key, err := splitBankAndKey(tt.bank, tt.rest)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantBank, bank)
			require.Equal(t, tt.wantKey, key)
		})
	}
}

func installDeleteMockDatabase(t *testing.T) sqlmock.Sqlmock {
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

func expectSaltKeysTable(mock sqlmock.Sqlmock, count int) {
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2`)).
		WithArgs("salt_keys", "BASE TABLE").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
}

func serveDeleteRequest(path string) *httptest.ResponseRecorder {
	router := gin.New()
	router.DELETE("/salt_keys/:bank/*key", DeleteSaltKeysBankKey)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, path, nil))
	return response
}

func pathKey(path string) string {
	const prefix = "/salt_keys/pki/master/keys/"
	return path[len(prefix):]
}
