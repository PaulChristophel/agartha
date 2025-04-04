package db

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/PaulChristophel/agartha/server/config"
	"github.com/stretchr/testify/assert"
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
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() {
		if derr := db.Close(); derr != nil {
			// Optionally log the error
			log.Printf("failed to close db connection: %v", derr)
		}
	}()

	// Set the global DB variable to the mock database
	DB, err = gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	assert.NoError(t, err)

	// Test the function
	ConnectToDatabase(options)

	// Ensure all expectations are met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
