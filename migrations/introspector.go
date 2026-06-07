package migrations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/SennovE/qrafter/ddl"
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

// ReadSchema reads a database schema with the given DBMS introspector.
func ReadSchema(ctx context.Context, db Database, introspector Introspector) (Schema, error) {
	if introspector == nil {
		return Schema{}, fmt.Errorf("migrations: introspector is nil")
	}
	return introspector.ReadSchema(ctx, db)
}

// ReadDDL reads a database schema and converts it into ddl statements.
func ReadDDL(ctx context.Context, db Database, introspector Introspector) (ddl.Statements, error) {
	schema, err := ReadSchema(ctx, db, introspector)
	if err != nil {
		return nil, err
	}
	return schema.DDL(), nil
}
