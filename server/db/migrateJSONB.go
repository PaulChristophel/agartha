package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/PaulChristophel/agartha/server/config"
	agartha "github.com/PaulChristophel/agartha/server/model/agartha"
	salt "github.com/PaulChristophel/agartha/server/model/salt/jsonb"
)

func MigrateJSONB(options config.SaltDBTables) error {
	// Configure JIDs
	exec := DB.Exec(`CREATE EXTENSION IF NOT EXISTS pgcrypto;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
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
	// LARGE but critical
	exec = DB.Exec(fmt.Sprintf("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_returns_fun_fullret ON %s (id, jid DESC) WHERE fun IN ('state.highstate', 'state.apply') AND (full_ret::jsonb->>'fun_args' = '[]');", options.SaltReturns))
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
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
	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_salt_cache_bank_substring ON salt_cache (substring(bank from 9));`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE OR REPLACE FUNCTION alter_time() RETURNS trigger
        LANGUAGE plpgsql
        AS $$
    BEGIN
    NEW.alter_time := current_timestamp;
    RETURN NEW;
    END;
    $$;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(fmt.Sprintf(`
    CREATE TRIGGER trigger_alter_time
    BEFORE UPDATE ON %s
    FOR EACH ROW
    WHEN (OLD.data IS DISTINCT FROM NEW.data)
    EXECUTE PROCEDURE alter_time();`, options.SaltCache))
	if exec.Error != nil && !strings.Contains(exec.Error.Error(), "SQLSTATE 42710") {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
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

	exec = DB.Exec(`
    CREATE OR REPLACE FUNCTION delete_session_user_map_entry()
    RETURNS TRIGGER AS $$
    BEGIN
        DELETE FROM session_user_map
        WHERE session_id = OLD.id;
        RETURN OLD;
    END;
    $$ LANGUAGE plpgsql;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE TRIGGER trigger_delete_session_user_map
    AFTER DELETE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION delete_session_user_map_entry();`)
	if exec.Error != nil && !strings.Contains(exec.Error.Error(), "SQLSTATE 42710") {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE OR REPLACE FUNCTION update_user_password_on_session_delete()
    RETURNS TRIGGER AS $$
    BEGIN
        UPDATE user_settings
        SET token = ''
        WHERE user_id = (
            SELECT user_id FROM session_user_map
            WHERE session_id = OLD.id
        );
        RETURN OLD;
    END;
    $$ LANGUAGE plpgsql;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE TRIGGER trigger_update_user_password_on_session_delete
    AFTER DELETE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_user_password_on_session_delete();`)
	if exec.Error != nil && !strings.Contains(exec.Error.Error(), "SQLSTATE 42710") {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE OR REPLACE VIEW vw_salt_minions AS
	SELECT 
		distinct on (substring(bank from 9))
		substring(bank from 9) as minion_id,
		data -> 'grains' as grains,
		data -> 'pillar' as pillar,
		id,
		alter_time
	FROM
		salt_cache 
	WHERE 
		psql_key = 'data'`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(fmt.Sprintf(`
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
        (a.full_ret::jsonb ->> 'fun_args'::text) = '[]'::text`, options.SaltReturns))
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE OR REPLACE VIEW vw_conformity AS
    WITH normalized AS (
            SELECT vw_salt_highstates.id,
                vw_salt_highstates.alter_time,
                vw_salt_highstates.success,
                    CASE
                        WHEN jsonb_typeof(vw_salt_highstates.return) = 'object'::text THEN vw_salt_highstates.return
                        ELSE '{}'::jsonb
                    END AS return
            FROM vw_salt_highstates
            ), results AS (
            SELECT normalized.id,
                normalized.alter_time,
                normalized.success,
                ((normalized.return -> item.key) ->> 'result'::text)::boolean AS result_value,
                (normalized.return -> item.key) -> 'changes'::text AS changes_value
            FROM normalized
                CROSS JOIN LATERAL jsonb_each(normalized.return) item(key, value)
            )
    SELECT results.id,
        results.alter_time,
        (count(*) FILTER (WHERE results.result_value = false) = 0)::boolean AS success,
        count(*) FILTER (WHERE results.result_value = true) AS true_count,
        count(*) FILTER (WHERE results.result_value = false) AS false_count,
        count(*) FILTER (WHERE results.result_value = true AND results.changes_value <> '{}'::jsonb) AS changed_count,
        count(*) FILTER (WHERE results.result_value = true AND results.changes_value = '{}'::jsonb) AS unchanged_count
    FROM results
    GROUP BY results.id, results.alter_time, results.success;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
    CREATE MATERIALIZED VIEW IF NOT EXISTS mat_conformity AS
    WITH normalized AS (
        SELECT vw_salt_highstates.id,
            vw_salt_highstates.alter_time,
            vw_salt_highstates.success,
                CASE
                    WHEN jsonb_typeof(vw_salt_highstates.return) = 'object'::text THEN vw_salt_highstates.return
                    ELSE '{}'::jsonb
                END AS return
        FROM vw_salt_highstates
        ), results AS (
        SELECT normalized.id,
            normalized.alter_time,
            normalized.success,
            ((normalized.return -> item.key) ->> 'result'::text)::boolean AS result_value,
            (normalized.return -> item.key) -> 'changes'::text AS changes_value
        FROM normalized
            CROSS JOIN LATERAL jsonb_each(normalized.return) item(key, value)
        )
    SELECT results.id,
        results.alter_time,
        (count(*) FILTER (WHERE results.result_value = false) = 0)::boolean AS success,
        count(*) FILTER (WHERE results.result_value = true) AS true_count,
        count(*) FILTER (WHERE results.result_value = false) AS false_count,
        count(*) FILTER (WHERE results.result_value = true AND results.changes_value <> '{}'::jsonb) AS changed_count,
        count(*) FILTER (WHERE results.result_value = true AND results.changes_value = '{}'::jsonb) AS unchanged_count
    FROM results
    GROUP BY results.id, results.alter_time, results.success;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_unique_idx ON mat_conformity (id);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_alter_time_idx ON mat_conformity (alter_time);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_success_idx ON mat_conformity (success);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_true_count_idx ON mat_conformity (true_count);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_false_count_idx ON mat_conformity (false_count);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_changed_count_idx ON mat_conformity (changed_count);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS mat_conformity_unchanged_count_idx ON mat_conformity (unchanged_count);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS mat_salt_cache_data_keys AS
		WITH RECURSIVE json_tree AS (
			-- Initial query to select the top-level keys and their values
			SELECT
				jsonb_each.key AS key,
				jsonb_each.value AS value,
				'''' || jsonb_each.key || '''' AS path,
				jsonb_typeof(jsonb_each.value) AS value_type
			FROM salt_cache,
			LATERAL jsonb_each(data)
			WHERE jsonb_typeof(data) = 'object'

			UNION ALL

			-- Recursive part to traverse nested objects
			SELECT
				jsonb_each.key AS key,
				jsonb_each.value AS value,
				json_tree.path || '.' || '''' || jsonb_each.key || '''' AS path,
				jsonb_typeof(jsonb_each.value) AS value_type
			FROM json_tree,
			LATERAL jsonb_each(json_tree.value)
			WHERE jsonb_typeof(json_tree.value) = 'object'
		)
		-- Select distinct paths where the value is not an object or array
		SELECT DISTINCT path
		FROM json_tree
		WHERE value_type IN ('string', 'number', 'boolean', 'array')
		ORDER BY path;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS mat_salt_cache_data_keys_unique_idx ON mat_salt_cache_data_keys (path);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS mat_salt_minions_grains_keys AS
		WITH RECURSIVE json_tree AS (
			-- Initial query to select the top-level keys and their values
			SELECT
				jsonb_each.key AS key,
				jsonb_each.value AS value,
				'''' || jsonb_each.key || '''' AS path,
				jsonb_typeof(jsonb_each.value) AS value_type
			FROM salt_cache,
			LATERAL jsonb_each(data->'grains')
			WHERE jsonb_typeof(data->'grains') = 'object'

			UNION ALL

			-- Recursive part to traverse nested objects
			SELECT
				jsonb_each.key AS key,
				jsonb_each.value AS value,
				json_tree.path || '.' || '''' || jsonb_each.key || '''' AS path,
				jsonb_typeof(jsonb_each.value) AS value_type
			FROM json_tree,
			LATERAL jsonb_each(json_tree.value)
			WHERE jsonb_typeof(json_tree.value) = 'object'
		)
		-- Select distinct paths where the value is not an object or array
		SELECT DISTINCT path
		FROM json_tree
		WHERE value_type IN ('string', 'number', 'boolean', 'array')
		ORDER BY path;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS mat_salt_minions_grains_keys_unique_idx ON mat_salt_minions_grains_keys (path);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`
		CREATE MATERIALIZED VIEW IF NOT EXISTS mat_salt_minions_pillar_keys AS
		WITH RECURSIVE json_tree AS (
			-- Initial query to select the top-level keys and their values
			SELECT
				jsonb_each.key AS key,
				jsonb_each.value AS value,
				'''' || jsonb_each.key || '''' AS path,
				jsonb_typeof(jsonb_each.value) AS value_type
			FROM salt_cache,
			LATERAL jsonb_each(data->'pillar')
			WHERE jsonb_typeof(data->'pillar') = 'object'

			UNION ALL

			-- Recursive part to traverse nested objects
			SELECT
				jsonb_each.key AS key,
				jsonb_each.value AS value,
				json_tree.path || '.' || '''' || jsonb_each.key || '''' AS path,
				jsonb_typeof(jsonb_each.value) AS value_type
			FROM json_tree,
			LATERAL jsonb_each(json_tree.value)
			WHERE jsonb_typeof(json_tree.value) = 'object'
		)
		-- Select distinct paths where the value is not an object or array
		SELECT DISTINCT path
		FROM json_tree
		WHERE value_type IN ('string', 'number', 'boolean', 'array')
		ORDER BY path;`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	exec = DB.Exec(`CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS mat_salt_minions_pillar_keys_unique_idx ON mat_salt_minions_pillar_keys (path);`)
	if exec.Error != nil {
		log.Printf("Error during migration: %v", exec.Error)
		return exec.Error
	}

	log.Printf("Database Migrated")
	return nil
}
