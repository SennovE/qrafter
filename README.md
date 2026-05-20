# qrafter

[![Go Reference](https://pkg.go.dev/badge/github.com/SennovE/qrafter.svg)](https://pkg.go.dev/github.com/SennovE/qrafter)

Qrafter forges dialect-aware SQL queries directly from Go structs.

It is a small query builder focused on typed table definitions, composable SQL
expressions, parameterized rendering, and scanning results back into the same
column fields.

## Install

```sh
go get github.com/SennovE/qrafter
```

## Quick start

```go
package main

import (
	"fmt"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

type User struct {
	q.Table `table:"users"`

	ID       q.Column[int] `db:"id"`
	UserName q.Column[string]
	Age      q.Column[int]
}

func main() {
	users := q.MustNewTable[User]()

	sql, args := q.Select(users.ID, users.UserName).
		Where(users.Age.Ge(18), users.UserName.Eq("Alice")).
		OrderBy(users.ID.Asc()).
		Limit(10).
		Render(dialect.PostgreSQL{})

	fmt.Println(sql)
	fmt.Println(args)
}
```

Output:

```text
SELECT "users"."id", "users"."user_name" FROM "users" WHERE "users"."age" >= $1 AND "users"."user_name" = $2 ORDER BY "users"."id" ASC LIMIT 10
[18 Alice]
```

## Features

- Typed table structs with `qrafter.Column[T]`
- Table configuration via a `TableConfig()` method or embedded `qrafter.Table`
- Automatic column binding from field names or `db` tags
- Dialect-aware identifier quoting and placeholders
- Parameterized `SELECT`, joins, grouping, ordering, limits, and offsets
- Parameterized `INSERT` with `VALUES`, `DEFAULT VALUES`, `INSERT ... SELECT`, and `RETURNING`
- Parameterized `UPDATE` with `SET`, `FROM`, `WHERE`, CTEs, and `RETURNING`
- Parameterized `DELETE` with `WHERE`, `USING`, CTEs, and `RETURNING`
- CTEs, recursive CTEs, compound queries, aggregates, and window functions
- `database/sql` and `sqlx`-friendly scanning helpers
