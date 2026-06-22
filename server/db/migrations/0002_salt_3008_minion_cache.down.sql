DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_pillar_keys;
DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_grains_keys;

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

CREATE MATERIALIZED VIEW mat_salt_minions_grains_keys AS
WITH RECURSIVE json_tree AS (
	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		jsonb_each.key AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM (
		SELECT data
		FROM salt_cache
		WHERE POSITION($nul$\u0000$nul$ IN data::text) = 0
			AND jsonb_typeof(data->'grains') = 'object'
	) salt_cache,
		LATERAL jsonb_each(data->'grains')

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

CREATE MATERIALIZED VIEW mat_salt_minions_pillar_keys AS
WITH RECURSIVE json_tree AS (
	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		'''' || jsonb_each.key || '''' AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM (
		SELECT data
		FROM salt_cache
		WHERE POSITION($nul$\u0000$nul$ IN data::text) = 0
			AND jsonb_typeof(data->'pillar') = 'object'
	) salt_cache,
		LATERAL jsonb_each(data->'pillar')

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
