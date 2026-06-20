package migrations

import (
	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	_ "github.com/lib/pq"
)

// MigrationConfig configures qrafter migration generation for this example.
var MigrationConfig = qmig.MigrationToolConfig{
	DriverName:     "postgres",
	DataSourceName: "postgres://qrafter:qrafter@localhost:55432/qrafter_demo?sslmode=disable",
	Introspector:   qmig.NewPostgreSQL(qmig.WithSchemas("public")),
	Dialect:        dialect.PostgreSQL{},
	Desired:        desiredSchema,
	VersionTable:   qmig.DefaultMigrationVersionTable,
}

// Registry stores generated migrations in version order.
var Registry = []qmig.Migration{}

func desiredSchema(d dialect.Renderer) qmig.Schema {
	var schema qmig.Schema
	qmig.RegisterTable[Organization](&schema, d)
	qmig.RegisterTable[User](&schema, d)
	return schema
}
