package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
)

func TestOracleDialectLimitOffsetVariants(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Select(q.Literal(false), users.UserName).
		Limit(5).
		Offset(10).
		MustRender(dialect.Oracle{})
	assert.Equal(t, `SELECT 0, "table"."user_name"
FROM "table"
OFFSET 10 ROWS FETCH NEXT 5 ROWS ONLY`, sql)
	assert.Empty(t, args)

	sql, args = q.Select(users.UserName).
		Offset(10).
		MustRender(dialect.Oracle{})
	assert.Equal(t, `SELECT "table"."user_name"
FROM "table"
OFFSET 10 ROWS`, sql)
	assert.Empty(t, args)
}

func TestOracleDDLBranches(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		SetDefault("active", false).
		DropDefault("nickname").
		DropNotNull("email").
		Render(dialect.Oracle{})
	assert.NoError(t, err)
	assert.Equal(t, `ALTER TABLE "users"
    MODIFY "active" DEFAULT 0,
    MODIFY "nickname" DEFAULT NULL,
    MODIFY "email" NULL`, sql)

	_, err = ddl.DropTable("users").Restrict().Render(dialect.Oracle{})
	var unsupported dialect.UnsupportedFeatureError
	assert.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "DROP TABLE RESTRICT", unsupported.Feature)
}

func TestSQLServerDialectLimitOffsetVariants(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Select(users.UserName).
		OrderBy(users.UserName.Asc().NullsFirst()).
		Limit(5).
		MustRender(dialect.SQLServer{})
	assert.Equal(t, `SELECT [table].[user_name]
FROM [table]
ORDER BY CASE WHEN [table].[user_name] IS NULL THEN 0 ELSE 1 END, [table].[user_name] ASC
OFFSET 0 ROWS FETCH NEXT 5 ROWS ONLY`, sql)
	assert.Empty(t, args)

	sql, args = q.Select(users.UserName).
		Offset(7).
		MustRender(dialect.SQLServer{})
	assert.Equal(t, `SELECT [table].[user_name]
FROM [table]
OFFSET 7 ROWS`, sql)
	assert.Empty(t, args)
}

func TestSQLServerUnsupportedDDLBranches(t *testing.T) {
	_, err := ddl.DropTable("users").Cascade().Render(dialect.SQLServer{})
	var unsupported dialect.UnsupportedFeatureError
	assert.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "DROP TABLE CASCADE/RESTRICT", unsupported.Feature)

	_, err = ddl.AlterTable("users").DropNotNull("email").Render(dialect.SQLServer{})
	assert.ErrorAs(t, err, &unsupported)
	assert.Equal(t, "ALTER COLUMN NULLABILITY", unsupported.Feature)
}

func TestSQLiteDDLBranches(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		RenameColumn("name", "full_name").
		AddColumn(ddl.Column("age", ddl.Integer())).
		DropColumn("legacy").
		Render(dialect.SQLite{})
	assert.NoError(t, err)
	assert.Equal(t, `ALTER TABLE "users" RENAME COLUMN "name" TO "full_name";
ALTER TABLE "users" ADD COLUMN "age" INTEGER;
ALTER TABLE "users" DROP COLUMN "legacy"`, sql)

	for _, stmt := range []ddl.Renderer{
		ddl.DropTable("users").Cascade(),
		ddl.AlterTable("users").SetDefault("active", true),
		ddl.AlterTable("users").AddConstraint(ddl.Unique("email")),
		ddl.AlterTable("users").DropConstraint("uq_users_email"),
	} {
		_, err := stmt.Render(dialect.SQLite{})
		var unsupported dialect.UnsupportedFeatureError
		assert.ErrorAs(t, err, &unsupported)
	}
}

func TestMySQLDDLBranches(t *testing.T) {
	sql, err := ddl.AlterTable("users").
		AlterColumnType("email", ddl.VarChar(320)).
		DropConstraint("uq_users_email").
		Render(dialect.MySQL{})
	assert.NoError(t, err)
	assert.Equal(t, "ALTER TABLE `users`\n    MODIFY COLUMN `email` VARCHAR(320),\n    DROP `uq_users_email`", sql)

	_, err = ddl.AlterIndex("ix_users_email").Rename("ix_users_email_new").Render(dialect.MySQL{})
	assert.EqualError(t, err, "MySQL requires table name")
}
