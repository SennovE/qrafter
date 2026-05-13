package tests

import (
	"testing"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Orders struct {
	UserID q.Column[int]
	Amount q.Column[int]
	Status q.Column[string]
}

func (Orders) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "orders",
	}
}

type Users struct {
	ID   q.Column[int]
	Name q.Column[string]
}

func (Users) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "users",
	}
}

type Numbers struct {
	N q.Column[int] `db:"n"`
}

func (Numbers) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "numbers",
	}
}

func TestSelectRender_WithCTE(t *testing.T) {
	OrdersTable := Orders{}
	require.NoError(t, q.Bind(&OrdersTable))
	UsersTable := Users{}
	require.NoError(t, q.Bind(&UsersTable))

	cte := q.
		Select(OrdersTable.UserID, q.Sum(OrdersTable.Amount)).
		Where(OrdersTable.Status.Eq("paid")).
		GroupBy(OrdersTable.UserID).
		CTE("total_amounts").
		WithColumns("user_id", "total")

	t.Run("Query with CTE binded to struct", func(t *testing.T) {
		type TotalAmountsTable struct {
			UserID q.Column[int]
			Total  q.Column[int]
		}

		TotalAmountsCTE := TotalAmountsTable{}
		err := cte.Bind(&TotalAmountsCTE)

		require.NoError(t, err)

		query := q.
			Select(UsersTable.Name, TotalAmountsCTE.Total).
			Join(cte, UsersTable.ID.Eq(TotalAmountsCTE.UserID)).
			Where(TotalAmountsCTE.Total.Gt(100))

		str, args := query.Render(dialect.PostgreSQL{})
		assert.Equal(
			t,
			`WITH "total_amounts" ("user_id", "total") AS (`+
				`SELECT "orders"."user_id", SUM("orders"."amount") FROM "orders" `+
				`WHERE "orders"."status" = 'paid' `+
				`GROUP BY "orders"."user_id"`+
				`) `+
				`SELECT "users"."name", "total_amounts"."total" FROM "users" `+
				`JOIN "total_amounts" ON "users"."id" = "total_amounts"."user_id" `+
				`WHERE "total_amounts"."total" > 100`,
			str,
		)
		assert.Empty(t, args)
	})

	t.Run("Query with CTE with obtaining columns by string name", func(t *testing.T) {
		query := q.
			Select(UsersTable.Name, cte.Column("total")).
			Join(cte, UsersTable.ID.Eq(cte.Column("user_id"))).
			Where(cte.Column("total").Gt(100))

		str, args := query.Render(dialect.PostgreSQL{})
		assert.Equal(
			t,
			`WITH "total_amounts" ("user_id", "total") AS (`+
				`SELECT "orders"."user_id", SUM("orders"."amount") FROM "orders" `+
				`WHERE "orders"."status" = 'paid' `+
				`GROUP BY "orders"."user_id"`+
				`) `+
				`SELECT "users"."name", "total_amounts"."total" FROM "users" `+
				`JOIN "total_amounts" ON "users"."id" = "total_amounts"."user_id" `+
				`WHERE "total_amounts"."total" > 100`,
			str,
		)
		assert.Empty(t, args)
	})
}

func TestSelectRender_WithRecursiveCTE(t *testing.T) {
	t.Run("Recursive CTE method", func(t *testing.T) {
		cte := q.
			Select(q.Literal(1)).
			CTE("numbers").
			Recursive().
			WithColumns("n")

		query := q.Select(cte.Column("n"))

		str, args := query.Render(dialect.PostgreSQL{})
		assert.Equal(
			t,
			`WITH RECURSIVE "numbers" ("n") AS (SELECT 1) `+
				`SELECT "numbers"."n" FROM "numbers"`,
			str,
		)
		assert.Empty(t, args)
	})

	t.Run("Recursive CTE shortcut", func(t *testing.T) {
		cte := q.
			Select(q.Literal(1)).
			RecursiveCTE("numbers").
			WithColumns("n")

		query := q.Select(cte.Column("n"))

		str, args := query.Render(dialect.PostgreSQL{})
		assert.Equal(
			t,
			`WITH RECURSIVE "numbers" ("n") AS (SELECT 1) `+
				`SELECT "numbers"."n" FROM "numbers"`,
			str,
		)
		assert.Empty(t, args)
	})

	t.Run("Recursive CTE with union all", func(t *testing.T) {
		NumbersTable := Numbers{}
		require.NoError(t, q.Bind(&NumbersTable))

		cte := q.
			Select(q.Literal(1)).
			UnionAll(
				q.Select(NumbersTable.N.Add(1)).
					Where(NumbersTable.N.Lt(3)),
			).
			RecursiveCTE("numbers").
			WithColumns("n")

		query := q.Select(cte.Column("n"))

		str, args := query.Render(dialect.PostgreSQL{})
		assert.Equal(
			t,
			`WITH RECURSIVE "numbers" ("n") AS (`+
				`SELECT 1 UNION ALL `+
				`SELECT "numbers"."n" + 1 FROM "numbers" WHERE "numbers"."n" < 3`+
				`) `+
				`SELECT "numbers"."n" FROM "numbers"`,
			str,
		)
		assert.Empty(t, args)
	})
}

