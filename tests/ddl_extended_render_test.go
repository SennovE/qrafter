package tests

import (
	"testing"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDDLAlterTableAllOperationsPostgreSQL(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		RenameColumn("name", "full_name").
		AddColumnIfNotExists(ddl.Column("nickname", ddl.VarChar(32)).Default("anon")).
		DropColumn("legacy").
		DropColumnIfExists("old_flag").
		SetNotNull("email").
		DropNotNull("middle_name").
		SetDefault("active", true).
		SetDefaultExpr("created_at", "now()").
		DropDefault("updated_at").
		AddConstraint(ddl.Check(ddl.Col("age").Ge(0)).Named("ck_users_age")).
		DropConstraint("old_constraint").
		DropConstraintIfExists("maybe_constraint").
		RenameConstraint("old_name", "new_name").
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `ALTER TABLE "users"
    RENAME COLUMN "name" TO "full_name",
    ADD COLUMN IF NOT EXISTS "nickname" VARCHAR(32) DEFAULT 'anon',
    DROP COLUMN "legacy",
    DROP COLUMN IF EXISTS "old_flag",
    ALTER COLUMN "email" SET NOT NULL,
    ALTER COLUMN "middle_name" DROP NOT NULL,
    ALTER COLUMN "active" SET DEFAULT TRUE,
    ALTER COLUMN "created_at" SET DEFAULT now(),
    ALTER COLUMN "updated_at" DROP DEFAULT,
    ADD CONSTRAINT "ck_users_age" CHECK ("age" >= 0),
    DROP CONSTRAINT "old_constraint",
    DROP CONSTRAINT IF EXISTS "maybe_constraint",
    RENAME CONSTRAINT "old_name" TO "new_name"`, sql)
}

func TestDDLIndexOptionsPostgreSQL(t *testing.T) {
	sql, err := ddl.CreateIndex("ix_users_search").
		Unique().
		Concurrently().
		IfNotExists().
		Using(ddl.IndexGin).
		On(
			"users",
			ddl.KeyCol("email").PrefixLength(10).Asc().NullsFirst(),
			ddl.Key(ddl.Func("lower", ddl.Col("name"))),
		).
		Include(ddl.Col("id"), ddl.RawExpr("created_at")).
		With("fastupdate", true).
		With("fillfactor", 90).
		NullsNotDistinct().
		Tablespace("fastspace").
		Where(ddl.Col("deleted_at").IsNull()).
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS "ix_users_search" ON "users" USING gin ("email"(10) ASC NULLS FIRST, lower("name")) INCLUDE ("id", created_at) NULLS NOT DISTINCT WITH (fastupdate = ON, fillfactor = 90) TABLESPACE "fastspace" WHERE "deleted_at" IS NULL`, sql)
}

func TestDDLIndexOptionsByDialect(t *testing.T) {
	sql, err := ddl.CreateIndex("ix_users_email").
		Clustered().
		OnCols("users", "email").
		Render(dialect.SQLServer{})
	require.NoError(t, err)
	assert.Equal(t, `CREATE CLUSTERED INDEX [ix_users_email] ON [users] ([email])`, sql)

	sql, err = ddl.CreateIndex("ix_users_email").
		NonClustered().
		OnCols("users", "email").
		Render(dialect.SQLServer{})
	require.NoError(t, err)
	assert.Equal(t, `CREATE NONCLUSTERED INDEX [ix_users_email] ON [users] ([email])`, sql)

	sql, err = ddl.CreateIndex("ix_users_email").
		Invisible().
		OnCols("users", "email").
		Render(dialect.MySQL{})
	require.NoError(t, err)
	assert.Equal(t, "CREATE INDEX `ix_users_email` ON `users` (`email`) INVISIBLE", sql)
}

func TestDDLDropIndexOptions(t *testing.T) {
	sql, err := ddl.DropIndex("ix_users_email").
		IfExists().
		Concurrently().
		Cascade().
		Render(dialect.PostgreSQL{})

	require.NoError(t, err)
	assert.Equal(t, `DROP INDEX CONCURRENTLY IF EXISTS "ix_users_email" CASCADE`, sql)

	sql, err = ddl.DropIndex("ix_users_email").
		OnTable("users").
		Online().
		Render(dialect.SQLServer{})

	require.NoError(t, err)
	assert.Equal(t, `DROP INDEX [ix_users_email] ON [users] ONLINE`, sql)
}

func TestDDLMustRenderAndPanics(t *testing.T) {
	assert.Equal(t, `DROP TABLE IF EXISTS "users"`, ddl.DropTable("users").IfExists().MustRender(dialect.PostgreSQL{}))
	assert.Panics(t, func() {
		ddl.CreateTable("empty").MustRender(dialect.PostgreSQL{})
	})
	assert.PanicsWithError(t, "foreign key reference is required", func() {
		ddl.CreateTable("bad").Constraints(ddl.ForeignKey("org_id")).MustRender(dialect.PostgreSQL{})
	})
}

func TestDDLStandaloneColumnAndExpressionRendering(t *testing.T) {
	sql, err := ddl.Render(
		dialect.PostgreSQL{},
		ddl.Column("org_id", ddl.Integer()).
			PrimaryKey().
			Unique().
			Check("org_id > 0").
			References("orgs"),
	)
	require.NoError(t, err)
	assert.Equal(t, `"org_id" INTEGER PRIMARY KEY UNIQUE CHECK (org_id > 0) REFERENCES "orgs"`, sql)

	sql, err = ddl.CreateTable("checks").
		Columns(ddl.Column("id", ddl.Integer())).
		Constraints(
			ddl.Check(
				ddl.Or(
					ddl.Col("score").Eq(0),
					ddl.And(
						ddl.Col("score").Gt(0),
						ddl.Col("score").Lt(100),
					),
				),
			),
			ddl.Check(
				ddl.Literal(10).
					Sub(ddl.Literal(7).Sub(3)).
					Div(ddl.Literal(2).Mul(3)).
					Gt(0),
			),
		).
		Render(dialect.PostgreSQL{})
	require.NoError(t, err)
	assert.Contains(t, sql, `CHECK ("score" = 0 OR "score" > 0 AND "score" < 100)`)
	assert.Contains(t, sql, `CHECK ((10 - (7 - 3)) / (2 * 3) > 0)`)
}
