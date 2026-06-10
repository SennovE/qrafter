package tests

import (
	"strings"
	"testing"

	"github.com/SennovE/qrafter/ddl"
	"github.com/SennovE/qrafter/dialect"
	"github.com/SennovE/qrafter/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonTableConfigToSchema(t *testing.T) {
	schema := migrations.TableConfigToSchema[Person](dialect.PostgreSQL{})
	table := schema.Tables[0]

	require.Equal(t, "public", table.Schema)
	require.Equal(t, "users", table.Name)
	assert.Equal(t, []string{
		"id",
		"org_id",
		"user_name",
		"email",
		"display_name",
		"age",
		"status",
		"is_admin",
		"is_verified",
		"profile",
		"last_login",
		"time_created",
		"time_updated",
		"deleted_at",
	}, schemaColumnNames(table.Columns))

	assertColumn(t, &table, "id", ddl.BigInt(), true, "")
	assertColumn(t, &table, "org_id", ddl.BigInt(), true, "")
	assertColumn(t, &table, "user_name", ddl.VarChar(64), true, "")
	assertColumn(t, &table, "email", ddl.Text(), true, "")
	assertColumn(t, &table, "display_name", ddl.VarChar(120), true, "")
	assertColumn(t, &table, "age", ddl.Integer(), false, "")
	assertColumn(t, &table, "status", ddl.Text(), true, "'pending'")
	assertColumn(t, &table, "is_admin", ddl.Boolean(), true, "false")
	assertColumn(t, &table, "is_verified", ddl.Boolean(), true, "FALSE")
	assertColumn(t, &table, "profile", ddl.JSONB(), true, "'{}'::jsonb")
	assertColumn(t, &table, "last_login", ddl.TimestampTZ(), false, "")
	assertColumn(t, &table, "time_created", ddl.TimestampTZ(), true, "now()")
	assertColumn(t, &table, "time_updated", ddl.TimestampTZ(), true, "now()")
	assertColumn(t, &table, "deleted_at", ddl.TimestampTZ(), false, "")

	require.Len(t, table.Constraints, 6)
	assertConstraint(t, &table, migrations.ConstraintPrimaryKey, "pk_users", []string{"id"}, "")
	assertConstraint(t, &table, migrations.ConstraintUnique, "", []string{"user_name"}, "")
	assertConstraint(t, &table, migrations.ConstraintUnique, "uq_users_org_email", []string{"org_id", "email"}, "")
	assertConstraint(t, &table, migrations.ConstraintCheck, "ck_users_age_valid", nil, `"age" IS NULL OR "age" >= 0`)
	assertConstraint(t, &table, migrations.ConstraintCheck, "", nil, "status IN ('pending', 'active', 'blocked', 'deleted')")

	fk := requireConstraint(t, &table, migrations.ConstraintForeignKey, "fk_users_org")
	assert.Equal(t, []string{"org_id"}, fk.Columns)
	assert.Equal(t, "public", fk.Reference.Schema)
	assert.Equal(t, "organizations", fk.Reference.TableName)
	assert.Equal(t, []string{"ID"}, fk.Reference.Columns)
	assert.Equal(t, ddl.Restrict, fk.OnDelete)
	assert.Equal(t, ddl.Cascade, fk.OnUpdate)

	require.Len(t, table.Indexes, 6)
	assertIndex(t, &table, "ix_users_org_id", false, "", []string{`"org_id"`}, "", "")
	assertIndex(t, &table, "ix_users_status", false, "", []string{`"status"`}, "", "")
	assertIndex(
		t,
		&table,
		"ix_users_org_active",
		false,
		"",
		[]string{`"org_id"`, `"user_name"`},
		"deleted_at IS NULL AND status = 'active'",
		"",
	)
	assertIndex(
		t,
		&table,
		"ux_users_org_active",
		true,
		"",
		[]string{`"org_id"`, `"user_name"`},
		`"deleted_at" IS NOT NULL`,
		"",
	)
	assertIndex(
		t,
		&table,
		"ux_users_org_lower_email_active",
		false,
		"",
		[]string{`"org_id" NULLS FIRST`, `lower("email")`},
		"",
		"",
	)
	assertIndex(
		t,
		&table,
		"ix_users_search_text_trgm",
		false,
		ddl.IndexGin,
		[]string{"search_text gin_trgm_ops"},
		"",
		"",
	)
}

