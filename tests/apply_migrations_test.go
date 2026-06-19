package tests

import (
	"context"
	"database/sql"
	"testing"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

const (
	firstMigrationVersion  = "20260619000100"
	secondMigrationVersion = "20260619000200"
	testVersionTable       = "test_schema_version"
)

func TestApplyMigrationsUsesRegistryAndTracksCurrentVersion(t *testing.T) {
	db := openMigrationTestDB(t)
	ctx := context.Background()

	upResult, err := qmig.ApplyMigrations(ctx, migrationApplyConfig(db, qmig.DirectionUp, "head"))
	require.NoError(t, err)
	assert.Equal(t, qmig.DirectionUp, upResult.Direction)
	assert.Equal(t, "", upResult.From)
	assert.Equal(t, secondMigrationVersion, upResult.To)
	assert.Equal(t, []string{firstMigrationVersion, secondMigrationVersion}, upResult.Applied)
	assertCurrentMigrationVersion(t, db, secondMigrationVersion)
	assertObjectExists(t, db, "table", "users", true)
	assertObjectExists(t, db, "index", "idx_users_name", true)
	assertObjectExists(t, db, "view", "active_users", true)

	noopResult, err := qmig.ApplyMigrations(ctx, migrationApplyConfig(db, qmig.DirectionUp, "head"))
	require.NoError(t, err)
	assert.Equal(t, secondMigrationVersion, noopResult.From)
	assert.Equal(t, secondMigrationVersion, noopResult.To)
	assert.Empty(t, noopResult.Applied)

	downToFirstResult, err := qmig.ApplyMigrations(
		ctx,
		migrationApplyConfig(db, qmig.DirectionDown, firstMigrationVersion),
	)
	require.NoError(t, err)
	assert.Equal(t, qmig.DirectionDown, downToFirstResult.Direction)
	assert.Equal(t, secondMigrationVersion, downToFirstResult.From)
	assert.Equal(t, firstMigrationVersion, downToFirstResult.To)
	assert.Equal(t, []string{secondMigrationVersion}, downToFirstResult.Applied)
	assertCurrentMigrationVersion(t, db, firstMigrationVersion)
	assertObjectExists(t, db, "table", "users", true)
	assertObjectExists(t, db, "index", "idx_users_name", false)
	assertObjectExists(t, db, "view", "active_users", false)

	downToBaseResult, err := qmig.ApplyMigrations(ctx, migrationApplyConfig(db, qmig.DirectionDown, ""))
	require.NoError(t, err)
	assert.Equal(t, firstMigrationVersion, downToBaseResult.From)
	assert.Equal(t, "", downToBaseResult.To)
	assert.Equal(t, []string{firstMigrationVersion}, downToBaseResult.Applied)
	assertCurrentMigrationVersion(t, db, "")
	assertObjectExists(t, db, "table", "users", false)
}

func openMigrationTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)
	return db
}

func migrationApplyConfig(db *sql.DB, direction qmig.MigrationDirection, target string) *qmig.MigrationApplyConfig {
	return &qmig.MigrationApplyConfig{
		DB:           db,
		Dialect:      dialect.SQLite{},
		Registry:     testMigrationRegistry(),
		Direction:    direction,
		Target:       target,
		VersionTable: testVersionTable,
	}
}

func testMigrationRegistry() []qmig.Migration {
	return []qmig.Migration{
		qmig.New(
			secondMigrationVersion,
			func() ddl.Statements {
				return ddl.Statements{
					ddl.CreateIndex("idx_users_name").On("users", ddl.KeyCol("name")),
					ddl.RawSQL("CREATE VIEW active_users AS SELECT id, name FROM users"),
				}
			},
			func() ddl.Statements {
				return ddl.Statements{
					ddl.RawSQL("DROP VIEW active_users"),
					ddl.DropIndex("idx_users_name"),
				}
			},
		),
		qmig.New(
			firstMigrationVersion,
			func() ddl.Statements {
				return ddl.Statements{
					ddl.CreateTable("users").Columns(
						ddl.Column("id", ddl.Integer()).PrimaryKey(),
						ddl.Column("name", ddl.Text()).NotNull(),
					),
				}
			},
			func() ddl.Statements {
				return ddl.Statements{
					ddl.DropTable("users"),
				}
			},
		),
	}
}

func assertCurrentMigrationVersion(t *testing.T, db *sql.DB, want string) {
	t.Helper()

	var got string
	err := db.QueryRow(`SELECT "version" FROM "`+testVersionTable+`" WHERE "id" = ?`, 1).Scan(&got)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func assertObjectExists(t *testing.T, db *sql.DB, kind, name string, want bool) {
	t.Helper()

	var count int
	err := db.QueryRow(
		`SELECT count(*) FROM sqlite_master WHERE type = ? AND name = ?`,
		kind,
		name,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, want, count > 0)
}
