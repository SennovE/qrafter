package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Database is implemented by *sql.DB, *sql.Tx, and *sql.Conn.
type Database interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Introspector reads a database schema into a normalized Schema snapshot.
type Introspector interface {
	ReadSchema(ctx context.Context, db Database) (Schema, error)
}

func getSchemaDiff(ctx context.Context, config *MigrationToolConfig) (*schemaDiff, error) {
	db, closeDB, err := migrationDatabase(ctx, config)
	if err != nil {
		return nil, err
	}
	defer func() { _ = closeDB() }()

	current, err := readSchema(ctx, db, config.Introspector)
	if err != nil {
		return nil, err
	}

	versionTable := migrationVersionTable(config.VersionTable)
	current = withoutMigrationVersionTable(current, versionTable)
	desired := cloneNormalizedSchema(config.Desired(config.Dialect))
	desired = withoutMigrationVersionTable(desired, versionTable)
	diff := diffSchemas(current, desired)
	return &diff, nil
}

func readSchema(ctx context.Context, db Database, introspector Introspector) (Schema, error) {
	if introspector == nil {
		return Schema{}, fmt.Errorf("migrations: introspector is nil")
	}
	return introspector.ReadSchema(ctx, db)
}

func migrationDatabase(ctx context.Context, config *MigrationToolConfig) (Database, func() error, error) {
	if config.DB != nil {
		return config.DB, noopClose, nil
	}

	db, err := sql.Open(config.DriverName, config.DataSourceName)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, nil, fmt.Errorf("ping database: %w; close database: %v", err, closeErr)
		}
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}
	return db, db.Close, nil
}

func noopClose() error {
	return nil
}

func withoutMigrationVersionTable(schema Schema, versionTable string) Schema {
	if versionTable == "" {
		return schema
	}

	tables := schema.Tables[:0]
	for i := range schema.Tables {
		if strings.EqualFold(schema.Tables[i].Name, versionTable) {
			continue
		}
		tables = append(tables, schema.Tables[i])
	}
	schema.Tables = tables
	return schema
}
