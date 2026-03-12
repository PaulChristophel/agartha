package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

func runSQLMigrations() (err error) {
	if config.AgarthaConfig == nil {
		return errors.New("config not initialized")
	}

	dbOptions := config.AgarthaConfig.DB
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbOptions.Host,
		dbOptions.Port,
		dbOptions.User,
		dbOptions.Password,
		dbOptions.DBName,
		dbOptions.SSLMode,
	)

	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}

	cleanupNeeded := true
	defer func() {
		if !cleanupNeeded {
			return
		}
		closeErr := sqlDB.Close()
		if closeErr == nil {
			return
		}
		if err == nil {
			err = closeErr
			return
		}
		err = errors.Join(err, closeErr)
	}()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return err
	}

	cleanupNeeded = false
	defer func() {
		sourceErr, dbErr := m.Close()
		closeErr := errors.Join(sourceErr, dbErr)
		if closeErr == nil {
			return
		}
		if err == nil {
			err = closeErr
			return
		}
		err = errors.Join(err, closeErr)
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