func TestSelectRender_WithMultipleCTEs(t *testing.T) {
	cte1 := q.Select(q.Literal(1)).CTE("cte1").WithColumns("c1")
	cte2 := q.Select(cte1.Column("c1")).CTE("cte2").WithColumns("c1")
	query := q.Select(cte1.Column("c1"), cte2.Column("c1")).CrossJoin(cte2)

	str, args := query.Render(dialect.PostgreSQL{})
	assert.Equal(
		t,
		`WITH "cte1" ("c1") AS (SELECT 1), `+
			`"cte2" ("c1") AS (SELECT "cte1"."c1" FROM "cte1") `+
			`SELECT "cte1"."c1", "cte2"."c1" FROM "cte1" CROSS JOIN "cte2"`,
		str,
	)
	assert.Empty(t, args)
}

type Node struct {
	ID       q.Column[int]
	ParentID q.Column[int]
	Value    q.Column[int]
}

func (Node) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "node",
	}
}

type NodeStatus struct {
	NodeID q.Column[int]
	Status q.Column[string]
}

func (NodeStatus) TableConfig() q.TableConfig {
	return q.TableConfig{
		Name: "node_status",
	}
}

func TestSelectRender_ComplexRecursiveQuery(t *testing.T) {
	NodeTable := Node{}
	NodeStatusTable := NodeStatus{}
	require.NoError(t, q.Bind(&NodeTable))
	require.NoError(t, q.Bind(&NodeStatusTable))

	level := q.Literal(1).As("level")
	base := q.
		Select(NodeTable.ID, NodeTable.ParentID, level).
		Join(NodeStatusTable, NodeTable.ID.Eq(NodeStatusTable.NodeID)).
		Where(NodeStatusTable.Status.Eq("active")).
		CTE("nodes").
		Recursive().
		WithColumns("id", "parent_id", "level")

	rlevel := base.Column("level").Add(1).As("level")

	recursive := q.
		Select(NodeTable.ID, NodeTable.ParentID, rlevel).
		Join(base, NodeTable.ParentID.Eq(base.Column("id")))

	cte := base.UnionAll(recursive.Limit(1)).CTE("nodes")

	query := q.
		Select(cte.Column("id"), cte.Column("parent_id"), cte.Column("level")).
		OrderBy(cte.Column("level"))

	str, args := query.Render(dialect.PostgreSQL{})
	assert.Equal(
		t,
		`WITH RECURSIVE "nodes" AS (`+
			`SELECT "node"."id", "node"."parent_id", 1 AS "level" `+
			`FROM "node" `+
			`JOIN "node_status" ON "node"."id" = "node_status"."node_id" `+
			`WHERE "node_status"."status" = 'active' `+
			`UNION ALL `+
			`(`+
			`SELECT "node"."id", "node"."parent_id", "nodes"."level" + 1 AS "level" `+
			`FROM "node" `+
			`JOIN "nodes" ON "node"."parent_id" = "nodes"."id" `+
			`LIMIT 1`+
			`)`+
			`) `+
			`SELECT "nodes"."id", "nodes"."parent_id", "nodes"."level" FROM "nodes" ORDER BY "nodes"."level"`,
		str,
	)
	assert.Empty(t, args)
}
