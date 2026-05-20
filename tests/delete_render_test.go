package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteRender_Basic(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	tests := []struct {
		name    string
		query   q.DeleteQuery
		wantSQL string
		args    []any
	}{
		{
			"Delete all",
			q.Delete(UserTable),
			`DELETE FROM "table"`,
			nil,
		},
		{
			"Where",
			q.Delete(UserTable).
				Where(UserTable.UserName.Eq("Alice"), UserTable.Age.Ge("18")),
			`DELETE FROM "table" WHERE "table"."user_name" = $1 AND "table"."userAge" >= $2`,
			[]any{"Alice", "18"},
		},
		{
			"Returning",
			q.Delete(UserTable).
				Where(UserTable.UserName.Eq("Alice")).
				Returning(UserTable.UserName),
			`DELETE FROM "table" WHERE "table"."user_name" = $1 RETURNING "table"."user_name"`,
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

func TestDeleteRender_WithUsing(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	ManagerTable, err := q.TableAlias(UserTable, "manager")
	require.NoError(t, err)

	query := q.
		Delete(UserTable).
		Using(ManagerTable).
		Where(
			UserTable.Age.Eq(ManagerTable.Age),
			ManagerTable.UserName.Eq("Bob"),
		)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`DELETE FROM "table" USING "table" AS "manager" `+
			`WHERE "table"."userAge" = "manager"."userAge" AND "manager"."user_name" = $1`,
		sql,
	)
	assert.Equal(t, []any{"Bob"}, args)
}

func TestDeleteRender_AutoUsingFromWhere(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	ManagerTable, err := q.TableAlias(UserTable, "manager")
	require.NoError(t, err)

	query := q.
		Delete(UserTable).
		Where(
			UserTable.Age.Eq(ManagerTable.Age),
			ManagerTable.UserName.Eq("Bob"),
		)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`DELETE FROM "table" USING "table" AS "manager" `+
			`WHERE "table"."userAge" = "manager"."userAge" AND "manager"."user_name" = $1`,
		sql,
	)
	assert.Equal(t, []any{"Bob"}, args)
}

func TestDeleteRender_WithCTEUsing(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	cte := q.
		Select(q.Param("Alice")).
		CTE("doomed_users").
		WithColumns("user_name")

	query := q.
		Delete(UserTable).
		Using(cte).
		Where(UserTable.UserName.Eq(cte.Column("user_name")))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`WITH "doomed_users" ("user_name") AS (SELECT $1) `+
			`DELETE FROM "table" USING "doomed_users" `+
			`WHERE "table"."user_name" = "doomed_users"."user_name"`,
		sql,
	)
	assert.Equal(t, []any{"Alice"}, args)
}

func TestDeleteRender_WithQuestionMarkArgs(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	query := q.
		Delete(UserTable).
		Where(UserTable.UserName.Eq("Alice"))

	sql, args := query.Render(dialect.BaseDialect{})

	assert.Equal(t, `DELETE FROM "table" WHERE "table"."user_name" = ?`, sql)
	assert.Equal(t, []any{"Alice"}, args)
}
