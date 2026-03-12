package db

import (
	"fmt"
	"strings"
)

func ensurePgcryptoExtension() error {
	exec := DB.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	return exec.Error
}

func ensureSaltReturnsIndex(table string) error {
	query := fmt.Sprintf(
		"CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_returns_fun_fullret ON %s (id, jid DESC) WHERE fun IN ('state.highstate', 'state.apply') AND (full_ret::jsonb->>'fun_args' = '[]');",
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
    CREATE TRIGGER trigger_alter_time
    BEFORE UPDATE ON %s
    FOR EACH ROW
    WHEN (OLD.data IS DISTINCT FROM NEW.data)
    EXECUTE PROCEDURE alter_time();`, table)
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
    FROM %s a
    WHERE
        (a.fun::text = 'state.highstate'::text OR a.fun::text = 'state.apply'::text) AND
        (a.full_ret::jsonb ->> 'fun_args'::text) = '[]'::text`, saltReturnsTable)
	return DB.Exec(query).Error
}
