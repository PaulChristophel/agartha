package db

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestConnectToDatabase(t *testing.T) {
	options := config.DBOptions{
		Host:     "db",
		Port:     5432,
		User:     "agartha",
		Password: "agartha",
		DBName:   "agartha",
		SSLMode:  "disable",
	}

	// Create a new mock database
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() {
		mock.ExpectClose()
		require.NoError(t, sqlDB.Close())
	})

	// Set the global DB variable to the mock database
	mockDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	originalOpenDatabase := openDatabase
	originalRetryDelay := retryDelay
	openDatabase = func(string) (*gorm.DB, error) { return mockDB, nil }
	retryDelay = func(time.Duration) {}
	t.Cleanup(func() {
		openDatabase = originalOpenDatabase
		retryDelay = originalRetryDelay
	})

	// Test the function
	ConnectToDatabase(options)

	// Ensure all expectations are met
	require.Same(t, mockDB, DB)
	require.NoError(t, mock.ExpectationsWereMet())
}
