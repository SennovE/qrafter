package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectRender_Basic(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	tests := []struct {
		name    string
		query   q.SelectQuery
		wantSQL string
		args    []any
	}{
		{
			"Lower priority is indicated in brackets for logical expressions",
			q.Select(UserTable.UserName).
				Where(
					q.And(
						UserTable.UserName.Eq("ABC"),
						q.Or(
							UserTable.Age.Ge("1"),
							q.Param("Test").Eq(UserTable.UserName),
						),
					),
				),
			`SELECT "table"."user_name" FROM "table" ` +
				`WHERE "table"."user_name" = $1 AND ("table"."userAge" >= $2 OR $3 = "table"."user_name")`,
			[]any{"ABC", "1", "Test"},
		},
		{
			"Lower priority is indicated in brackets for math expressions",
			q.Select(UserTable.Age.Add(q.Literal(1)).Mul(q.Literal(2))),
			`SELECT ("table"."userAge" + 1) * 2 FROM "table"`,
			nil,
		},
		{
			"The right peer for a non-associative expression is indicated in brackets",
			q.Select(q.Literal(10).Sub(q.Literal(7).Sub(q.Literal(3)))),
			`SELECT 10 - (7 - 3)`,
			nil,
		},
		{
			"Group By",
			q.Select(UserTable.UserName, UserTable.Age.Add(1)).
				GroupBy(UserTable.UserName).
				Limit(10),
			`SELECT "table"."user_name", "table"."userAge" + $1 FROM "table" ` +
				`GROUP BY "table"."user_name" ` +
				`LIMIT 10`,
			[]any{1},
		},
		{
			"Functions",
			q.Select(
				q.Func("LOWER", UserTable.UserName).As("lower_name"),
				q.Func("COALESCE", UserTable.Age, q.Literal("0")),
			).Where(
				q.Func("LOWER", UserTable.UserName).Eq("bob"),
			),
			`SELECT ` +
				`LOWER("table"."user_name") AS "lower_name", ` +
				`COALESCE("table"."userAge", '0') FROM "table" ` +
				`WHERE LOWER("table"."user_name") = $1`,
			[]any{"bob"},
		},
		{
			"Aggregations and Having",
			q.Select(
				UserTable.UserName,
				q.Count().As("users_count"),
				q.Count(q.Distinct(UserTable.Age)).As("distinct_ages"),
				q.Max(UserTable.Age).As("max_age"),
			).GroupBy(
				UserTable.UserName,
			).Having(
				q.Count().Gt(1),
				q.Max(UserTable.Age).Ge("18"),
			).Limit(10),
			`SELECT ` +
				`"table"."user_name", ` +
				`COUNT(*) AS "users_count", ` +
				`COUNT(DISTINCT "table"."userAge") AS "distinct_ages", ` +
				`MAX("table"."userAge") AS "max_age" FROM "table" ` +
				`GROUP BY "table"."user_name" ` +
				`HAVING COUNT(*) > $1 AND MAX("table"."userAge") >= $2 ` +
				`LIMIT 10`,
			[]any{1, "18"},
		},
		{
			"Order By",
			q.Select(UserTable.UserName).
				OrderBy(
					UserTable.UserName.Asc(),
					UserTable.Age.Desc().NullsLast(),
				).
				Limit(10).
				Offset(10),
			`SELECT "table"."user_name" FROM "table" ` +
				`ORDER BY "table"."user_name" ASC, "table"."userAge" DESC NULLS LAST ` +
				`LIMIT 10 OFFSET 10`,
			nil,
		},
		{
			"Order By with Aggregations",
			q.Select(UserTable.UserName, q.Count().As("users_count")).
				GroupBy(UserTable.UserName).
				Having(q.Count().Gt(q.Literal(1))).
				OrderBy(q.Count().Desc()),
			`SELECT "table"."user_name", COUNT(*) AS "users_count" FROM "table" ` +
				`GROUP BY "table"."user_name" ` +
				`HAVING COUNT(*) > 1 ` +
				`ORDER BY COUNT(*) DESC`,
			nil,
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

func TestSelectRender_WithJoin(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	ManagerTable, err := q.TableAlias(UserTable, "manager")
	require.NoError(t, err)

	tests := []struct {
		name    string
		query   q.SelectQuery
		wantSQL string
		args    []any
	}{
		{
			"Basic Join",
			q.Select(UserTable.UserName, ManagerTable.UserName).
				Join(ManagerTable, UserTable.Age.Eq(ManagerTable.Age)).
				Where(ManagerTable.UserName.Eq("Bob")),
			`SELECT "table"."user_name", "manager"."user_name" FROM "table" ` +
				`JOIN "table" AS "manager" ON "table"."userAge" = "manager"."userAge" ` +
				`WHERE "manager"."user_name" = $1`,
			[]any{"Bob"},
		},
		{
			"Left Join",
			q.Select(UserTable.UserName).
				LeftJoin(ManagerTable, UserTable.Age.Eq(ManagerTable.Age)),
			`SELECT "table"."user_name" FROM "table" ` +
				`LEFT JOIN "table" AS "manager" ON "table"."userAge" = "manager"."userAge"`,
			nil,
		},
		{
			"Group By with Join",
			q.Select(ManagerTable.UserName).
				Join(ManagerTable, UserTable.Age.Eq(ManagerTable.Age)).
				GroupBy(ManagerTable.UserName),
			`SELECT "manager"."user_name" FROM "table" ` +
				`JOIN "table" AS "manager" ON "table"."userAge" = "manager"."userAge" ` +
				`GROUP BY "manager"."user_name"`,
			nil,
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

func TestSelectRender_WithUnion(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	tests := []struct {
		name    string
		query   q.CompoundQuery
		wantSQL string
		args    []any
	}{
		{
			"Union",
			q.Select(q.Literal(1)).
				Union(q.Select(q.Literal(2))),
			`SELECT 1 UNION SELECT 2`,
			nil,
		},
		{
			"Union All",
			q.Select(UserTable.UserName).
				Where(UserTable.Age.Lt(18)).
				UnionAll(
					q.Select(UserTable.UserName).
						Where(UserTable.Age.Ge(65)),
				),
			`SELECT "table"."user_name" FROM "table" WHERE "table"."userAge" < $1 ` +
				`UNION ALL ` +
				`SELECT "table"."user_name" FROM "table" WHERE "table"."userAge" >= $2`,
			[]any{18, 65},
		},
		{
			"Union with final limit",
			q.Select(q.Literal(1)).
				UnionAll(q.Select(q.Literal(2))).
				Limit(1),
			`SELECT 1 UNION ALL SELECT 2 LIMIT 1`,
			nil,
		},
		{
			"Union with local limit in right arm",
			q.Select(q.Literal(1)).
				UnionAll(
					q.Select(q.Literal(2)).
						Limit(1),
				),
			`SELECT 1 UNION ALL (SELECT 2 LIMIT 1)`,
			nil,
		},
		{
			"Union with local limit in left arm and final limit",
			q.Select(q.Literal(1)).
				Limit(1).
				UnionAll(q.Select(q.Literal(2))).
				Limit(10),
			`(SELECT 1 LIMIT 1) UNION ALL SELECT 2 LIMIT 10`,
			nil,
		},
		{
			"Union with local limit in compound left arm",
			q.Select(q.Literal(1)).
				Union(q.Select(q.Literal(2))).
				Limit(1).
				UnionAll(q.Select(q.Literal(3))).
				Limit(10),
			`(SELECT 1 UNION SELECT 2 LIMIT 1) UNION ALL SELECT 3 LIMIT 10`,
			nil,
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

func TestSelectRender_WithWindowFunctions(t *testing.T) {
	UserTable := User{}
	require.NoError(t, q.Bind(&UserTable))

	tests := []struct {
		name    string
		query   q.SelectQuery
		wantSQL string
	}{
		{
			"Row number with partition and order",
			q.Select(
				UserTable.UserName,
				q.RowNumber().
					Over(q.Window().
						PartitionBy(UserTable.Age).
						OrderBy(UserTable.UserName.Asc()),
					).
					As("rn"),
			),
			`SELECT "table"."user_name", ` +
				`ROW_NUMBER() OVER (PARTITION BY "table"."userAge" ORDER BY "table"."user_name" ASC) AS "rn" ` +
				`FROM "table"`,
		},
		{
			"Aggregate window",
			q.Select(
				UserTable.UserName,
				q.Count().Over(q.PartitionBy(UserTable.Age)).As("age_count"),
			),
			`SELECT "table"."user_name", COUNT(*) OVER (PARTITION BY "table"."userAge") AS "age_count" ` +
				`FROM "table"`,
		},
		{
			"Window order contributes source table",
			q.Select(
				q.RowNumber().
					Over(q.Window().OrderBy(UserTable.Age.Desc())).
					As("rn"),
			),
			`SELECT ROW_NUMBER() OVER (ORDER BY "table"."userAge" DESC) AS "rn" FROM "table"`,
		},
		{
			"Empty over clause",
			q.Select(q.Count().Over().As("total")),
			`SELECT COUNT(*) OVER () AS "total"`,
		},
		{
			"Window frame between bounds",
			q.Select(
				UserTable.UserName,
				q.Sum(UserTable.Age).
					Over(q.Window().
						OrderBy(UserTable.UserName.Asc()).
						Frame(q.Rows().Between(q.UnboundedPreceding(), q.CurrentRow())),
					).
					As("running_age"),
			),
			`SELECT "table"."user_name", ` +
				`SUM("table"."userAge") OVER (` +
				`ORDER BY "table"."user_name" ASC ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW` +
				`) AS "running_age" FROM "table"`,
		},
		{
			"Window frame single bound",
			q.Select(
				q.Count().
					Over(q.Window().
						OrderBy(UserTable.Age.Asc()).
						Frame(q.Rows().Preceding(1)),
					).
					As("nearby_count"),
			),
			`SELECT COUNT(*) OVER (ORDER BY "table"."userAge" ASC ROWS 1 PRECEDING) AS "nearby_count" FROM "table"`,
		},
		{
			"Window frame string offset",
			q.Select(
				q.Count().
					Over(q.Window().
						OrderBy(UserTable.UserName.Asc()).
						Frame(q.Range().Between(q.Preceding("INTERVAL '1 day'"), q.CurrentRow())),
					).
					As("daily_count"),
			),
			`SELECT COUNT(*) OVER (` +
				`ORDER BY "table"."user_name" ASC RANGE BETWEEN INTERVAL '1 day' PRECEDING AND CURRENT ROW` +
				`) AS "daily_count" FROM "table"`,
		},
		{
			"Window frame custom bound",
			q.Select(
				q.Count().
					Over(q.Window().
						OrderBy(UserTable.Age.Asc()).
						Frame(q.Rows().Bound(q.FrameBound("CURRENT ROW"))),
					).
					As("current_count"),
			),
			`SELECT COUNT(*) OVER (ORDER BY "table"."userAge" ASC ROWS CURRENT ROW) AS "current_count" FROM "table"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, args := tt.query.Render(dialect.PostgreSQL{})
			assert.Equal(t, tt.wantSQL, str)
			assert.Empty(t, args)
		})
	}
}
