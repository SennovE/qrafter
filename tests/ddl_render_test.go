package tests

import (
	"testing"
	"time"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DDLUser struct {
	q.Table `table:"users"`

	ID        q.Column[int64]  `db:"id"`
	Email     q.Column[string] `db:"email"`
	OrgID     q.Column[int64]  `db:"org_id"`
	DeletedAt q.Column[string] `db:"deleted_at"`
	CreatedAt q.Column[string] `db:"created_at"`
}

type DDLOrg struct {
	q.Table `table:"orgs"`

	ID q.Column[int64] `db:"id"`
}

type DDLInferredUser struct {
	q.Table `table:"inferred_users"`

	ID        q.Column[int64]  `db:"id"`
	Email     q.Column[string] `db:"email" ddl:"VARCHAR(320)"`
	Active    q.Column[bool]
	CreatedAt q.Column[time.Time] `ddl:"type:TIMESTAMP"`
	Data      q.Column[[]byte]    `db:"data"`
	Score     q.Column[float64]   `db:"score" ddl:"NUMERIC(10, 2)"`
	Ignored   q.Column[string]    `ddl:"-"`
}

func TestDDLCreateTablePostgreSQL(t *testing.T) {
	users := q.MustNewTable[DDLUser]()
	orgs := q.MustNewTable[DDLOrg]()

	sql, err := ddl.CreateTable(users).
		IfNotExists().
		Columns(
			ddl.Column(users.ID, ddl.BigSerial()).PrimaryKey(),
			ddl.Column(users.Email, ddl.Text()).NotNull(),
			ddl.Column(users.OrgID, ddl.BigInt()),
			ddl.Column(users.CreatedAt, ddl.TimestampTZ()).NotNull().DefaultExpr("now()"),
		).
		Constraints(
			ddl.Unique(users.Email).Named("users_email_key"),
			ddl.ForeignKey(users.OrgID).
				References(orgs, orgs.ID).
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

func TestDDLCreateTableFromModelInfersColumnTypes(t *testing.T) {
	users := q.MustNewTable[DDLInferredUser]()

	sql, err := ddl.CreateTable(users).
		FromModel().
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE TABLE "inferred_users" (
    "id" BIGINT,
    "email" VARCHAR(320),
    "active" BOOLEAN,
    "created_at" TIMESTAMP,
    "data" BYTEA,
    "score" NUMERIC(10, 2)
)`, sql)
}

func TestDDLColumnUsesExplicitType(t *testing.T) {
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.CreateTable("manual_users").
		Columns(
			ddl.Column(users.Email, ddl.Text()),
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
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.CreateTable(users).
		Columns(
			ddl.Column(users.ID, ddl.BigSerial()).PrimaryKey(),
			ddl.Column(users.Email, ddl.VarChar(255)).NotNull().Unique(),
		).
		Render(dialect.MySQL{})

	require.NoError(t, err)
	assert.Equal(t, "CREATE TABLE `users` (\n    `id` BIGINT AUTO_INCREMENT PRIMARY KEY,\n    `email` VARCHAR(255) NOT NULL UNIQUE\n)", sql)
}

func TestDDLAlterTablePostgreSQL(t *testing.T) {
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.AlterTable(users).
		AddColumn(ddl.Column("nickname", ddl.VarChar(64)).Default("")).
		SetNotNull(users.Email).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, "ALTER TABLE \"users\"\nADD COLUMN \"nickname\" VARCHAR(64) DEFAULT '';\nALTER TABLE \"users\"\nALTER COLUMN \"email\" SET NOT NULL", sql)
}

func TestDDLAlterTableSQLiteUnsupported(t *testing.T) {
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.AlterTable(users).
		AlterColumnType(users.Email, ddl.Text()).
		Render(dialect.SQLite{})

	assert.Empty(t, sql)
	assert.EqualError(t, err, "SQLite dialect does not support ALTER COLUMN TYPE")
}

func TestDDLCreateIndexPostgreSQL(t *testing.T) {
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.CreateIndex("users_email_active_idx").
		Unique().
		IfNotExists().
		On(users, users.Email).
		Where(users.DeletedAt.IsNull()).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE UNIQUE INDEX IF NOT EXISTS "users_email_active_idx" ON "users" ("email") WHERE "deleted_at" IS NULL`, sql)
}

func TestDDLCreateIndexMySQLUnsupportedPartialIndex(t *testing.T) {
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.CreateIndex("users_email_active_idx").
		On(users, users.Email).
		Where("deleted_at IS NULL").
		Render(dialect.MySQL{})

	assert.Empty(t, sql)
	assert.EqualError(t, err, "MySQL dialect does not support PARTIAL INDEX")
}

func TestDDLDropIndexMySQL(t *testing.T) {
	users := q.MustNewTable[DDLUser]()

	sql, err := ddl.DropIndex("users_email_idx").
		On(users).
		Render(dialect.MySQL{})

	require.NoError(t, err)
	assert.Equal(t, "DROP INDEX `users_email_idx` ON `users`", sql)
}
