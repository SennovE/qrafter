package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateRender_Basic(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	tests := []struct {
		name    string
		query   q.UpdateQuery
		wantSQL string
		args    []any
	}{
		{
			"Set where",
			q.Update(UserTable).
				Set(UserTable.UserName, "Alice").
				Where(UserTable.Age.Ge("18")),
			`UPDATE "table"
SET "user_name" = $1
WHERE "table"."userAge" >= $2`,
			[]any{"Alice", "18"},
		},
		{
			"Multiple set with returning",
			q.Update(UserTable).
				Set(UserTable.UserName, "Alice").
				Set(UserTable.Age, "18").
				Where(UserTable.UserName.Eq("old")).
				Returning(UserTable.UserName, UserTable.Age),
			`UPDATE "table"
SET "user_name" = $1, "userAge" = $2
WHERE "table"."user_name" = $3
RETURNING "table"."user_name", "table"."userAge"`,
			[]any{"Alice", "18", "old"},
		},
		{
			"Default value",
			q.Update(UserTable).
				Set(UserTable.Age, q.Default()).
				Where(UserTable.UserName.Eq("Alice")),
			`UPDATE "table"
SET "userAge" = DEFAULT
WHERE "table"."user_name" = $1`,
			[]any{"Alice"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, args := tt.query.Render(dialect.PostgreSQL{})
			assert.Equal(t, tt.wantSQL, str)
			assert.Equal(t, tt.args, args)
		})
	}
}

func TestUpdateRender_SetFrom(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	UserTable.UserName.Set("Alice")
	UserTable.Age.Set("18")

	query := q.
		Update(UserTable).
		SetFrom(UserTable).
		Where(UserTable.UserName.Eq("old"))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`UPDATE "table"
SET "user_name" = $1, "userAge" = $2
WHERE "table"."user_name" = $3`,
		sql,
	)
	assert.Equal(t, []any{"Alice", "18", "old"}, args)
}

func TestUpdateRender_WithFrom(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	ManagerTable, err := q.TableAlias(UserTable, "manager")
	require.NoError(t, err)

	query := q.
		Update(UserTable).
		Set(UserTable.Age, ManagerTable.Age).
		From(ManagerTable).
		Where(ManagerTable.UserName.Eq("Bob"))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`UPDATE "table"
SET "userAge" = "manager"."userAge"
FROM "table" AS "manager"
WHERE "manager"."user_name" = $1`,
		sql,
	)
	assert.Equal(t, []any{"Bob"}, args)
}

func TestUpdateRender_AutoFromFromSetAndWhere(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	ManagerTable, err := q.TableAlias(UserTable, "manager")
	require.NoError(t, err)

	query := q.
		Update(UserTable).
		Set(UserTable.Age, ManagerTable.Age).
		Where(UserTable.UserName.Eq(ManagerTable.UserName))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`UPDATE "table"
SET "userAge" = "manager"."userAge"
FROM "table" AS "manager"
WHERE "table"."user_name" = "manager"."user_name"`,
		sql,
	)
	assert.Empty(t, args)
}

func TestUpdateRender_WithCTEFrom(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	cte := q.
		Select(q.Param("Alice"), q.Param("18")).
		CTE("source_users").
		WithColumns("user_name", "age")

	query := q.
		Update(UserTable).
		Set(UserTable.Age, cte.Column("age")).
		From(cte).
		Where(UserTable.UserName.Eq(cte.Column("user_name")))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`WITH "source_users" ("user_name", "age") AS (
    SELECT $1, $2
)
UPDATE "table"
SET "userAge" = "source_users"."age"
FROM "source_users"
WHERE "table"."user_name" = "source_users"."user_name"`,
		sql,
	)
	assert.Equal(t, []any{"Alice", "18"}, args)
}

func TestUpdateRender_WithQuestionMarkArgs(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	query := q.
		Update(UserTable).
		Set(UserTable.UserName, "Alice").
		Where(UserTable.Age.Ge("18"))

	sql, args := query.Render(dialect.BaseDialect{})

	assert.Equal(t, `UPDATE "table"
SET "user_name" = ?
WHERE "table"."userAge" >= ?`,
		sql,
	)
	assert.Equal(t, []any{"Alice", "18"}, args)
}
