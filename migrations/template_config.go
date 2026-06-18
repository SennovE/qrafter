package migrations

import (
	"fmt"
	"path/filepath"
	"strings"
)


const configFilename = "qrafter_config.go"

const configTemplate = `package migrations

import (
	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	_ "{{.DriverImportPath}}"
)

var MigrationConfig = qmig.MigrationToolConfig{
	DriverName:     "{{.DriverName}}",
	DataSourceName: "{{.DatabaseDSN}}",
	Introspector:   {{.DatabaseIntrospector}},
	Dialect:        {{.DialectRenderer}},
	Desired:        desiredSchema,
}

var Registry = []qmig.Migration{
}

func desiredSchema(d dialect.Renderer) qmig.Schema {
	var schema qmig.Schema
	// qmig.RegisterTable[YourTable](&schema, d)
	return schema
}

`

type configOptions struct {
	DriverImportPath   string
	DriverName         string
	Dialect            string
	DatabaseDSN        string
}

type configTemplateData struct {
	DriverImportPath     string
	DriverName           string
	DatabaseDSN          string
	DialectRenderer      string
	DatabaseIntrospector string
}

func GenerateMigrationsConfig(args []string) error {
	path, options, err := revisionCommandOptionsFromArgs(args)
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
		return fmt.Errorf("create revision command file: %w", err)
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
	return renderGoTemplate("revision command", configTemplate, data)
}

func dialectOptions(d string) (renderer, introspector string) {
	d = strings.ToLower(d)
	switch d {
	case "postgres", "pg", "postgresql", "pgx":
		return "dialect.PostgreSQL{}", `qmig.NewPostgreSQL(qmig.WithSchemas("public"))`
	}
	return "nil /* TODO: set dialect renderer manually */", "nil /* TODO: set database introspector manually */"
}
