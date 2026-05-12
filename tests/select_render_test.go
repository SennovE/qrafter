package tests

import (
	"testing"

	"github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectRender_ParenthesizesLowerPrecedencePredicate(t *testing.T) {
	u := User{}
	require.NoError(t, qrafter.Bind(&u))

	query := qrafter.Select(u.UserName).Where(
		qrafter.And(
			u.UserName.Eq("ABC"),
			qrafter.Or(
				u.Age.Ge("1"),
				qrafter.Const("Test").Eq(u.UserName),
			),
		),
	)

	assert.Equal(
		t,
		`SELECT "table"."user_name" FROM "table" WHERE "table"."user_name" = 'ABC' AND ("table"."userAge" >= '1' OR 'Test' = "table"."user_name")`,
		query.Render(dialect.PostgreSQL{}),
	)
}

func TestSelectRender_ParenthesizesLowerPrecedenceExpression(t *testing.T) {
	u := User{}
	require.NoError(t, qrafter.Bind(&u))

	query := qrafter.Select(
		u.Age.Add(1).Mul(2),
	)

	assert.Equal(
		t,
		`SELECT ("table"."userAge" + 1) * 2 FROM "table"`,
		query.Render(dialect.PostgreSQL{}),
	)
}

func TestSelectRender_ParenthesizesRightPeerForNonAssociativeExpression(t *testing.T) {
	query := qrafter.Select(
		qrafter.Const(10).Sub(qrafter.Const(7).Sub(3)),
	)

	assert.Equal(t, `SELECT 10 - (7 - 3)`, query.Render(dialect.PostgreSQL{}))
}

func TestSelectRender_Join(t *testing.T) {
	u := User{}
	require.NoError(t, qrafter.Bind(&u))

	manager, err := qrafter.TableAlias(u, "manager")
	require.NoError(t, err)

	query := qrafter.Select(u.UserName, manager.UserName).
		Join(manager, u.Age.Eq(manager.Age)).
		Where(manager.UserName.Eq("Bob"))

	assert.Equal(
		t,
		`SELECT "table"."user_name", "manager"."user_name" FROM "table" JOIN "table" AS "manager" ON "table"."userAge" = "manager"."userAge" WHERE "manager"."user_name" = 'Bob'`,
		query.Render(dialect.PostgreSQL{}),
	)
}

func TestSelectRender_LeftJoin(t *testing.T) {
	u := User{}
	require.NoError(t, qrafter.Bind(&u))

	manager, err := qrafter.TableAlias(u, "manager")
	require.NoError(t, err)

	query := qrafter.
		Select(u.UserName).
		LeftJoin(
			manager,
			u.Age.Eq(manager.Age),
		)

	assert.Equal(
		t,
		`SELECT "table"."user_name" FROM "table" LEFT JOIN "table" AS "manager" ON "table"."userAge" = "manager"."userAge"`,
		query.Render(dialect.PostgreSQL{}),
	)
}

func TestSelectRender_GroupBy(t *testing.T) {
	u := User{}
	require.NoError(t, qrafter.Bind(&u))

	query := qrafter.Select(u.UserName, u.Age.Add(1)).
		GroupBy(u.UserName).
		Limit(10)

	assert.Equal(
		t,
		`SELECT "table"."user_name", "table"."userAge" + 1 FROM "table" GROUP BY "table"."user_name" LIMIT 10`,
		query.Render(dialect.PostgreSQL{}),
	)
}

func TestSelectRender_GroupByJoinedTable(t *testing.T) {
	u := User{}
	require.NoError(t, qrafter.Bind(&u))

	manager, err := qrafter.TableAlias(u, "manager")
	require.NoError(t, err)

	query := qrafter.Select(manager.UserName).
		Join(manager, u.Age.Eq(manager.Age)).
		GroupBy(manager.UserName)

	assert.Equal(
		t,
		`SELECT "manager"."user_name" FROM "table" JOIN "table" AS "manager" ON "table"."userAge" = "manager"."userAge" GROUP BY "manager"."user_name"`,
		query.Render(dialect.PostgreSQL{}),
	)
}
