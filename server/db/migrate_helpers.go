package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/PaulChristophel/agartha/server/config"
)

const (
	migrationAdvisoryLockKey1 = 1828852296
	migrationAdvisoryLockKey2 = 1752396916
)

func withMigrationLock(fn func() error) (err error) {
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
	defer func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	ctx := context.Background()
	conn, err := sqlDB.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	if err := acquireMigrationLock(ctx, conn); err != nil {
		return err
	}
	defer func() {
		if unlockErr := releaseMigrationLock(ctx, conn); unlockErr != nil {
			err = errors.Join(err, unlockErr)
		}
	}()

	return fn()
}

func acquireMigrationLock(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(
		ctx,
		"SELECT pg_advisory_lock($1, $2);",
		migrationAdvisoryLockKey1,
		migrationAdvisoryLockKey2,
	)
	return err
}

func releaseMigrationLock(ctx context.Context, conn *sql.Conn) error {
	_, err := conn.ExecContext(
		ctx,
		"SELECT pg_advisory_unlock($1, $2);",
		migrationAdvisoryLockKey1,
		migrationAdvisoryLockKey2,
	)
	return err
}

func ensurePgcryptoExtension() error {
	exec := DB.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	return exec.Error
}

func ensureSaltReturnsIndex(table string) error {
	query := fmt.Sprintf(
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_returns_fun_fullret ON %s (id, jid DESC) WHERE fun IN ('state.highstate', 'state.apply') AND POSITION($nul$\\u0000$nul$ IN full_ret::text) = 0 AND (full_ret::jsonb->>'fun_args' = '[]');",
		table,
	)
	return DB.Exec(query).Error
}

func ensureSaltCacheIndex(table string) error {
	query := fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_cache_bank_substring ON %s (substring(bank from 9));", table)
	return DB.Exec(query).Error
}

func ensureAlterTimeFunction() error {
	const query = `
    CREATE OR REPLACE FUNCTION alter_time() RETURNS trigger
        LANGUAGE plpgsql
        AS $$
    BEGIN
    NEW.alter_time := current_timestamp;
    RETURN NEW;
    END;
    $$;`
	return DB.Exec(query).Error
}

func ensureAlterTimeTrigger(table string) error {
	query := fmt.Sprintf(`
    DROP TRIGGER IF EXISTS trigger_alter_time ON %s;
    CREATE TRIGGER trigger_alter_time
    BEFORE UPDATE ON %s
    FOR EACH ROW
    WHEN (OLD.data IS DISTINCT FROM NEW.data)
    EXECUTE PROCEDURE alter_time();`, table, table)
	err := DB.Exec(query).Error
	if err != nil && !strings.Contains(err.Error(), "SQLSTATE 42710") {
		return err
	}
	return nil
}

func ensureSaltHighstatesView(saltReturnsTable string) error {
	query := fmt.Sprintf(`
    CREATE OR REPLACE VIEW vw_salt_highstates AS
    SELECT
        distinct on(a.id)
        a.fun,
        a.jid,
        a.return::jsonb,
        a.full_ret::jsonb,
        a.id,
        a.success::boolean,
        a.alter_time
    FROM (
        SELECT *
        FROM %s
        WHERE POSITION($nul$\u0000$nul$ IN return::text) = 0
            AND POSITION($nul$\u0000$nul$ IN full_ret::text) = 0
    ) a
    WHERE
        (a.fun::text = 'state.highstate'::text OR a.fun::text = 'state.apply'::text) AND
        (a.full_ret::jsonb ->> 'fun_args'::text) = '[]'::text`, saltReturnsTable)
	return DB.Exec(query).Error
}
