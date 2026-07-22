package model

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	return db, mock
}

func testAuthUser() AuthUser {
	return AuthUser{
		Password:    "plain_password",
		IsSuperuser: false,
		Username:    "testuser",
		FirstName:   "Test",
		LastName:    "User",
		Email:       "testuser@example.com",
		IsStaff:     false,
		IsActive:    true,
		DateJoined:  time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC),
	}
}

func TestAuthUserCreate(t *testing.T) {
	db, mock := setupTestDB(t)
	user := testAuthUser()

	mock.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO auth_user (password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined)
		VALUES (crypt($1, gen_salt('bf', 8)), $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`)).
		WithArgs(user.Password, user.LastLogin, user.IsSuperuser, user.Username, user.FirstName, user.LastName, user.Email, user.IsStaff, user.IsActive, user.DateJoined).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))

	require.NoError(t, user.Create(db))
	require.Equal(t, uint(42), user.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthUserDelete(t *testing.T) {
	db, mock := setupTestDB(t)
	user := testAuthUser()
	user.ID = 42

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "auth_user" WHERE "auth_user"."id" = $1`)).
		WithArgs(user.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, user.Delete(db, user.ID))
	require.NoError(t, mock.ExpectationsWereMet())
}
