package db

import (
	"fmt"
	"log"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Declare the variable for the database
var DB *gorm.DB

func ConnectToDatabase(options config.DBOptions) {
	var err error

	log := logger.GetLogger()

	// Connection URL to connect to Postgres Database
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		options.Host,
		options.Port,
		options.User,
		options.Password,
		options.DBName,
		options.SSLMode,
	)
	// Retry logic
	for i := range 3 {
		// Connect to the DB and initialize the DB variable
		DB, err = gorm.Open(postgres.Open(dsn))
		if err == nil {
			log.Info(
				"Connection opened to database",
				zap.String("user", options.User),
				zap.String("host", options.Host),
				zap.Int("port", options.Port),
				zap.String("dbname", options.DBName),
				zap.String("sslmode", options.SSLMode),
			)
			return
		}
		log.Warn(
			"Failed to connect to database. Retrying",
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", 3),
			zap.Error(err),
		)
		time.Sleep(5 * time.Second)
	}

	// If still unable to connect after 3 attempts, panic
	log.Fatal(
		"Failed to connect to database after max attempts",
		zap.Int("max_attempts", 3),
		zap.Error(err),
	)
}
