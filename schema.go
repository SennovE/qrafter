package qrafter

import "github.com/SennovE/qrafter/ddl"

// ColumnConfig describes explicit DDL overrides for a table column.
//
// Values from ColumnConfig usually override information inferred from the Go
// field type and struct tags.
type ColumnConfig struct {
	// Type is the database column type.
	//
	// If zero, the type is inferred from the Go column type.
	Type ddl.Type

	// NotNull marks the column as NOT NULL.
	NotNull bool

	// Unique marks the column as having a single-column UNIQUE constraint.
	Unique bool

	// PrimaryKey marks the column as part of the table primary key.
	PrimaryKey bool

	// Default is the column DEFAULT expression.
	//
	// If Empty, no explicit default is applied.
	Default ddl.Expression
}

// ColumnsConfig describes explicit DDL overrides by column.
type ColumnsConfig map[DDLColumn]ColumnConfig

// IndexesConfig describes additional table indexes.
type IndexesConfig []ddl.CreateIndexStmt

// ConstraintsConfig describes additional table-level constraints.
type ConstraintsConfig []ddl.TableConstraint

// TableConfig describes explicit DDL configuration for a qrafter table.
//
// It complements metadata inferred from the table struct fields and tags.
type TableConfig struct {
	// Schema is the database schema name, for example "public".
	Schema string

	// Name is the database table name.
	Name string

	// Columns contains per-column DDL overrides.
	Columns ColumnsConfig

	// Constraints contains additional table-level constraints.
	Constraints []ddl.TableConstraint

	// Indexes contains additional table indexes.
	Indexes []ddl.CreateIndexStmt
}
