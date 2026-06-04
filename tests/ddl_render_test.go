package tests

import (
	"testing"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDDLCreateTablePostgreSQL(t *testing.T) {
	sql, err := ddl.CreateTable("users").
		IfNotExists().
		Columns(
			ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
			ddl.Column("email", ddl.Text()).NotNull(),
			ddl.Column("org_id", ddl.BigInt()),
			ddl.Column("created_at", ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
		).
		Constraints(
			ddl.Unique("email").Named("users_email_key"),
			ddl.ForeignKey("org_id").
				References("orgs", "id").
				OnDelete(ddl.Cascade).
				Named("users_org_id_fk"),
		).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE TABLE IF NOT EXISTS "users" (
    "id" BIGSERIAL PRIMARY KEY,
    "email" TEXT NOT NULL,
    "org_id" BIGINT,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT "users_email_key" UNIQUE ("email"),
    CONSTRAINT "users_org_id_fk" FOREIGN KEY ("org_id") REFERENCES "orgs" ("id") ON DELETE CASCADE
)`, sql)
}

func TestDDLColumnUsesExplicitType(t *testing.T) {
	sql, err := ddl.CreateTable("manual_users").
		Columns(
			ddl.Column("email", ddl.Text()),
			ddl.Column("nickname", ddl.VarChar(64)),
		).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE TABLE "manual_users" (
    "email" TEXT,
    "nickname" VARCHAR(64)
)`, sql)
}

func TestDDLCreateTableMySQL(t *testing.T) {
	sql, err := ddl.CreateTable("users").
		Columns(
			ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
			ddl.Column("email", ddl.VarChar(255)).NotNull().Unique(),
		).
		Render(dialect.MySQL{})

	require.NoError(t, err)
	assert.Equal(t, "CREATE TABLE `users` (\n    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,\n    `email` VARCHAR(255) NOT NULL UNIQUE\n)", sql)
}

func TestDDLAlterTablePostgreSQL(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		AddColumn(ddl.Column("nickname", ddl.VarChar(64)).Default("")).
		SetNotNull("email").
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `ALTER TABLE "users"
    ADD COLUMN "nickname" VARCHAR(64) DEFAULT '',
    ALTER COLUMN "email" SET NOT NULL`, sql)
}

func TestDDLAlterTableSQLiteUnsupported(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		AlterColumnType("email", ddl.Text()).
		Render(dialect.SQLite{})

	assert.Empty(t, sql)
	assert.EqualError(t, err, "SQLite dialect does not support ALTER COLUMN TYPE")
}

func TestDDLAlterColumnTypePostgreSQL(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		AlterColumnType("email", ddl.Text()).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `ALTER TABLE "users"
    ALTER COLUMN "email" TYPE TEXT`, sql)
}

func TestDDLCreateIndexPostgreSQL(t *testing.T) {
	sql, err := ddl.CreateIndex("users_email_active_idx").
		Unique().
		IfNotExists().
		OnCols("users", "email").
		Where(ddl.Col("deleted_at").IsNull()).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE UNIQUE INDEX IF NOT EXISTS "users_email_active_idx" ON "users" ("email") WHERE "deleted_at" IS NULL`, sql)
}

func TestDDLCreateIndexKeyOptionsPostgreSQL(t *testing.T) {
	sql, err := ddl.CreateIndex("users_email_idx").
		On("users", ddl.KeyCol("email").
			Collate("en_US").
			OpClass("text_pattern_ops").
			Desc().
			NullsLast()).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE INDEX "users_email_idx" ON "users" ("email" COLLATE "en_US" text_pattern_ops DESC NULLS LAST)`, sql)
}

func TestDDLCreateIndexMySQLUnsupportedPartialIndex(t *testing.T) {
	sql, err := ddl.CreateIndex("users_email_active_idx").
		OnCols("users", "email").
		Where(ddl.Col("deleted_at").IsNull()).
		Render(dialect.MySQL{})

	assert.Empty(t, sql)
	assert.EqualError(t, err, "MySQL dialect does not support PARTIAL INDEX")
}

func TestDDLDropIndexMySQL(t *testing.T) {
	sql, err := ddl.DropIndex("users_email_idx").
		OnTable("users").
		Render(dialect.MySQL{})

	require.NoError(t, err)
	assert.Equal(t, "DROP INDEX `users_email_idx` ON `users`", sql)
}

func TestDDLAlterIndexRenamePostgreSQL(t *testing.T) {
	sql, err := ddl.AlterIndex("users_email_idx").
		Rename("users_email_key").
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `ALTER INDEX "users_email_idx" RENAME TO "users_email_key"`, sql)
}
