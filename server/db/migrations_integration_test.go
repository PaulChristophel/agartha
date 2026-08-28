//go:build integration

package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const migrationTestDatabaseURL = "AGARTHA_TEST_DATABASE_URL"

type minionFixture struct {
	ID     string
	OS     string
	Role   string
	Source string
}

func TestSQLMigrationsSaltCacheCompatibility(t *testing.T) {
	databaseURL := os.Getenv(migrationTestDatabaseURL)
	if databaseURL == "" {
		t.Skipf("%s is not set", migrationTestDatabaseURL)
	}
	requireTestDatabase(t, databaseURL)

	verificationDB, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, verificationDB.Close()) })
	require.NoError(t, verificationDB.Ping())

	resetMigrationSchema(t, verificationDB)
	t.Cleanup(func() { resetMigrationSchema(t, verificationDB) })
	createMigrationBaseline(t, verificationDB)
	insertSaltCacheFixtures(t, verificationDB)
	insertSaltReturnFixtures(t, verificationDB)

	migrator := newTestMigrator(t, databaseURL)
	t.Cleanup(func() {
		sourceErr, databaseErr := migrator.Close()
		require.NoError(t, errors.Join(sourceErr, databaseErr))
	})

	require.NoError(t, migrator.Up())
	requireMigrationVersion(t, migrator, 3)
	requireMinionFixtures(t, verificationDB)
	requireMaterializedKeys(t, verificationDB)
	requireLatestHighstateIgnoresRunningCollision(t, verificationDB)
	requireConformityUsesCacheWithoutKeyTable(t, verificationDB)
	requireConformityUsesAcceptedDatabaseKeys(t, verificationDB)

	require.NoError(t, migrator.Steps(-1))
	requireMigrationVersion(t, migrator, 2)
	requireMinionFixtures(t, verificationDB)

	require.NoError(t, migrator.Steps(-1))
	requireMigrationVersion(t, migrator, 1)
	requireLegacyViewAfterDown(t, verificationDB)

	require.NoError(t, migrator.Steps(2))
	requireMigrationVersion(t, migrator, 3)
	requireMinionFixtures(t, verificationDB)
	requireMaterializedKeys(t, verificationDB)
}

func requireTestDatabase(t *testing.T, databaseURL string) {
	t.Helper()

	parsedURL, err := url.Parse(databaseURL)
	require.NoError(t, err)
	require.Equal(t, "postgres", parsedURL.Scheme)
	require.Equal(t, "agartha_migration_test", path.Base(parsedURL.Path), "refusing to reset a non-test database")
}

