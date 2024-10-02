package db

import (
	"fmt"
	"log"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Declare the variable for the database
var DB *gorm.DB

func ConnectToDatabase(options config.DBOptions) {
	var err error
	// Connection URL to connect to Postgres Database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", options.Host, options.Port, options.User, options.Password, options.DBName, options.SSLMode)

	// Retry logic
	for i := 0; i < 3; i++ {
		// Connect to the DB and initialize the DB variable
		DB, err = gorm.Open(postgres.Open(dsn))
		if err == nil {
			log.Printf("Connection Opened to Database postgres://%s:***@%s:%d/%s?sslmode=%s", options.User, options.Host, options.Port, options.DBName, options.SSLMode)
			return
		}
		log.Printf("Failed to connect to database. Attempt %d/3. Retrying in 5 seconds...", i+1)
		time.Sleep(5 * time.Second)
	}

	// If still unable to connect after 3 attempts, panic
	panic(fmt.Sprintf("failed to connect to database after 3 attempts: %v", err))
}
