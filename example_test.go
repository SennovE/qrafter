package qrafter_test

import (
	"database/sql"
	"fmt"
	"log"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

type exampleUser struct {
	ID       q.Column[int] `db:"id"`
	UserName q.Column[string]
	Age      q.Column[int]
}

func (exampleUser) TableConfig() q.TableConfig {
	return q.TableConfig{Name: "users"}
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
	// SELECT "users"."id", "users"."user_name" FROM "users" WHERE "users"."age" >= $1 AND "users"."user_name" = $2 ORDER BY "users"."id" ASC LIMIT 10
	// [18 Alice]
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
	// SELECT "users"."user_name", "manager"."user_name" FROM "users" JOIN "users" AS "manager" ON "users"."age" = "manager"."age"
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
