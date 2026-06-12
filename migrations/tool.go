package migrations

import (
	"context"
	"database/sql"
	"fmt"

	q "github.com/SennovE/qrafter"
	"github.com/SennovE/qrafter/dialect"
)

// MigrationToolConfig configures database schema reading and migration code
// generation.
type MigrationToolConfig struct {
	DB Database

	DriverName     string
	DataSourceName string

	Introspector Introspector
	Dialect      dialect.Renderer
	Desired      Schema

	Code MigrationCodeOptions
}

// GeneratedMigration is the full result of reading the database, diffing it
// against the desired schema, and generating migration code.
type GeneratedMigration struct {
	Current Schema
	Desired Schema
	Diff    SchemaDiff

	DDL  Migration
	Code []byte
}

// GenerateMigrationCode reads the current database schema, diffs it against
// the desired schema, and generates ddl-based Go migration functions.
func GenerateMigrationCode(ctx context.Context, config *MigrationToolConfig) (GeneratedMigration, error) {
	if err := validateMigrationToolConfig(config); err != nil {
		return GeneratedMigration{}, err
	}

	db, closeDB, err := migrationDatabase(ctx, config)
	if err != nil {
		return GeneratedMigration{}, err
	}
	defer func() { _ = closeDB() }()

	current, err := ReadSchema(ctx, db, config.Introspector)
	if err != nil {
		return GeneratedMigration{}, err
	}

	desired := cloneNormalizedSchema(config.Desired)
	diff := DiffSchemas(current, desired)
	code, err := GenerateMigrationCodeFromDiff(diff, config.Code)
	if err != nil {
		return GeneratedMigration{}, err
	}

	return GeneratedMigration{
		Current: current,
		Desired: desired,
		Diff:    diff,
		DDL:     diff.DDL(),
		Code:    code,
	}, nil
}

// GenerateMigrationCodeForTableConfig builds the desired schema from a qrafter
// table config and then generates migration code against the database schema.
func GenerateMigrationCodeForTableConfig[T q.TableConfigProvider](
	ctx context.Context,
	config *MigrationToolConfig,
) (GeneratedMigration, error) {
	if config == nil {
		return GeneratedMigration{}, fmt.Errorf("migrations: config is nil")
	}
	if config.Dialect == nil {
		return GeneratedMigration{}, fmt.Errorf("migrations: dialect is nil")
	}
	configCopy := *config
	configCopy.Desired = TableConfigToSchema[T](config.Dialect)
	return GenerateMigrationCode(ctx, &configCopy)
}

func validateMigrationToolConfig(config *MigrationToolConfig) error {
	if config == nil {
		return fmt.Errorf("migrations: config is nil")
	}
	if config.Introspector == nil {
		return fmt.Errorf("migrations: introspector is nil")
	}
	if config.DB != nil {
		return nil
	}
	if config.DriverName == "" {
		return fmt.Errorf("migrations: driver name is required")
	}
	if config.DataSourceName == "" {
		return fmt.Errorf("migrations: data source name is required")
	}
	return nil
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
