package db

import (
	"log"

	agartha "github.com/PaulChristophel/agartha/server/model/agartha"
	salt "github.com/PaulChristophel/agartha/server/model/salt"
)

// // InitializeDatabaseSchema migrates the database schema
// func InitializeDatabaseSchema() {
// 	DB.Exec("CREATE SEQUENCE IF NOT EXISTS seq_salt_events_id;")
// 	DB.AutoMigrate(&model.JID{})
// 	DB.AutoMigrate(&model.SaltReturn{})
// 	DB.AutoMigrate(&model.SaltEvent{})
// 	DB.AutoMigrate(&model.SaltCache{})
// 	DB.Exec(fmt.Sprintf(`CREATE OR REPLACE FUNCTION data_changed() RETURNS trigger
// 		LANGUAGE plpgsql
// 		AS $$
// 	BEGIN
// 		NEW.data_changed := current_timestamp;
// 		RETURN NEW;
// 	END;
// 	$$;

// 	CREATE TRIGGER trigger_data_changed
// 	BEFORE UPDATE ON %s
// 	FOR EACH ROW
// 	WHEN (OLD.data IS DISTINCT FROM NEW.data)
// 	EXECUTE PROCEDURE data_changed();`, model.SaltCache{}.TableName()))
// 	DB.Exec(`CREATE MATERIALIZED VIEW IF NOT EXISTS mat_salt_highstates as SELECT DISTINCT ON (a.id) a.fun,
// 	a.jid,
// 	a.return::jsonb AS return,
// 	a.full_ret::jsonb AS full_ret,
// 	a.id,
// 	a.success,
// 	a.alter_time,
// 	b.load::jsonb ->> 'user'::text AS "user",
// 	a.uuid
// 	FROM salt_returns a
// 	 LEFT JOIN jids b ON a.jid::text = b.jid::text
// 	WHERE (a.fun::text = 'state.highstate'::text OR a.fun::text = 'state.apply'::text) AND (a.full_ret::jsonb ->> 'fun_args'::text) = '[]'::text
// 	ORDER BY a.id, a.jid DESC;`)
// 	fmt.Println("Database Migrated")
// }

func Migrate() {
	// Configure JIDs
	DB.AutoMigrate(&salt.JID{})

	// Configure SaltReturns
	DB.AutoMigrate(&salt.SaltReturn{})

	// Configure SaltEvents
	DB.AutoMigrate(&salt.SaltEvent{})
	DB.Exec("CREATE SEQUENCE IF NOT EXISTS seq_salt_events_id;")
	DB.Exec("ALTER TABLE salt_events ALTER COLUMN id SET DEFAULT nextval('seq_salt_events_id');")

	// Configure SaltCache
	DB.AutoMigrate(&salt.SaltCache{})
	DB.Exec(`
    CREATE OR REPLACE FUNCTION alter_time() RETURNS trigger
        LANGUAGE plpgsql
        AS $$
    BEGIN
    NEW.alter_time := current_timestamp;
    RETURN NEW;
    END;
    $$;

    CREATE TRIGGER trigger_alter_time
    BEFORE UPDATE ON salt_cache
    FOR EACH ROW
    WHEN (OLD.data IS DISTINCT FROM NEW.data)
    EXECUTE PROCEDURE alter_time();`)

	// Configure AuthUser
	DB.AutoMigrate(&agartha.AuthUser{})

	// Configure UserSettings
	DB.AutoMigrate(&agartha.UserSettings{})

	// Configure Session
	DB.AutoMigrate(&agartha.Session{})

	// Configure SessionUserMap
	DB.AutoMigrate(&agartha.SessionUserMap{})

	DB.Exec(`
    CREATE MATERIALIZED VIEW IF NOT EXISTS mat_salt_highstates AS
    SELECT
        distinct on(a.id)
        a.fun,
        a.jid,
        a.return::jsonb,
        a.full_ret::jsonb,
        a.id,
        a.success,
        a.alter_time,
        b.load::jsonb->>'user' as user
    FROM salt_returns a
    LEFT JOIN jids b ON a.jid::text = b.jid::text
    WHERE
        (a.fun::text = 'state.highstate'::text OR a.fun::text = 'state.apply'::text) AND
        (a.full_ret::jsonb ->> 'fun_args'::text) = '[]'::text
    ORDER BY
        a.id, a.jid DESC;`)

	DB.Exec(`
    CREATE MATERIALIZED VIEW IF NOT EXISTS mat_conformity AS
    WITH normalized AS (
            SELECT mat_salt_highstates.id,
                mat_salt_highstates.alter_time,
                mat_salt_highstates.success,
                mat_salt_highstates."user",
                    CASE
                        WHEN jsonb_typeof(mat_salt_highstates.return) = 'object'::text THEN mat_salt_highstates.return
                        ELSE '{}'::jsonb
                    END AS return
            FROM mat_salt_highstates
            ), results AS (
            SELECT normalized.id,
                normalized.alter_time,
                normalized.success,
                normalized."user",
                ((normalized.return -> item.key) ->> 'result'::text)::boolean AS result_value,
                (normalized.return -> item.key) -> 'changes'::text AS changes_value
            FROM normalized
                CROSS JOIN LATERAL jsonb_each(normalized.return) item(key, value)
            )
    SELECT results.id,
        results.alter_time,
        results.success::boolean AS success,
        results."user",
        count(*) FILTER (WHERE results.result_value = true) AS true_count,
        count(*) FILTER (WHERE results.result_value = false) AS false_count,
        count(*) FILTER (WHERE results.result_value = true AND results.changes_value <> '{}'::jsonb) AS changed_count,
        count(*) FILTER (WHERE results.result_value = true AND results.changes_value = '{}'::jsonb) AS unchanged_count
    FROM results
    GROUP BY results.id, results.alter_time, results.success, results."user";`)
	log.Printf("Database Migrated")
}
