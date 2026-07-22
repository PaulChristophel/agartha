package db

import (
	"errors"
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
		Host:                "db",
		Port:                5432,
		User:                "agartha",
		Password:            "agartha",
		DBName:              "agartha",
		SSLMode:             "disable",
		RetryAttempts:       3,
		RetryInitialBackoff: time.Second,
		RetryMaxBackoff:     5 * time.Second,
		RetryMultiplier:     2,
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
	require.NoError(t, ConnectToDatabase(options))

	// Ensure all expectations are met
	require.Same(t, mockDB, DB)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConnectToDatabaseReturnsErrorAfterConfiguredBackoff(t *testing.T) {
	options := config.DBOptions{
		Host:                "db",
		Port:                5432,
		User:                "agartha",
		Password:            "not-a-placeholder",
		DBName:              "agartha",
		SSLMode:             "require",
		RetryAttempts:       5,
		RetryInitialBackoff: time.Second,
		RetryMaxBackoff:     3 * time.Second,
		RetryMultiplier:     2,
	}

	originalDB := DB
	originalOpenDatabase := openDatabase
	originalRetryDelay := retryDelay
	DB = nil
	t.Cleanup(func() {
		DB = originalDB
		openDatabase = originalOpenDatabase
		retryDelay = originalRetryDelay
	})

	attempts := 0
	openDatabase = func(string) (*gorm.DB, error) {
		attempts++
		return nil, errors.New("database unavailable")
	}
	var delays []time.Duration
	retryDelay = func(delay time.Duration) { delays = append(delays, delay) }

	err := ConnectToDatabase(options)
	require.ErrorContains(t, err, "failed to connect to database after 5 attempts")
	require.Equal(t, 5, attempts)
	require.Equal(t, []time.Duration{time.Second, 2 * time.Second, 3 * time.Second, 3 * time.Second}, delays)
	require.Nil(t, DB)
}

func TestReadyPingsDatabase(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.ExpectPing()
	mock.ExpectPing()
	mock.ExpectClose()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)
	originalDB := DB
	DB = gormDB
	t.Cleanup(func() { DB = originalDB })

	require.NoError(t, Ready(t.Context()))
	require.NoError(t, sqlDB.Close())
	require.NoError(t, mock.ExpectationsWereMet())
}
