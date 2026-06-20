package migrations

import (
	"fmt"
	"path/filepath"
	"strings"
)

const configFilename = "qrafter_config.go"

const (
	dialectNamePostgres   = "postgres"
	dialectNamePG         = "pg"
	dialectNamePostgreSQL = "postgresql"
	dialectNamePGX        = "pgx"
)

const configTemplate = `package migrations

import (
	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	_ "{{.DriverImportPath}}"
)

// MigrationConfig configures qrafter migration generation for this project.
var MigrationConfig = qmig.MigrationToolConfig{
	DriverName:     "{{.DriverName}}",
	DataSourceName: "{{.DatabaseDSN}}",
	Introspector:   {{.DatabaseIntrospector}},
	Dialect:        {{.DialectRenderer}},
	Desired:        desiredSchema,
	VersionTable:   qmig.DefaultMigrationVersionTable,
}

// Registry stores generated migrations in version order.
var Registry = []qmig.Migration{
}

func desiredSchema(d dialect.Renderer) qmig.Schema {
	var schema qmig.Schema
	// qmig.RegisterTable[YourTable](&schema, d)
	return schema
}

`

type configOptions struct {
	DriverImportPath string
	DriverName       string
	Dialect          string
	DatabaseDSN      string
}

type configTemplateData struct {
	DriverImportPath     string
	DriverName           string
	DatabaseDSN          string
	DialectRenderer      string
	DatabaseIntrospector string
}

// GenerateMigrationsConfig creates the user-editable qrafter migration config.
func GenerateMigrationsConfig(args []string) error {
	return generateMigrationsConfig(args)
}

func generateMigrationsConfig(args []string) error {
	path, options, err := configOptionsFromArgs(args)
	if err != nil {
		return err
	}
	return writeConfigFile(path, options)
}

func writeConfigFile(path string, options *configOptions) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("migrations directory is required")
	}
	code, err := generateConfigCode(options)
	if err != nil {
		return err
	}
	filePath := filepath.Join(path, configFilename)
	if err := createFile(filePath, code); err != nil {
		return fmt.Errorf("create migrations config file: %w", err)
	}
	return nil
}

func generateConfigCode(options *configOptions) ([]byte, error) {
	dialectRenderer, databaseIntrospector := dialectOptions(options.Dialect)
	data := configTemplateData{
		DriverImportPath:     options.DriverImportPath,
		DriverName:           options.DriverName,
		DatabaseDSN:          options.DatabaseDSN,
		DialectRenderer:      dialectRenderer,
		DatabaseIntrospector: databaseIntrospector,
	}
	return renderGoTemplate("migrations config", configTemplate, data)
}

func dialectOptions(d string) (renderer, introspector string) {
	d = strings.ToLower(d)
	switch d {
	case dialectNamePostgres, dialectNamePG, dialectNamePostgreSQL, dialectNamePGX:
		return "dialect.PostgreSQL{}", `qmig.NewPostgreSQL(qmig.WithSchemas("public"))`
	}
	return "nil /* TODO: set dialect renderer manually */", "nil /* TODO: set database introspector manually */"
}
