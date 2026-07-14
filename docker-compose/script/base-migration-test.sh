#!/bin/sh
set -eu

cd "$(dirname "$0")/../.."

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose-bare.yaml}"
if [ -z "${PODMAN_COMPOSE:-}" ]; then
	if command -v podman >/dev/null 2>&1; then
		PODMAN_COMPOSE="podman compose"
	elif [ -x /opt/homebrew/bin/podman ]; then
		PODMAN_COMPOSE="/opt/homebrew/bin/podman compose"
	else
		echo "podman not found; set PODMAN_COMPOSE to the compose command to use" >&2
		exit 127
	fi
fi

compose() {
	$PODMAN_COMPOSE -f "$COMPOSE_FILE" "$@"
}

cleanup() {
	compose down -v --remove-orphans >/dev/null 2>&1 || true
}

wait_for_db() {
	for _ in $(seq 1 30); do
		if compose exec -T db pg_isready -U agartha -d agartha -h 127.0.0.1 -p 5432 >/dev/null 2>&1; then
			return 0
		fi
		sleep 1
	done

	echo "database did not become ready" >&2
	return 1
}

run_migrate() {
	compose run --rm -e "AGARTHA_DB_TABLES_USE_JSONB=$1" migrate
}

psql() {
	compose exec -T db psql -U agartha -d agartha -v ON_ERROR_STOP=1 "$@"
}

trap cleanup EXIT INT TERM

echo "==> JSONB migration: fresh, repeat, and concurrent"
cleanup
compose up -d db
wait_for_db
run_migrate true
run_migrate true
run_migrate true &
pid_one=$!
run_migrate true &
pid_two=$!
wait "$pid_one"
wait "$pid_two"

echo "==> text JSON migration: escaped-null regression coverage"
cleanup
compose up -d db
wait_for_db
run_migrate false

psql -c "INSERT INTO salt_returns (fun, jid, \"return\", full_ret, id, success, alter_time) VALUES ('state.highstate', '20260514133100000000', '{\"bad\":\"\\u0000\"}', '{\"fun_args\":[]}', 'poisoned-return-minion', 'true', now()) ON CONFLICT (jid, id) DO UPDATE SET \"return\" = EXCLUDED.\"return\", full_ret = EXCLUDED.full_ret; DELETE FROM schema_migrations;"
run_migrate false

psql -c "DROP INDEX IF EXISTS idx_salt_returns_fun_fullret; INSERT INTO salt_returns (fun, jid, \"return\", full_ret, id, success, alter_time) VALUES ('state.highstate', '20260514133400000000', '{\"ok\":true}', '{\"fun_args\":[],\"bad\":\"\\u0000\"}', 'poisoned-full-ret-minion', 'true', now()) ON CONFLICT (jid, id) DO UPDATE SET \"return\" = EXCLUDED.\"return\", full_ret = EXCLUDED.full_ret; DELETE FROM schema_migrations;"
run_migrate false

echo "==> base compose migration test passed"
