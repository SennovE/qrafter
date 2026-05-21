package qrafter_test

import (
	"database/sql"
	"fmt"
	"log"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

type exampleUser struct {
	q.Table `table:"users"`

	ID       q.Column[int] `db:"id"`
	UserName q.Column[string]
	Age      q.Column[int]
}

type Node struct {
	q.Table `table:"node"`

	ID       q.Column[int]
	ParentID q.Column[int]
	Value    q.Column[int]
}

type NodeStatus struct {
	q.Table `table:"node_status"`

	NodeID q.Column[int]
	Status q.Column[string]
}

func ExampleSelect() {
	users := q.MustNewTable[exampleUser]()

	sql, args := q.Select(users.ID, users.UserName).
		Where(users.Age.Ge(18), users.UserName.Eq("Alice")).
		OrderBy(users.ID.Asc()).
		Limit(10).
		Render(dialect.PostgreSQL{})

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// SELECT "users"."id", "users"."user_name"
	// FROM "users"
	// WHERE "users"."age" >= $1 AND "users"."user_name" = $2
	// ORDER BY "users"."id" ASC
	// LIMIT 10
	// [18 Alice]
}

func ExampleInsert() {
	users := q.MustNewTable[exampleUser]()

	sql, args := q.Insert(users).
		Columns(users.UserName, users.Age).
		Values("Alice", 18).
		Returning(users.ID).
		Render(dialect.PostgreSQL{})

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// INSERT INTO "users" ("user_name", "age")
	// VALUES ($1, $2)
	// RETURNING "users"."id"
	// [Alice 18]
}

func ExampleUpdate() {
	users := q.MustNewTable[exampleUser]()

	sql, args := q.Update(users).
		Set(users.UserName, "Alice").
		Where(users.ID.Eq(1)).
		Returning(users.ID, users.UserName).
		Render(dialect.PostgreSQL{})

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// UPDATE "users"
	// SET "user_name" = $1
	// WHERE "users"."id" = $2
	// RETURNING "users"."id", "users"."user_name"
	// [Alice 1]
}

func ExampleDelete() {
	users := q.MustNewTable[exampleUser]()

	sql, args := q.Delete(users).
		Where(users.Age.Lt(18)).
		Returning(users.ID).
		Render(dialect.PostgreSQL{})

	fmt.Println(sql)
	fmt.Println(args)

	// Output:
	// DELETE FROM "users"
	// WHERE "users"."age" < $1
	// RETURNING "users"."id"
	// [18]
}

func ExampleTableAlias() {
	users := q.MustNewTable[exampleUser]()
	managers, err := q.TableAlias(users, "manager")
	if err != nil {
		panic(err)
	}

	sql, _ := q.Select(users.UserName, managers.UserName).
		Join(managers, users.Age.Eq(managers.Age)).
		Render(dialect.PostgreSQL{})

	fmt.Println(sql)

	// Output:
	// SELECT "users"."user_name", "manager"."user_name"
	// FROM "users"
	// JOIN "users" AS "manager" ON "users"."age" = "manager"."age"
}

func ExampleScanDest() {
	db, err := sql.Open("postgres", "postgres://user:pass@localhost/app")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	user := q.MustNewTable[exampleUser]()
	query, _ := q.Select(user.ID, user.UserName, user.Age).Render(dialect.PostgreSQL{})

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		dest, err := q.ScanDest(&user)
		if err != nil {
			log.Fatal(err)
		}
		if err := rows.Scan(dest...); err != nil {
			log.Fatal(err)
		}

		fmt.Println(user.ID.Get(), user.UserName.Get(), user.Age.Get())
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func ExampleSelectQuery_CTEs() {
	cte1 := q.Select(q.Literal(1)).CTE("cte1").WithColumns("c1")
	query := q.Select(cte1.Column("c1"))

	sql, _ := query.Render(dialect.PostgreSQL{})
	fmt.Println(sql)
	// Output:
	// WITH "cte1" ("c1") AS (
	//     SELECT 1
	// )
	// SELECT "cte1"."c1"
	// FROM "cte1"
}

func ExampleSelectQuery_CTEs_complex_recursive_query() {
	NodeTable := q.MustNewTable[Node]()
	NodeStatusTable := q.MustNewTable[NodeStatus]()

	level := q.Literal(1).As("level")
	base := q.
		Select(NodeTable.ID, NodeTable.ParentID, level).
		Join(NodeStatusTable, NodeTable.ID.Eq(NodeStatusTable.NodeID)).
		Where(NodeStatusTable.Status.Eq(q.Literal("active"))).
		CTE("nodes").
		Recursive().
		WithColumns("id", "parent_id", "level")

	rlevel := base.Column("level").Add(q.Literal(1)).As("level")

	recursive := q.
		Select(NodeTable.ID, NodeTable.ParentID, rlevel).
		Join(base, NodeTable.ParentID.Eq(base.Column("id")))

	cte := base.UnionAll(recursive.Limit(1)).CTE("nodes")

	query := q.
		Select(cte.Column("id"), cte.Column("parent_id"), cte.Column("level")).
		OrderBy(cte.Column("level"))

	sql, _ := query.Render(dialect.PostgreSQL{})
	fmt.Println(sql)
	// Output:
	// WITH RECURSIVE "nodes" AS (
	//     SELECT "node"."id", "node"."parent_id", 1 AS "level"
	//     FROM "node"
	//     JOIN "node_status" ON "node"."id" = "node_status"."node_id"
	//     WHERE "node_status"."status" = 'active'
	//     UNION ALL
	//     (SELECT "node"."id", "node"."parent_id", "nodes"."level" + 1 AS "level"
	//     FROM "node"
	//     JOIN "nodes" ON "node"."parent_id" = "nodes"."id"
	//     LIMIT 1)
	// )
	// SELECT "nodes"."id", "nodes"."parent_id", "nodes"."level"
	// FROM "nodes"
	// ORDER BY "nodes"."level"
}
