# qrafter

[![Go Reference](https://pkg.go.dev/badge/github.com/SennovE/qrafter.svg)](https://pkg.go.dev/github.com/SennovE/qrafter)
[![Go CI](https://github.com/SennovE/qrafter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/SennovE/qrafter/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/SennovE/qrafter)](https://goreportcard.com/report/github.com/SennovE/qrafter)

**qrafter is a fluent, type-safe SQL query builder for Go — no ORM, no codegen, just typed SQL-shaped Go.**

qrafter helps you build parameterized SQL from typed Go table structs.
You define tables once, compose queries from typed columns, and render SQL plus
driver arguments for `database/sql`, `sqlx`, and similar packages.

It is designed for Go developers who want a Go-style way to build explicit SQL: keep queries readable and under control, while avoiding fragile hand-written column names, placeholders, and query fragments.

## Why qrafter?

Use qrafter when you want:

- Typed table and column references with `qrafter.Column[T]`
- SQL that still looks and feels like SQL
- Parameterized queries instead of interpolated user values
- Dialect-aware identifier quoting and placeholders
- Compatibility with your existing database driver and connection pool
- A lightweight query builder instead of a full ORM
- No code generation step in your build workflow

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

	sql, args, err := q.Select(users.ID, users.UserName).
		Where(
			users.Age.Ge(18),
			users.UserName.Eq("Alice"),
		).
		OrderBy(users.ID.Asc()).
		Limit(10).
		Render(dialect.PostgreSQL{})
	if err != nil {
		panic(err)
	}

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

## Larger examples

More application-shaped examples live in [examples](examples):

* [database_sql](examples/database_sql) shows repository-style code with
  `database/sql`, context-aware execution, and typed query rendering.
* [reporting](examples/reporting) builds a larger analytical query with joins,
  grouping, a CTE, and a window function.
* [schema](examples/schema) renders DDL for tables, constraints, indexes, and
  table alterations.

## How it works

A qrafter table is a Go struct with typed column fields:

```go
type User struct {
	q.Table `table:"users"`

	ID       q.Column[int] `db:"id"`
	UserName q.Column[string]
	Age      q.Column[int]
}
```

`q.MustNewTable[User]()` binds the struct fields to SQL table and column names.
Queries are then composed from those typed columns and rendered for a selected
SQL dialect.

Field names are converted into column names automatically, or you can override
them with `db` tags.

## When to use it

qrafter is useful when you want typed query composition while still keeping
control over the generated SQL.

Good fits:

* services that already use `database/sql` or `sqlx`
* projects that prefer explicit SQL over ORM abstractions
* codebases where query fragments need to be composed safely
* applications that want typed table and column references without codegen

Less ideal fits:

* projects that want a full ORM
* applications that expect automatic relationship loading
* teams that prefer writing raw SQL files and generating Go code from them
* projects that need schema migrations as part of the same tool

## Features

* Typed table structs with `qrafter.Column[T]`
* Table configuration via embedded `qrafter.Table` or `TableConfig()`
* Automatic column binding from field names or `db` tags
* Custom field-to-column mapping through `qrafter.NameMapper`
* Dialect-aware identifier quoting and placeholders
* Human-readable multiline SQL rendering
* Parameterized `SELECT`, joins, grouping, ordering, limits, and offsets
* Parameterized `INSERT` with `VALUES`, `DEFAULT VALUES`, `INSERT ... SELECT`, and `RETURNING`
* Parameterized `UPDATE` with `SET`, `FROM`, `WHERE`, CTEs, and `RETURNING`
* Parameterized `DELETE` with `WHERE`, `USING`, CTEs, and `RETURNING`
* CTEs and recursive CTEs
* Compound queries such as `UNION` and `UNION ALL`
* Aggregates and window functions
* DDL builders for tables, columns, constraints, and indexes
* Centralized SQL compiler with dialect override hooks for database-specific syntax
* `database/sql` and `sqlx`-friendly scanning helpers

## DDL

Schema statements live in the separate `ddl` package so the root package can
stay focused on query building:

```go
sql, err := ddl.CreateTable("users").
	Columns(
		ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
		ddl.Column("email", ddl.VarChar(320)).NotNull().Unique(),
		ddl.Column("created_at", ddl.TimestampTZ()).DefaultExpr("now()"),
	).
	Render(dialect.PostgreSQL{})
```

DDL rendering returns an error when a dialect cannot safely render a requested
feature, such as SQLite column type changes or MySQL partial indexes.

Constraints and indexes are built explicitly from table and column names:

```go
sql, err := ddl.Statements{
	ddl.CreateTable("users").
		Columns(
			ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
			ddl.Column("org_id", ddl.BigInt()).NotNull(),
			ddl.Column("email", ddl.VarChar(320)).NotNull(),
		).
		Constraints(
			ddl.Unique("email").Named("users_email_key"),
			ddl.ForeignKey("org_id").
				References("orgs", "id").
				OnDelete(ddl.Cascade),
		),
	ddl.CreateIndex("users_email_idx").
		IfNotExists().
		OnCols("users", "email"),
}.Render(dialect.PostgreSQL{})
```

## Dialects

qrafter currently includes:

* `dialect.BaseDialect` for ANSI-style double-quoted identifiers and `?` placeholders
* `dialect.PostgreSQL` for PostgreSQL-style `$1`, `$2`, ... placeholders
* `dialect.MySQL` for backtick-quoted identifiers, MySQL `LIMIT`/`OFFSET`,
  empty-row inserts, multi-table `UPDATE`/`DELETE`, and NULL ordering emulation
* `dialect.SQLite` for SQLite literals, `LIMIT`/`OFFSET`, and fail-fast
  handling for unsupported `DELETE USING`
* `dialect.Oracle` for Oracle placeholders, boolean literals,
  `OFFSET`/`FETCH`, and Oracle-specific DDL overrides
* `dialect.SQLServer` for bracket-quoted identifiers, `@p1` placeholders,
  `OFFSET`/`FETCH`, NULL ordering emulation, and SQL Server DDL overrides

Rendering is intentionally centralized. Query and DDL builders store statement
state; the compiler renders statements, expressions, clauses, and DDL nodes; a
dialect supplies primitive rules such as identifier quoting, literals,
placeholders, and `LIMIT`/`OFFSET`, and can override specific compiler nodes
with `CompileNode`.

New dialects can start with the primitive methods and then override focused
nodes for features such as `RETURNING`, `UPDATE` sources, `DELETE` sources,
joins, default inserts, NULL ordering, partial indexes, and dialect-specific
`ALTER TABLE` forms.

New dialects can be added by implementing `dialect.Renderer`.

## Comparison

| Approach            | Good when                                             | Tradeoff                                                            |
| ------------------- | ----------------------------------------------------- | ------------------------------------------------------------------- |
| Raw `database/sql`  | You want full control over every query                | SQL strings, placeholders, and column names are maintained manually |
| SQL code generation | You want generated Go code from SQL files             | Adds a generation step and a SQL-first workflow                     |
| ORM                 | You want high-level model and relationship management | SQL can become less explicit and harder to control                  |
| qrafter             | You want typed SQL-shaped Go without ORM or codegen   | It is a lightweight query builder, not a full database framework    |

## Project status

qrafter is pre-v1. The API may still change while the package evolves.

Feedback is especially welcome around:

* API naming
* query composition ergonomics
* dialect support
* real-world usage with `database/sql` and `sqlx`

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the local
development workflow and pull request guidelines.

Good first areas to explore:

* Add examples for common query patterns
* Improve dialect coverage
* Expand integration tests
* Polish package documentation on pkg.go.dev
