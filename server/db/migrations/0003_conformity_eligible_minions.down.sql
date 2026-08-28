DO $$
DECLARE
	salt_returns_table regclass;
BEGIN
	SELECT DISTINCT source.oid::regclass
	INTO STRICT salt_returns_table
	FROM pg_rewrite AS rewrite
		JOIN pg_depend AS dependency
			ON dependency.classid = 'pg_rewrite'::regclass
			AND dependency.objid = rewrite.oid
			AND dependency.refclassid = 'pg_class'::regclass
		JOIN pg_class AS source
			ON source.oid = dependency.refobjid
	WHERE rewrite.ev_class = 'vw_salt_highstates'::regclass
		AND source.relkind IN ('r', 'p', 'f');

	DROP MATERIALIZED VIEW IF EXISTS mat_conformity;
	DROP VIEW IF EXISTS vw_conformity;
	DROP VIEW vw_salt_highstates;

	EXECUTE format($view$
		CREATE VIEW vw_salt_highstates AS
		SELECT DISTINCT ON (a.id)
			a.fun,
			a.jid,
			a.return::jsonb AS return,
			a.full_ret::jsonb AS full_ret,
			a.id,
			a.success::boolean AS success,
			a.alter_time
		FROM (
			SELECT *
			FROM %s
			WHERE POSITION($nul$\u0000$nul$ IN return::text) = 0
				AND POSITION($nul$\u0000$nul$ IN full_ret::text) = 0
		) AS a
		WHERE a.fun IN ('state.highstate', 'state.apply')
			AND (a.full_ret::jsonb ->> 'fun_args') = '[]'
	$view$, salt_returns_table);
END;
$$;

CREATE VIEW vw_conformity AS
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

CREATE MATERIALIZED VIEW mat_conformity AS
SELECT *
FROM vw_conformity;

CREATE UNIQUE INDEX mat_conformity_unique_idx ON mat_conformity (id);
CREATE INDEX mat_conformity_alter_time_idx ON mat_conformity (alter_time);
CREATE INDEX mat_conformity_success_idx ON mat_conformity (success);
CREATE INDEX mat_conformity_true_count_idx ON mat_conformity (true_count);
CREATE INDEX mat_conformity_false_count_idx ON mat_conformity (false_count);
CREATE INDEX mat_conformity_changed_count_idx ON mat_conformity (changed_count);
CREATE INDEX mat_conformity_unchanged_count_idx ON mat_conformity (unchanged_count);

DROP FUNCTION conformity_minion_ids();
