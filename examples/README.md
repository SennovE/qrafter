# qrafter examples

This directory contains larger examples that show qrafter in application-shaped
code, beyond the short snippets in the package documentation.

## database_sql

`database_sql` shows repository-style code on top of `database/sql`: build a
typed query, render it for a dialect, execute it with context-aware database
methods, and scan the result into application structs.

It intentionally does not import a database driver so the root module stays
dependency-free. In a real application, import the driver you use.

## reporting

`reporting` builds a larger analytical query with joins, grouping, a CTE, and a
window function:

```sh
go run ./examples/reporting
```

## schema

`schema` renders a small schema plan with tables, constraints, indexes, and an
ALTER TABLE statement. The same builders can be rendered with any included
dialect, including PostgreSQL, MySQL, SQLite, Oracle, and SQL Server:

```sh
go run ./examples/schema
```

## migrations

`migrations` is a standalone module with a PostgreSQL Docker Compose file and a
ready qrafter migration config. It shows the full flow: start a database,
generate a migration, apply it, and revert it.

```sh
cd examples/migrations
docker compose up -d
go run github.com/SennovE/qrafter/cmd/qrafter-migrations revision --dir ./migrations --comment create_organizations_and_users
go run github.com/SennovE/qrafter/cmd/qrafter-migrations up --dir ./migrations --to head
```
