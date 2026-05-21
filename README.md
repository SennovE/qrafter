# qrafter

[![Go Reference](https://pkg.go.dev/badge/github.com/SennovE/qrafter.svg)](https://pkg.go.dev/github.com/SennovE/qrafter)
[![Go CI](https://github.com/SennovE/qrafter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/SennovE/qrafter/actions/workflows/go.yml)

qrafter is a small type-safe SQL query builder for Go.

It is a query builder for people who want to keep writing SQL-shaped Go
code. Table structs define the available columns, queries compose from those
typed columns, and rendering produces SQL plus driver arguments
for `database/sql`, `sqlx`, and similar packages.

## Why qrafter?

- Define tables once as Go structs with typed `qrafter.Column[T]` fields
- Compose `SELECT`, `INSERT`, `UPDATE`, and `DELETE` statements fluently
- Render parameterized SQL instead of interpolating user values
- Keep using your existing database driver, connection pool, and scanning flow
- Scan results back into the same column fields when that fits your code

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
SELECT "users"."id", "users"."user_name"
FROM "users"
WHERE "users"."age" >= $1 AND "users"."user_name" = $2
ORDER BY "users"."id" ASC
LIMIT 10
[18 Alice]
```

## When to use it

Use qrafter when you want typed query composition in Go, but still want to see
and control the SQL being generated.

It is probably not the right fit if you want a full ORM, schema migrations,
model lifecycle hooks, relationship loading, or generated code from an existing
database schema.

## Features

- Typed table structs with `qrafter.Column[T]`
- Table configuration via a `TableConfig()` method or embedded `qrafter.Table`
- Automatic column binding from field names or `db` tags
- Dialect-aware identifier quoting and placeholders
- Human-readable multiline SQL rendering
- Parameterized `SELECT`, joins, grouping, ordering, limits, and offsets
- Parameterized `INSERT` with `VALUES`, `DEFAULT VALUES`, `INSERT ... SELECT`, and `RETURNING`
- Parameterized `UPDATE` with `SET`, `FROM`, `WHERE`, CTEs, and `RETURNING`
- Parameterized `DELETE` with `WHERE`, `USING`, CTEs, and `RETURNING`
- CTEs, recursive CTEs, compound queries, aggregates, and window functions
- `database/sql` and `sqlx`-friendly scanning helpers

## Dialects

Qrafter currently includes:

- `dialect.BaseDialect` for ANSI-style double-quoted identifiers and `?`
  placeholders
- `dialect.PostgreSQL` for PostgreSQL-style `$1`, `$2`, ... placeholders

New dialects can be added by implementing `dialect.Renderer`.

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the local
development workflow and pull request guidelines.

Good first areas to explore:

- Add examples for common query patterns
- Improve dialect coverage
- Expand integration tests
- Polish package documentation on pkg.go.dev
