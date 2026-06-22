package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdditionalWindowFunctions(t *testing.T) {
	users := q.MustNewTable[User]()

	sql, args := q.Select(
		q.Rank().Over(q.PartitionBy(users.Age)).As("rank"),
		q.DenseRank().Over(q.Window().OrderBy(users.UserName.Desc())).As("dense_rank"),
		q.Lag(users.UserName, 1, "none").Over(q.Window().OrderBy(users.UserName.Asc())).As("prev_name"),
		q.Lead(users.UserName).Over(q.Window().
			OrderBy(users.UserName.Asc()).
			Frame(q.Groups().Between(q.CurrentRow(), q.UnboundedFollowing())),
		).As("next_name"),
		q.Count().Over(q.Window().Frame(q.Rows().Following(1))).As("following_count"),
	).MustRender(dialect.PostgreSQL{})

	assert.Equal(t, `SELECT RANK() OVER (PARTITION BY "table"."userAge") AS "rank", DENSE_RANK() OVER (ORDER BY "table"."user_name" DESC) AS "dense_rank", LAG("table"."user_name", $1, $2) OVER (ORDER BY "table"."user_name" ASC) AS "prev_name", LEAD("table"."user_name") OVER (ORDER BY "table"."user_name" ASC GROUPS BETWEEN CURRENT ROW AND UNBOUNDED FOLLOWING) AS "next_name", COUNT(*) OVER (ROWS 1 FOLLOWING) AS "following_count"
FROM "table"`, sql)
	assert.Equal(t, []any{1, "none"}, args)
}

func TestCompoundRenderAndRecursiveCTE(t *testing.T) {
	query := q.Select(q.Literal(1)).
		Union(q.Select(q.Literal(2))).
		UnionAll(q.Select(q.Literal(3))).
		OrderBy(q.Literal(1).Desc()).
		Offset(1)

	sql, args, err := query.Render(dialect.PostgreSQL{})
	require.NoError(t, err)
	assert.Equal(t, `(SELECT 1
UNION
SELECT 2)
UNION ALL
SELECT 3
ORDER BY 1 DESC
OFFSET 1`, sql)
	assert.Empty(t, args)

	cte := query.RecursiveCTE("numbers")
	sql, args = q.Select(cte.Column("n")).MustRender(dialect.PostgreSQL{})
	assert.Contains(t, sql, `WITH RECURSIVE "numbers" AS`)
	assert.Empty(t, args)
}

func TestZeroValueCTE(t *testing.T) {
	var cte q.CommonTableExpression
	assert.Empty(t, cte.TableConfig().Name)
	assert.Empty(t, cte.TableRef().Name)

	cte = cte.WithColumns("id").Recursive()
	assert.Equal(t, []string{"id"}, cte.TableRef().CTE.Columns)
	assert.True(t, cte.TableRef().CTE.Recursive)
}
