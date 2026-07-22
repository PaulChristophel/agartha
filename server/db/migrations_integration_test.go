//go:build integration

package db

import (
	"database/sql"
	"errors"
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

	migrator := newTestMigrator(t, databaseURL)
	t.Cleanup(func() {
		sourceErr, databaseErr := migrator.Close()
		require.NoError(t, errors.Join(sourceErr, databaseErr))
	})

	require.NoError(t, migrator.Up())
	requireMigrationVersion(t, migrator, 2)
	requireMinionFixtures(t, verificationDB)
	requireMaterializedKeys(t, verificationDB)

	require.NoError(t, migrator.Steps(-1))
	requireMigrationVersion(t, migrator, 1)
	requireLegacyViewAfterDown(t, verificationDB)

	require.NoError(t, migrator.Steps(1))
	requireMigrationVersion(t, migrator, 2)
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

		CREATE VIEW vw_salt_highstates AS
		SELECT
			NULL::text AS id,
			NULL::timestamptz AS alter_time,
			NULL::boolean AS success,
			NULL::jsonb AS return
		WHERE false;

		CREATE TABLE sessions (id text PRIMARY KEY);
		CREATE TABLE session_user_map (session_id text NOT NULL, user_id bigint NOT NULL);
		CREATE TABLE user_settings (user_id bigint PRIMARY KEY, token text NOT NULL DEFAULT '');
	`)
	require.NoError(t, err)
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
