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
		qrafter.Eq(u.UserName, qrafter.Const("ABC")),
		qrafter.Or(
			qrafter.Ge(u.Age, qrafter.Const("1")),
			qrafter.Eq(qrafter.Const("Test"), u.UserName),
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
		qrafter.Mul(
			qrafter.Sum(u.Age, qrafter.Const(1)),
			qrafter.Const(2),
		),
	)

	assert.Equal(
		t,
		`SELECT ("table"."userAge" + 1) * 2 FROM "table"`,
		query.Render(dialect.PostgreSQL{}),
	)
}

func TestSelectRender_ParenthesizesRightPeerForNonAssociativeExpression(t *testing.T) {
	query := qrafter.Select(
		qrafter.Sub(
			qrafter.Const(10),
			qrafter.Sub(qrafter.Const(7), qrafter.Const(3)),
		),
	)

	assert.Equal(t, `SELECT 10 - (7 - 3)`, query.Render(dialect.PostgreSQL{}))
}