func TestPersonSchemaRendersDDL(t *testing.T) {
	schema := migrations.TableConfigToSchema[Person](dialect.PostgreSQL{})

	sqlText, err := schema.DDL().Render(dialect.PostgreSQL{})
	require.NoError(t, err)

	for _, fragment := range []string{
		`"status" TEXT NOT NULL DEFAULT 'pending'`,
		`"is_admin" BOOLEAN NOT NULL DEFAULT false`,
		`"is_verified" BOOLEAN NOT NULL DEFAULT FALSE`,
		`"profile" JSONB NOT NULL DEFAULT '{}'::jsonb`,
		`CONSTRAINT "pk_users" PRIMARY KEY ("id")`,
		`CONSTRAINT "uq_users_org_email" UNIQUE ("org_id", "email")`,
		`CONSTRAINT "fk_users_org" FOREIGN KEY ("org_id") REFERENCES "organizations" ("ID") ON DELETE RESTRICT ON UPDATE CASCADE`,
		`CREATE INDEX "ix_users_status" ON "users" ("status")`,
		`CREATE INDEX "ix_users_search_text_trgm" ON "users" USING gin (search_text gin_trgm_ops)`,
	} {
		assert.Contains(t, sqlText, fragment)
	}
}

func schemaColumnNames(columns []migrations.Column) []string {
	out := make([]string, len(columns))
	for i := range columns {
		out[i] = columns[i].Name
	}
	return out
}

func assertColumn(t *testing.T, table *migrations.Table, name string, typ ddl.Type, notNull bool, defaultExpr string) {
	t.Helper()

	column := requireColumn(t, table, name)
	assert.Equal(t, typ.Name, column.Type.Name, name)
	assert.Equal(t, notNull, column.NotNull, name)
	assert.Equal(t, defaultExpr != "", column.HasDefault, name)
	assert.Equal(t, normalizeSchemaSQL(defaultExpr), normalizeSchemaSQL(column.DefaultExpr), name)
}

func requireColumn(t *testing.T, table *migrations.Table, name string) migrations.Column {
	t.Helper()
	for i := range table.Columns {
		if table.Columns[i].Name == name {
			return table.Columns[i]
		}
	}
	t.Fatalf("column %q not found", name)
	return migrations.Column{}
}

func assertConstraint(
	t *testing.T,
	table *migrations.Table,
	kind migrations.ConstraintKind,
	name string,
	columns []string,
	checkExpr string,
) {
	t.Helper()

	constraint := requireConstraint(t, table, kind, name)
	assert.Equal(t, columns, constraint.Columns)
	assert.Equal(t, normalizeSchemaSQL(checkExpr), normalizeSchemaSQL(constraint.CheckExpr))
	assert.Equal(t, "public", constraint.Schema)
	assert.Equal(t, "users", constraint.TableName)
}

func requireConstraint(
	t *testing.T,
	table *migrations.Table,
	kind migrations.ConstraintKind,
	name string,
) migrations.Constraint {
	t.Helper()
	for i := range table.Constraints {
		if table.Constraints[i].Kind == kind && table.Constraints[i].Name == name {
			return table.Constraints[i]
		}
	}
	t.Fatalf("constraint %s/%q not found", kind, name)
	return migrations.Constraint{}
}

func assertIndex(
	t *testing.T,
	table *migrations.Table,
	name string,
	unique bool,
	method ddl.IndexMethod,
	keys []string,
	predicate string,
	tablespace string,
) {
	t.Helper()

	index := requireIndex(t, table, name)
	assert.Equal(t, "public", index.Schema)
	assert.Equal(t, "public", index.TableSchema)
	assert.Equal(t, "users", index.TableName)
	assert.Equal(t, unique, index.Unique)
	assert.Equal(t, method, index.Method)
	assert.Equal(t, tablespace, index.Tablespace)
	assert.Equal(t, normalizeSchemaSQL(predicate), normalizeSchemaSQL(index.Predicate))
	assert.Equal(t, keys, indexKeyExpressions(index.Keys))
}

func requireIndex(t *testing.T, table *migrations.Table, name string) migrations.Index {
	t.Helper()
	for i := range table.Indexes {
		if table.Indexes[i].Name == name {
			return table.Indexes[i]
		}
	}
	t.Fatalf("index %q not found", name)
	return migrations.Index{}
}

func indexKeyExpressions(keys []migrations.IndexKey) []string {
	out := make([]string, len(keys))
	for i := range keys {
		out[i] = keys[i].Expression
	}
	return out
}

func normalizeSchemaSQL(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, `"`, "")
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}
