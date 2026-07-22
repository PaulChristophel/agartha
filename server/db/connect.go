package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/PaulChristophel/agartha/server/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Declare the variable for the database
var DB *gorm.DB

var openDatabase = func(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn))
}

var retryDelay = time.Sleep

func ConnectToDatabase(options config.DBOptions) error {
	// Connection URL to connect to Postgres Database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", options.Host, options.Port, options.User, options.Password, options.DBName, options.SSLMode)

	backoff := options.RetryInitialBackoff
	var lastErr error
	for attempt := 1; attempt <= options.RetryAttempts; attempt++ {
		candidate, err := openDatabase(dsn)
		if err == nil {
			DB = candidate
			log.Printf("Connection Opened to Database postgres://%s:***@%s:%d/%s?sslmode=%s", options.User, options.Host, options.Port, options.DBName, options.SSLMode)
			return nil
		}
		lastErr = err
		if attempt == options.RetryAttempts {
			break
		}
		log.Printf("Failed to connect to database. Attempt %d/%d. Retrying in %s...", attempt, options.RetryAttempts, backoff)
		retryDelay(backoff)
		backoff = nextBackoff(backoff, options.RetryMaxBackoff, options.RetryMultiplier)
	}

	return fmt.Errorf("failed to connect to database after %d attempts: %w", options.RetryAttempts, lastErr)
}

func nextBackoff(current, maximum time.Duration, multiplier float64) time.Duration {
	next := time.Duration(float64(current) * multiplier)
	if next <= 0 || next > maximum {
		return maximum
	}
	return next
}

// Ready verifies that the configured database handle can serve requests.
func Ready(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database is not initialized")
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("get database handle: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}
