# qrafter migrations example

This is a small standalone module that shows the full migration loop:

1. start PostgreSQL in Docker;
2. describe the desired schema in Go;
3. generate a migration from the live database diff;
4. apply it;
5. revert it.

The example uses a local `replace` in `go.mod`, so commands run against the
checkout you are editing.

## Start PostgreSQL

```sh
docker compose up -d
```

The database listens on `localhost:55432`:

```text
postgres://qrafter:qrafter@localhost:55432/qrafter_demo?sslmode=disable
```

## Generate a Migration

The desired schema lives in [`migrations/tables.go`](migrations/tables.go), and
the qrafter config lives in
[`migrations/qrafter_config.go`](migrations/qrafter_config.go).

Run the migration tool from this directory:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations revision \
  --dir ./migrations \
  --comment create_organizations_and_users
```

This creates a timestamped Go migration file in `./migrations` and appends it to
`Registry` in `qrafter_config.go`.

## Apply and Revert

Apply all registered migrations:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations up \
  --dir ./migrations \
  --to head
```

Revert everything back to base:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations down \
  --dir ./migrations \
  --to base
```

qrafter stores the current version in `qrafter_schema_version`.

## Try Another Change

Edit `migrations/tables.go`, for example add a column to `User`, then run:

```sh
go run github.com/SennovE/qrafter/cmd/qrafter-migrations revision \
  --dir ./migrations \
  --comment add_user_column

go run github.com/SennovE/qrafter/cmd/qrafter-migrations up \
  --dir ./migrations \
  --to head
```

Generated migrations are regular Go files. You can edit them and add custom SQL
with `qddl.RawSQL(...)`.

## Reset

```sh
docker compose down -v
```
