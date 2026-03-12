-- Core SQL objects managed through golang-migrate.

CREATE OR REPLACE VIEW vw_salt_minions AS
SELECT DISTINCT ON (substring(bank FROM 9))
	substring(bank FROM 9) AS minion_id,
	data -> 'grains' AS grains,
	data -> 'pillar' AS pillar,
	id,
	alter_time
FROM
	salt_cache
WHERE
	psql_key = 'data';

CREATE OR REPLACE VIEW vw_conformity AS
WITH normalized AS (
	SELECT
		vw_salt_highstates.id,
		vw_salt_highstates.alter_time,
		vw_salt_highstates.success,
		CASE
			WHEN jsonb_typeof(vw_salt_highstates.return) = 'object' THEN vw_salt_highstates.return
			ELSE '{}'::jsonb
		END AS return
	FROM vw_salt_highstates
),
results AS (
	SELECT
		normalized.id,
		normalized.alter_time,
		normalized.success,
		((normalized.return -> item.key) ->> 'result')::boolean AS result_value,
		(normalized.return -> item.key) -> 'changes' AS changes_value
	FROM normalized
		CROSS JOIN LATERAL jsonb_each(normalized.return) item(key, value)
)
SELECT
	results.id,
	results.alter_time,
	(COUNT(*) FILTER (WHERE results.result_value = false) = 0)::boolean AS success,
	COUNT(*) FILTER (WHERE results.result_value = true) AS true_count,
	COUNT(*) FILTER (WHERE results.result_value = false) AS false_count,
	COUNT(*) FILTER (WHERE results.result_value = true AND results.changes_value <> '{}'::jsonb) AS changed_count,
	COUNT(*) FILTER (WHERE results.result_value = true AND results.changes_value = '{}'::jsonb) AS unchanged_count
FROM results
GROUP BY
	results.id,
	results.alter_time,
	results.success;

DROP MATERIALIZED VIEW IF EXISTS mat_conformity;
CREATE MATERIALIZED VIEW mat_conformity AS
WITH normalized AS (
	SELECT
		vw_salt_highstates.id,
		vw_salt_highstates.alter_time,
		vw_salt_highstates.success,
		CASE
			WHEN jsonb_typeof(vw_salt_highstates.return) = 'object' THEN vw_salt_highstates.return
			ELSE '{}'::jsonb
		END AS return
	FROM vw_salt_highstates
),
results AS (
	SELECT
		normalized.id,
		normalized.alter_time,
		normalized.success,
		((normalized.return -> item.key) ->> 'result')::boolean AS result_value,
		(normalized.return -> item.key) -> 'changes' AS changes_value
	FROM normalized
		CROSS JOIN LATERAL jsonb_each(normalized.return) item(key, value)
)
SELECT
	results.id,
	results.alter_time,
	(COUNT(*) FILTER (WHERE results.result_value = false) = 0)::boolean AS success,
	COUNT(*) FILTER (WHERE results.result_value = true) AS true_count,
	COUNT(*) FILTER (WHERE results.result_value = false) AS false_count,
	COUNT(*) FILTER (WHERE results.result_value = true AND results.changes_value <> '{}'::jsonb) AS changed_count,
	COUNT(*) FILTER (WHERE results.result_value = true AND results.changes_value = '{}'::jsonb) AS unchanged_count
FROM results
GROUP BY
	results.id,
	results.alter_time,
	results.success;

CREATE UNIQUE INDEX IF NOT EXISTS mat_conformity_unique_idx ON mat_conformity (id);
CREATE INDEX IF NOT EXISTS mat_conformity_alter_time_idx ON mat_conformity (alter_time);
CREATE INDEX IF NOT EXISTS mat_conformity_success_idx ON mat_conformity (success);
CREATE INDEX IF NOT EXISTS mat_conformity_true_count_idx ON mat_conformity (true_count);
CREATE INDEX IF NOT EXISTS mat_conformity_false_count_idx ON mat_conformity (false_count);
CREATE INDEX IF NOT EXISTS mat_conformity_changed_count_idx ON mat_conformity (changed_count);
CREATE INDEX IF NOT EXISTS mat_conformity_unchanged_count_idx ON mat_conformity (unchanged_count);

DROP MATERIALIZED VIEW IF EXISTS mat_salt_cache_data_keys;
CREATE MATERIALIZED VIEW mat_salt_cache_data_keys AS
WITH RECURSIVE json_tree AS (
	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		'''' || jsonb_each.key || '''' AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM salt_cache,
		LATERAL jsonb_each(data)
	WHERE jsonb_typeof(data) = 'object'

	UNION ALL

	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		json_tree.path || '.' || '''' || jsonb_each.key || '''' AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM json_tree,
		LATERAL jsonb_each(json_tree.value)
	WHERE jsonb_typeof(json_tree.value) = 'object'
)
SELECT DISTINCT path
FROM json_tree
WHERE value_type IN ('string', 'number', 'boolean', 'array')
ORDER BY path;

CREATE UNIQUE INDEX IF NOT EXISTS mat_salt_cache_data_keys_unique_idx ON mat_salt_cache_data_keys (path);

DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_grains_keys;
CREATE MATERIALIZED VIEW mat_salt_minions_grains_keys AS
WITH RECURSIVE json_tree AS (
	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		jsonb_each.key AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM salt_cache,
		LATERAL jsonb_each(data->'grains')
	WHERE jsonb_typeof(data->'grains') = 'object'

	UNION ALL

	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		json_tree.path || ':' || jsonb_each.key AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM json_tree,
		LATERAL jsonb_each(json_tree.value)
	WHERE jsonb_typeof(json_tree.value) = 'object'
)
SELECT DISTINCT path
FROM json_tree
WHERE value_type IN ('string', 'number', 'boolean', 'array')
ORDER BY path;

CREATE UNIQUE INDEX IF NOT EXISTS mat_salt_minions_grains_keys_unique_idx ON mat_salt_minions_grains_keys (path);

DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_pillar_keys;
CREATE MATERIALIZED VIEW mat_salt_minions_pillar_keys AS
WITH RECURSIVE json_tree AS (
	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		'''' || jsonb_each.key || '''' AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM salt_cache,
		LATERAL jsonb_each(data->'pillar')
	WHERE jsonb_typeof(data->'pillar') = 'object'

	UNION ALL

	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		json_tree.path || '.' || '''' || jsonb_each.key || '''' AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM json_tree,
		LATERAL jsonb_each(json_tree.value)
	WHERE jsonb_typeof(json_tree.value) = 'object'
)
SELECT DISTINCT path
FROM json_tree
WHERE value_type IN ('string', 'number', 'boolean', 'array')
ORDER BY path;

CREATE UNIQUE INDEX IF NOT EXISTS mat_salt_minions_pillar_keys_unique_idx ON mat_salt_minions_pillar_keys (path);

CREATE OR REPLACE FUNCTION delete_session_user_map_entry()
RETURNS TRIGGER AS $$
BEGIN
	DELETE FROM session_user_map
	WHERE session_id = OLD.id;
	RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_delete_session_user_map ON sessions;
CREATE TRIGGER trigger_delete_session_user_map
AFTER DELETE ON sessions
FOR EACH ROW
EXECUTE FUNCTION delete_session_user_map_entry();

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
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_user_password_on_session_delete ON sessions;
CREATE TRIGGER trigger_update_user_password_on_session_delete
AFTER DELETE ON sessions
FOR EACH ROW
EXECUTE FUNCTION update_user_password_on_session_delete();
