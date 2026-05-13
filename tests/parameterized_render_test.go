package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectRender_WithArgs(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	query := q.
		Select(UserTable.UserName).
		Where(
			UserTable.UserName.Eq(q.Param(`bob' OR TRUE --`)),
			UserTable.Age.Ge(q.Param(18)),
		)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`SELECT "table"."user_name" FROM "table" `+
			`WHERE "table"."user_name" = $1 AND "table"."userAge" >= $2`,
		sql,
	)
	assert.Equal(t, []any{`bob' OR TRUE --`, 18}, args)
}

func TestCompoundQueryRender_WithArgs(t *testing.T) {
	query := q.
		Select(q.Literal(1)).
		UnionAll(q.Select(q.Param(2))).
		Limit(1)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(t, `SELECT 1 UNION ALL SELECT $1 LIMIT 1`, sql)
	assert.Equal(t, []any{2}, args)
}

func TestCTERender_WithArgs(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	cte := q.
		Select(UserTable.UserName).
		Where(UserTable.Age.Ge(q.Param(18))).
		CTE("adult_users").
		WithColumns("user_name")

	query := q.
		Select(cte.Column("user_name")).
		Where(cte.Column("user_name").Eq(q.Param("Alice")))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`WITH "adult_users" ("user_name") AS (`+
			`SELECT "table"."user_name" FROM "table" WHERE "table"."userAge" >= $1`+
			`) `+
			`SELECT "adult_users"."user_name" FROM "adult_users" `+
			`WHERE "adult_users"."user_name" = $2`,
		sql,
	)
	assert.Equal(t, []any{18, "Alice"}, args)
}

func TestSelectRender_WithQuestionMarkArgs(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	query := q.
		Select(UserTable.UserName).
		Where(UserTable.UserName.Eq(q.Param("Alice")))

	sql, args := query.Render(dialect.BaseDialect{})

	assert.Equal(
		t,
		`SELECT "table"."user_name" FROM "table" WHERE "table"."user_name" = ?`,
		sql,
	)
	assert.Equal(t, []any{"Alice"}, args)
}

func TestSelectRender_WithArgsKeepsConstantsInline(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	query := q.
		Select(q.Literal(1)).
		Where(UserTable.UserName.Eq("Alice"))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`SELECT 1 FROM "table" WHERE "table"."user_name" = $1`,
		sql,
	)
	assert.Equal(t, []any{"Alice"}, args)
}
