DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_pillar_keys;
DROP MATERIALIZED VIEW IF EXISTS mat_salt_minions_grains_keys;

CREATE OR REPLACE VIEW vw_salt_minions AS
WITH old_minions AS (
	SELECT
		substring(bank FROM 9) AS minion_id,
		data -> 'grains' AS grains,
		data -> 'pillar' AS pillar,
		id,
		alter_time
	FROM salt_cache
	WHERE psql_key = 'data'
		AND bank LIKE 'minions/%'
),
new_grains AS (
	SELECT DISTINCT ON (normalized.minion_id)
		normalized.minion_id,
		normalized.grains,
		normalized.id,
		normalized.alter_time
	FROM (
		SELECT
			CASE
				WHEN bank LIKE 'minions/%' THEN substring(bank FROM 9)
				ELSE bank::text
			END AS minion_id,
			data AS grains,
			id,
			alter_time,
			CASE
				WHEN bank LIKE 'minions/%' THEN 1
				ELSE 0
			END AS bank_priority
		FROM salt_cache
		WHERE psql_key = 'grains'
	) normalized
	ORDER BY normalized.minion_id, normalized.bank_priority, normalized.alter_time DESC NULLS LAST
),
new_pillar AS (
	SELECT DISTINCT ON (normalized.minion_id)
		normalized.minion_id,
		normalized.pillar,
		normalized.id,
		normalized.alter_time
	FROM (
		SELECT
			CASE
				WHEN bank LIKE 'minions/%' THEN substring(bank FROM 9)
				ELSE bank::text
			END AS minion_id,
			data AS pillar,
			id,
			alter_time,
			CASE
				WHEN bank LIKE 'minions/%' THEN 1
				ELSE 0
			END AS bank_priority
		FROM salt_cache
		WHERE psql_key = 'pillar'
	) normalized
	ORDER BY normalized.minion_id, normalized.bank_priority, normalized.alter_time DESC NULLS LAST
),
new_minions AS (
	SELECT
		COALESCE(new_grains.minion_id, new_pillar.minion_id) AS minion_id,
		new_grains.grains,
		new_pillar.pillar,
		COALESCE(new_grains.id, new_pillar.id) AS id,
		CASE
			WHEN new_grains.alter_time IS NOT NULL AND new_pillar.alter_time IS NOT NULL
				THEN GREATEST(new_grains.alter_time, new_pillar.alter_time)
			ELSE COALESCE(new_grains.alter_time, new_pillar.alter_time)
		END AS alter_time
	FROM new_grains
		FULL OUTER JOIN new_pillar ON new_grains.minion_id = new_pillar.minion_id
)
SELECT
	COALESCE(new_minions.minion_id, old_minions.minion_id) AS minion_id,
	COALESCE(new_minions.grains, old_minions.grains) AS grains,
	COALESCE(new_minions.pillar, old_minions.pillar) AS pillar,
	COALESCE(new_minions.id, old_minions.id) AS id,
	CASE
		WHEN new_minions.alter_time IS NOT NULL AND old_minions.alter_time IS NOT NULL
			THEN GREATEST(new_minions.alter_time, old_minions.alter_time)
		ELSE COALESCE(new_minions.alter_time, old_minions.alter_time)
	END AS alter_time
FROM new_minions
	FULL OUTER JOIN old_minions ON new_minions.minion_id = old_minions.minion_id;

CREATE MATERIALIZED VIEW mat_salt_minions_grains_keys AS
WITH RECURSIVE json_tree AS (
	SELECT
		jsonb_each.key AS key,
		jsonb_each.value AS value,
		jsonb_each.key AS path,
		jsonb_typeof(jsonb_each.value) AS value_type
	FROM (
		SELECT grains
		FROM vw_salt_minions
		WHERE POSITION($nul$\u0000$nul$ IN grains::text) = 0
			AND jsonb_typeof(grains) = 'object'
	) salt_minions,
		LATERAL jsonb_each(grains)

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
		SELECT pillar
		FROM vw_salt_minions
		WHERE POSITION($nul$\u0000$nul$ IN pillar::text) = 0
			AND jsonb_typeof(pillar) = 'object'
	) salt_minions,
		LATERAL jsonb_each(pillar)

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
