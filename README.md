# qrafter

[![Go Reference](https://pkg.go.dev/badge/github.com/SennovE/qrafter.svg)](https://pkg.go.dev/github.com/SennovE/qrafter)
[![Go CI](https://github.com/SennovE/qrafter/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/SennovE/qrafter/actions/workflows/go.yml)
[![Coverage Status](https://coveralls.io/repos/github/SennovE/qrafter/badge.svg?branch=main)](https://coveralls.io/github/SennovE/qrafter?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/SennovE/qrafter)](https://goreportcard.com/report/github.com/SennovE/qrafter)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go)

qrafter is a fluent, type-safe SQL toolkit for Go. It gives you typed query
composition, DDL builders, schema introspection, and generated Go migrations
without becoming an ORM.

Use qrafter when you want explicit SQL, typed table/column references,
database/sql compatibility, and migration files you can read and edit.

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

## Migrations

The migration tool lives in `cmd/qrafter-migrations`. It creates a Go config
file, compares the configured schema with the live database, generates Go
migration files, registers them, and applies or reverts them.

Show CLI help:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations@latest help
go run github.com/SennovE/qrafter/cmd/qrafter-migrations@latest help revision
```

Create `./migrations/qrafter_config.go`:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations@latest init \
  --dir ./migrations \
  --driver-import github.com/lib/pq \
  --driver postgres \
  --dialect postgres \
  --dsn "postgres://app_user:app_password@localhost:5432/app_db?sslmode=disable"
```

The generated config is regular Go code. Add your table configs to
`desiredSchema`:

```go
package migrations

import (
	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	_ "github.com/lib/pq"
)

var MigrationConfig = qmig.MigrationToolConfig{
	DriverName:     "postgres",
	DataSourceName: "postgres://app_user:app_password@localhost:5432/app_db?sslmode=disable",
	Introspector:   qmig.NewPostgreSQL(qmig.WithSchemas("public")),
	Dialect:        dialect.PostgreSQL{},
	Desired:        desiredSchema,
	VersionTable:   qmig.DefaultMigrationVersionTable,
}

var Registry = []qmig.Migration{
}

func desiredSchema(d dialect.Renderer) qmig.Schema {
	var schema qmig.Schema
	qmig.RegisterTable[User](&schema, d) // Add your tables
	return schema
}
```

Generate a migration from the live database diff:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations@latest revision \
  --dir ./migrations \
  --comment create_users
```

This creates a timestamped Go file and appends it to `Registry` in
`qrafter_config.go`. Generated migrations return `ddl.Statements`, so you can
edit them and add custom statements such as `qddl.RawSQL("CREATE EXTENSION ...")`
when needed.

Apply or revert registered migrations:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations@latest up --dir ./migrations --to head
go run github.com/SennovE/qrafter/cmd/qrafter-migrations@latest down --dir ./migrations --to base
```

The apply command stores the current version in
`qrafter_schema_version` by default. Override it in config with
`VersionTable` or from CLI with `--version-table`.

You can test this yourself with [examples/migrations](examples/migrations).

## DDL Builders

Schema statements live in the `ddl` package:

```go
sql, err := ddl.CreateTable("users").
	Columns(
		ddl.Column("id", ddl.BigSerial()).PrimaryKey(),
		ddl.Column("email", ddl.VarChar(320)).NotNull().Unique(),
		ddl.Column("created_at", ddl.TimestampTZ()).DefaultExpr("now()"),
	).
	Render(dialect.PostgreSQL{})
```

DDL rendering is dialect-aware and returns an error when a dialect cannot safely
render a requested feature.

## Dialects

qrafter currently includes:


dialect     | DML | DDL | migrations
----------- |:---:|:---:|:----------:
BaseDialect | yes | yes | n/a
PostgreSQL  | yes | yes | yes
MySQL       | yes | yes | no
SQLite      | yes | yes | no
Oracle      | yes | yes | no
SQLServer   | yes | yes | no

## Examples

More application-shaped examples live in [examples](examples):

- [database_sql](examples/database_sql) shows repository-style code with
  `database/sql`.
- [reporting](examples/reporting) builds a larger analytical query with joins,
  grouping, a CTE, and a window function.
- [schema](examples/schema) renders DDL for tables, constraints, indexes, and
  table alterations.
- [migrations](examples/migrations) is a standalone module with Docker Compose
  that generates, applies, and reverts qrafter migrations against PostgreSQL.

## Project Status

qrafter is pre-v1. The API may still change while the package evolves.

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for the local
development workflow and pull request guidelines.
