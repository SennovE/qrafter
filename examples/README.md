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
ALTER TABLE statement:

```sh
go run ./examples/schema
```