func resetMigrationSchema(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;`)
	require.NoError(t, err)
}

func createMigrationBaseline(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`
		CREATE TABLE salt_cache (
			bank text NOT NULL,
			psql_key text NOT NULL,
			data jsonb,
			id text NOT NULL,
			alter_time timestamptz
		);

		CREATE TABLE salt_returns (
			fun text NOT NULL,
			jid text NOT NULL,
			return text NOT NULL,
			full_ret text NOT NULL,
			id text NOT NULL,
			success text NOT NULL,
			alter_time timestamptz
		);

		CREATE TABLE sessions (id text PRIMARY KEY);
		CREATE TABLE session_user_map (session_id text NOT NULL, user_id bigint NOT NULL);
		CREATE TABLE user_settings (user_id bigint PRIMARY KEY, token text NOT NULL DEFAULT '');
	`)
	require.NoError(t, err)
	_, err = database.Exec(legacySaltHighstatesViewQuery("salt_returns"))
	require.NoError(t, err)
}

func legacySaltHighstatesViewQuery(saltReturnsTable string) string {
	return fmt.Sprintf(`
		CREATE VIEW vw_salt_highstates AS
		SELECT DISTINCT ON (a.id)
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
		) AS a
		WHERE a.fun IN ('state.highstate', 'state.apply')
			AND (a.full_ret::jsonb ->> 'fun_args') = '[]'`, saltReturnsTable)
}

func insertSaltCacheFixtures(t *testing.T, database *sql.DB) {
	t.Helper()

	fixtures := []struct {
		bank      string
		key       string
		data      string
		id        string
		alterTime time.Time
	}{
		{
			bank:      "minions/legacy-one",
			key:       "data",
			data:      `{"grains":{"os":"legacy-linux"},"pillar":{"role":"legacy"}}`,
			id:        "legacy-id",
			alterTime: time.Date(2026, time.July, 1, 1, 0, 0, 0, time.UTC),
		},
		{
			bank:      "grains",
			key:       "modern-one",
			data:      `{"os":"modern-linux","nested":{"arch":"amd64"}}`,
			id:        "modern-grains-id",
			alterTime: time.Date(2026, time.July, 2, 1, 0, 0, 0, time.UTC),
		},
		{
			bank:      "pillar",
			key:       "modern-one",
			data:      `{"role":"modern"}`,
			id:        "modern-pillar-id",
			alterTime: time.Date(2026, time.July, 2, 2, 0, 0, 0, time.UTC),
		},
		{
			bank:      "grains",
			key:       "suffix-one:base",
			data:      `{"os":"suffix-linux"}`,
			id:        "suffix-grains-id",
			alterTime: time.Date(2026, time.July, 3, 1, 0, 0, 0, time.UTC),
		},
		{
			bank:      "pillar",
			key:       "suffix-one:prod",
			data:      `{"role":"suffix-fallback"}`,
			id:        "suffix-pillar-id",
			alterTime: time.Date(2026, time.July, 3, 2, 0, 0, 0, time.UTC),
		},
		{
			bank:      "minions/preferred-one",
			key:       "data",
			data:      `{"grains":{"os":"stale-legacy"},"pillar":{"role":"stale-legacy"}}`,
			id:        "preferred-legacy-id",
			alterTime: time.Date(2026, time.July, 4, 1, 0, 0, 0, time.UTC),
		},
		{
			bank:      "grains",
			key:       "preferred-one",
			data:      `{"os":"preferred-modern"}`,
			id:        "preferred-grains-id",
			alterTime: time.Date(2026, time.July, 4, 2, 0, 0, 0, time.UTC),
		},
		{
			bank:      "pillar",
			key:       "preferred-one",
			data:      `{"role":"preferred-modern"}`,
			id:        "preferred-pillar-id",
			alterTime: time.Date(2026, time.July, 4, 3, 0, 0, 0, time.UTC),
		},
	}

	for _, fixture := range fixtures {
		_, err := database.Exec(
			`INSERT INTO salt_cache (bank, psql_key, data, id, alter_time) VALUES ($1, $2, $3, $4, $5)`,
			fixture.bank,
			fixture.key,
			fixture.data,
			fixture.id,
			fixture.alterTime,
		)
		require.NoError(t, err)
	}
}

func insertSaltReturnFixtures(t *testing.T, database *sql.DB) {
	t.Helper()

	fixtures := []struct {
		jid        string
		returnData string
		fullRet    string
		id         string
		success    string
		alterTime  time.Time
	}{
		{
			jid:        "20260704010000000000",
			returnData: `{"test_|-cached_|-cached_|-succeed_without_changes":{"result":true,"changes":{}}}`,
			id:         "modern-one",
			success:    "true",
			alterTime:  time.Date(2026, time.July, 4, 1, 0, 0, 0, time.UTC),
		},
		{
			jid:        "20260704020000000000",
			returnData: `["The function \"state.highstate\" is running as PID 1234 and was started at 2026, Jul 04 00:59:00.000000 with jid 20260704005900000000"]`,
			fullRet:    `{"fun_args":[],"retcode":1}`,
			id:         "modern-one",
			success:    "false",
			alterTime:  time.Date(2026, time.July, 4, 2, 0, 0, 0, time.UTC),
		},
		{
			jid:        "20260704030000000000",
			returnData: `{"test_|-suffix_|-suffix_|-succeed_with_changes":{"result":true,"changes":{"old":"before","new":"after"}}}`,
			id:         "suffix-one",
			success:    "true",
			alterTime:  time.Date(2026, time.July, 4, 3, 0, 0, 0, time.UTC),
		},
		{
			jid:        "20260704040000000000",
			returnData: `{"test_|-uncached_|-uncached_|-fail":{"result":false,"changes":{}}}`,
			id:         "uncached-one",
			success:    "false",
			alterTime:  time.Date(2026, time.July, 4, 4, 0, 0, 0, time.UTC),
		},
		{
			jid:        "20260704060000000000",
			returnData: `{"test_|-older_|-older_|-succeed_without_changes":{"result":true,"changes":{}}}`,
			id:         "ordering-one",
			success:    "true",
			alterTime:  time.Date(2026, time.July, 4, 5, 0, 0, 0, time.UTC),
		},
		{
			jid:        "20260704050000000000",
			returnData: `{"test_|-newer_|-newer_|-succeed_without_changes":{"result":true,"changes":{}}}`,
			id:         "ordering-one",
			success:    "true",
			alterTime:  time.Date(2026, time.July, 4, 6, 0, 0, 0, time.UTC),
		},
		{
			jid:        "20260704070000000000",
			returnData: `["The function \"state.highstate\" is running as PID 5678 and was started at 2026, Jul 04 06:59:00.000000 with jid 20260704065900000000"]`,
			fullRet:    `{"fun_args":[],"retcode":2}`,
			id:         "different-retcode",
			success:    "false",
			alterTime:  time.Date(2026, time.July, 4, 7, 0, 0, 0, time.UTC),
		},
	}

	for _, fixture := range fixtures {
		_, err := database.Exec(`
			INSERT INTO salt_returns (fun, jid, return, full_ret, id, success, alter_time)
			VALUES ('state.highstate', $1, $2, COALESCE(NULLIF($3, ''), '{"fun_args":[]}'), $4, $5, $6)
		`, fixture.jid, fixture.returnData, fixture.fullRet, fixture.id, fixture.success, fixture.alterTime)
		require.NoError(t, err)
	}
}

func newTestMigrator(t *testing.T, databaseURL string) *migrate.Migrate {
	t.Helper()

	migrationDB, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)
	driver, err := postgres.WithInstance(migrationDB, &postgres.Config{})
	require.NoError(t, err)
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	require.NoError(t, err)
	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	require.NoError(t, err)
	return migrator
}

func requireMigrationVersion(t *testing.T, migrator *migrate.Migrate, expected uint) {
	t.Helper()

	version, dirty, err := migrator.Version()
	require.NoError(t, err)
	require.Equal(t, expected, version)
	require.False(t, dirty)
}

func requireMinionFixtures(t *testing.T, database *sql.DB) {
	t.Helper()

	rows, err := database.Query(`
		SELECT
			minion_id,
			COALESCE(grains ->> 'os', ''),
			COALESCE(pillar ->> 'role', ''),
			id
		FROM vw_salt_minions
		ORDER BY minion_id
	`)
	require.NoError(t, err)

	var actual []minionFixture
	for rows.Next() {
		var fixture minionFixture
		require.NoError(t, rows.Scan(&fixture.Source, &fixture.OS, &fixture.Role, &fixture.ID))
		actual = append(actual, fixture)
	}
	require.NoError(t, rows.Err())
	require.NoError(t, rows.Close())
	require.Equal(t, []minionFixture{
		{ID: "legacy-id", OS: "legacy-linux", Role: "legacy", Source: "legacy-one"},
		{ID: "modern-grains-id", OS: "modern-linux", Role: "modern", Source: "modern-one"},
		{ID: "preferred-grains-id", OS: "preferred-modern", Role: "preferred-modern", Source: "preferred-one"},
		{ID: "suffix-grains-id", OS: "suffix-linux", Role: "suffix-fallback", Source: "suffix-one:base"},
	}, actual)
}

func requireMaterializedKeys(t *testing.T, database *sql.DB) {
	t.Helper()

	var grainsNestedPathCount int
	require.NoError(t, database.QueryRow(
		`SELECT count(*) FROM mat_salt_minions_grains_keys WHERE path = 'nested:arch'`,
	).Scan(&grainsNestedPathCount))
	require.Equal(t, 1, grainsNestedPathCount)

	var pillarRolePathCount int
	require.NoError(t, database.QueryRow(
		`SELECT count(*) FROM mat_salt_minions_pillar_keys WHERE path = $1`,
		`'role'`,
	).Scan(&pillarRolePathCount))
	require.Equal(t, 1, pillarRolePathCount)
}

func requireLatestHighstateIgnoresRunningCollision(t *testing.T, database *sql.DB) {
	t.Helper()

	var modernJID string
	require.NoError(t, database.QueryRow(
		`SELECT jid FROM vw_salt_highstates WHERE id = 'modern-one'`,
	).Scan(&modernJID))
	require.Equal(t, "20260704010000000000", modernJID)

	var orderingJID string
	require.NoError(t, database.QueryRow(
		`SELECT jid FROM vw_salt_highstates WHERE id = 'ordering-one'`,
	).Scan(&orderingJID))
	require.Equal(t, "20260704050000000000", orderingJID)

	var differentRetcodeJID string
	require.NoError(t, database.QueryRow(
		`SELECT jid FROM vw_salt_highstates WHERE id = 'different-retcode'`,
	).Scan(&differentRetcodeJID))
	require.Equal(t, "20260704070000000000", differentRetcodeJID)
}

func requireConformityUsesCacheWithoutKeyTable(t *testing.T, database *sql.DB) {
	t.Helper()

	rows, err := database.Query(`SELECT id FROM mat_conformity ORDER BY id`)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()

	var ids []string
	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"modern-one", "suffix-one"}, ids)
}

func requireConformityUsesAcceptedDatabaseKeys(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`
		CREATE TABLE salt_keys (
			bank text NOT NULL,
			psql_key text NOT NULL,
			data jsonb NOT NULL
		);
		INSERT INTO salt_keys (bank, psql_key, data) VALUES
			('pki/master/keys', 'modern-one', '{"state":"accepted"}'),
			('pki/master/keys', 'suffix-one', '{"state":"pending"}'),
			('pki/master/keys', 'uncached-one', '{"state":"accepted"}');
		REFRESH MATERIALIZED VIEW mat_conformity;
	`)
	require.NoError(t, err)

	rows, err := database.Query(`SELECT id FROM mat_conformity ORDER BY id`)
	require.NoError(t, err)
	defer func() { require.NoError(t, rows.Close()) }()

	var ids []string
	for rows.Next() {
		var id string
		require.NoError(t, rows.Scan(&id))
		ids = append(ids, id)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, []string{"modern-one", "uncached-one"}, ids)
}

func requireLegacyViewAfterDown(t *testing.T, database *sql.DB) {
	t.Helper()

	var legacyCount int
	require.NoError(t, database.QueryRow(
		`SELECT count(*) FROM vw_salt_minions WHERE minion_id = 'legacy-one'`,
	).Scan(&legacyCount))
	require.Equal(t, 1, legacyCount)

	var modernCount int
	require.NoError(t, database.QueryRow(
		`SELECT count(*) FROM vw_salt_minions WHERE minion_id = 'modern-one'`,
	).Scan(&modernCount))
	require.Zero(t, modernCount)
}
