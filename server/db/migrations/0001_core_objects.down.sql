DROP TRIGGER IF EXISTS trigger_update_user_password_on_session_delete ON sessions;
DROP FUNCTION IF EXISTS update_user_password_on_session_delete();

DROP TRIGGER IF EXISTS trigger_delete_session_user_map ON sessions;
DROP FUNCTION IF EXISTS delete_session_user_map_entry();

DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_pillar_keys;
DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_grains_keys;
DROP MATERIALIZED VIEW IF EXISTS mat_salt_cache_data_keys;
DROP MATERIALIZED VIEW IF EXISTS mat_conformity;

DROP VIEW IF EXISTS vw_conformity;
DROP VIEW IF EXISTS vw_salt_minions;
