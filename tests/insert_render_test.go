package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
)

func TestInsertRender_Basic(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	tests := []struct {
		name    string
		query   q.InsertQuery
		wantSQL string
		args    []any
	}{
		{
			"Values",
			q.Insert(UserTable).
				Columns(UserTable.UserName, UserTable.Age).
				Values("Alice", "18"),
			`INSERT INTO "table" ("user_name", "userAge")
VALUES ($1, $2)`,
			[]any{"Alice", "18"},
		},
		{
			"Multiple values with returning",
			q.Insert(UserTable).
				Columns(UserTable.UserName, UserTable.Age).
				Values("Alice", "18").
				Values("Bob", "21").
				Returning(UserTable.UserName),
			`INSERT INTO "table" ("user_name", "userAge")
VALUES ($1, $2), ($3, $4)
RETURNING "table"."user_name"`,
			[]any{"Alice", "18", "Bob", "21"},
		},
		{
			"Values rows",
			q.Insert(UserTable).
				Columns(UserTable.UserName, UserTable.Age).
				ValuesRows([][]any{
					{"Alice", "18"},
					{"Bob", q.Default()},
				}),
			`INSERT INTO "table" ("user_name", "userAge")
VALUES ($1, $2), ($3, DEFAULT)`,
			[]any{"Alice", "18", "Bob"},
		},
		{
			"Set",
			q.Insert(UserTable).
				Set(UserTable.UserName, "Alice").
				Set(UserTable.Age, q.Default()),
			`INSERT INTO "table" ("user_name", "userAge")
VALUES ($1, DEFAULT)`,
			[]any{"Alice"},
		},
		{
			"Default values",
			q.Insert(UserTable).
				DefaultValues(),
			`INSERT INTO "table"
DEFAULT VALUES`,
			nil,
		},
		{
			"Insert from select",
			q.Insert(UserTable).
				Columns(UserTable.UserName, UserTable.Age).
				FromSelect(
					q.Select(UserTable.UserName, UserTable.Age).
						Where(UserTable.Age.Ge("18")),
				),
			`INSERT INTO "table" ("user_name", "userAge")
SELECT "table"."user_name", "table"."userAge"
FROM "table"
WHERE "table"."userAge" >= $1`,
			[]any{"18"},
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

func TestInsertRender_ValuesFrom(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	UserTable.UserName.Set("Alice")
	UserTable.Age.Set("18")

	query := q.
		Insert(UserTable).
		ValuesFrom(UserTable)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(t, `INSERT INTO "table" ("user_name", "userAge")
VALUES ($1, $2)`, sql)
	assert.Equal(t, []any{"Alice", "18"}, args)
}

func TestInsertRender_ValuesFromWithSelectedColumns(t *testing.T) {
	UserTable := q.MustNewTable[User]()
	UserTable.UserName.Set("Alice")

	query := q.
		Insert(UserTable).
		Columns(UserTable.UserName).
		ValuesFrom(UserTable)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(t, `INSERT INTO "table" ("user_name")
VALUES ($1)`, sql)
	assert.Equal(t, []any{"Alice"}, args)
}

func TestInsertRender_ValuesRowsFromSlice(t *testing.T) {
	first := q.MustNewTable[User]()
	first.UserName.Set("Alice")
	first.Age.Set("18")

	second := q.MustNewTable[User]()
	second.UserName.Set("Bob")
	second.Age.Set("21")

	query := q.
		Insert(first).
		ValuesRowsFrom([]User{first, second})

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`INSERT INTO "table" ("user_name", "userAge")
VALUES ($1, $2), ($3, $4)`,
		sql,
	)
	assert.Equal(t, []any{"Alice", "18", "Bob", "21"}, args)
}

func TestInsertRender_ValuesRowsFromPointerSliceWithSelectedColumns(t *testing.T) {
	first := q.MustNewTable[User]()
	first.UserName.Set("Alice")
	first.Age.Set("18")

	second := q.MustNewTable[User]()
	second.UserName.Set("Bob")
	second.Age.Set("21")

	rows := []*User{&first, &second}

	query := q.
		Insert(first).
		Columns(first.Age, first.UserName).
		ValuesRowsFrom(&rows)

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`INSERT INTO "table" ("userAge", "user_name")
VALUES ($1, $2), ($3, $4)`,
		sql,
	)
	assert.Equal(t, []any{"18", "Alice", "21", "Bob"}, args)
}

func TestInsertRender_FromSelectWithCTE(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	cte := q.
		Select(q.Param("Alice"), q.Param("18")).
		CTE("new_users").
		WithColumns("user_name", "age")

	query := q.
		Insert(UserTable).
		Columns(UserTable.UserName, UserTable.Age).
		FromSelect(q.Select(cte.Column("user_name"), cte.Column("age")))

	sql, args := query.Render(dialect.PostgreSQL{})

	assert.Equal(
		t,
		`WITH "new_users" ("user_name", "age") AS (
    SELECT $1, $2
)
INSERT INTO "table" ("user_name", "userAge")
SELECT "new_users"."user_name", "new_users"."age"
FROM "new_users"`,
		sql,
	)
	assert.Equal(t, []any{"Alice", "18"}, args)
}

func TestInsertRender_WithQuestionMarkArgs(t *testing.T) {
	UserTable := q.MustNewTable[User]()

	query := q.
		Insert(UserTable).
		Columns(UserTable.UserName).
		Values("Alice")

	sql, args := query.Render(dialect.BaseDialect{})

	assert.Equal(t, `INSERT INTO "table" ("user_name")
VALUES (?)`, sql)
	assert.Equal(t, []any{"Alice"}, args)
}
