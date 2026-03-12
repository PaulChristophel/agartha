package db

import (
	"log"

	"github.com/PaulChristophel/agartha/server/config"
	agartha "github.com/PaulChristophel/agartha/server/model/agartha"
	salt "github.com/PaulChristophel/agartha/server/model/salt"
)

func Migrate(options config.SaltDBTables) error {
	if err := ensurePgcryptoExtension(); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	err := DB.Table(options.JIDs).AutoMigrate(&salt.JID{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Not useful and resource intensive
	// exec = DB.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_jids_load ON %s USING gin (to_tsvector('english', load))", options.JIDs))
	// if exec.Error != nil {
	// 	log.Printf("Error during migration: %v", exec.Error)
	// 	return exec.Error
	// }

	// LARGE and potentially not useful
	// exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_jids_load_jsonb ON jids USING gin (("load"::jsonb)) WITH (fastupdate=ON)`)
	// if exec.Error != nil {
	// 	log.Printf("Error during migration: %v", exec.Error)
	// 	return exec.Error
	// }

	// Configure SaltReturns
	err = DB.Table(options.SaltReturns).AutoMigrate(&salt.SaltReturn{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}
	if err := ensureSaltReturnsIndex(options.SaltReturns); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// LARGE and potentially not useful
	// exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_returns_return ON salt_returns USING gin (("return"::jsonb)) WITH (fastupdate=ON);`)
	// if exec.Error != nil {
	// 	log.Printf("Error during migration: %v", exec.Error)
	// 	return exec.Error
	// }

	// exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_returns_full_ret ON salt_returns USING gin ((full_ret::jsonb)) WITH (fastupdate=ON);`)
	// if exec.Error != nil {
	// 	log.Printf("Error during migration: %v", exec.Error)
	// 	return exec.Error
	// }

	// Configure SaltEvents
	err = DB.Table(options.SaltEvents).AutoMigrate(&salt.SaltEvent{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Configure SaltCache
	err = DB.Table(options.SaltCache).AutoMigrate(&salt.SaltCache{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}
	if err := ensureSaltCacheIndex(options.SaltCache); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	if err := ensureAlterTimeFunction(); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	if err := ensureAlterTimeTrigger(options.SaltCache); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Configure AuthUser
	err = DB.AutoMigrate(&agartha.AuthUser{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Configure JobTemplates
	err = DB.AutoMigrate(&agartha.JobTemplate{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Configure UserSettings
	err = DB.AutoMigrate(&agartha.UserSettings{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Configure Session
	err = DB.AutoMigrate(&agartha.Session{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	// Configure SessionUserMap
	err = DB.AutoMigrate(&agartha.SessionUserMap{})
	if err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	if err := ensureSaltHighstatesView(options.SaltReturns); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	if err := runSQLMigrations(); err != nil {
		log.Printf("Error during migration: %v", err)
		return err
	}

	log.Printf("Database Migrated")
	return nil
}
