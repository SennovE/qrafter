package migrations

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/SennovE/qrafter/dialect"
)

const (
	defaultRevisionTimeLayout = "20060102150405"
)

const migrationTemplate = `package migrations

import qddl "github.com/SennovE/qrafter/ddl"

func Up{{.MigrationNumber}}() qddl.Statements {
{{- if .UpStatements }}
	return qddl.Statements{
{{- range .UpStatements }}
		{{ . }},
{{- end }}
	}
{{- else }}
	return nil
{{- end }}
}

func Down{{.MigrationNumber}}() qddl.Statements {
{{- if .DownStatements }}
	return qddl.Statements{
{{- range .DownStatements }}
		{{ . }},
{{- end }}
	}
{{- else }}
	return nil
{{- end }}
}
`

// MigrationToolConfig configures database schema reading and migration code
// generation.
type MigrationToolConfig struct {
	DB Database

	DriverName     string
	DataSourceName string

	Introspector Introspector
	Dialect      dialect.Renderer
	Desired      func(dialect.Renderer) Schema
	VersionTable string
}

type migrationTemplateData struct {
	MigrationNumber string
	UpStatements    []string
	DownStatements  []string
}

// MakeMigration generates a migration file and returns its path.
func MakeMigration(ctx context.Context, comment, outDir string, config *MigrationToolConfig) (string, error) {
	if err := validateMigrationToolConfig(config); err != nil {
		return "", err
	}

	filename, revisionVersion := migrationFileName(comment)
	diff, err := getSchemaDiff(ctx, config)
	if err != nil {
		return "", err
	}

	code, err := generateMigrationFileText(*diff, revisionVersion)
	if err != nil {
		return "", err
	}

	if outDir == "" {
		outDir = "."
	}

	path := filepath.Join(outDir, filename)
	if err := createFile(path, code); err != nil {
		return "", fmt.Errorf("create migration file: %w", err)
	}

	if err := appendMigrationToRegistry(outDir, revisionVersion); err != nil {
		return "", err
	}
	return path, nil
}

func validateMigrationToolConfig(config *MigrationToolConfig) error {
	if config == nil {
		return fmt.Errorf("migrations: config is nil")
	}
	if config.Introspector == nil {
		return fmt.Errorf("migrations: introspector is nil")
	}
	if config.Desired == nil {
		return fmt.Errorf("migrations: desired schema function is nil")
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

func revision() string {
	return time.Now().UTC().Format(defaultRevisionTimeLayout)
}

func migrationFileName(comment string) (filename, migrationNumber string) {
	revision := revision()
	if comment == "" {
		comment = "migration"
	} else {
		comment = strings.ReplaceAll(comment, " ", "_")
	}
	return revision + "_" + comment + ".go", revision
}

func renderGoTemplate(name, source string, data any) ([]byte, error) {
	t, err := template.New(name).Option("missingkey=error").Parse(source)
	if err != nil {
		return nil, fmt.Errorf("parse %s template: %w", name, err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("execute %s template: %w", name, err)
	}
	return b.Bytes(), nil
}

func generateMigrationFileText(diff schemaDiff, migrationNumber string) ([]byte, error) {
	steps := migrationSteps(diff)
	return renderGoTemplate("migration", migrationTemplate, migrationTemplateData{
		MigrationNumber: migrationNumber,
		UpStatements:    upStepCodes(steps),
		DownStatements:  downStepCodes(steps),
	})
}
