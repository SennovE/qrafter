package migrations

import (
	"bytes"
	"context"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/SennovE/qrafter/dialect"
)

const (
	defaultMigrationName      = "migration"
	defaultMigrationFileMode  = 0o644
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
	Desired      Schema
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
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return "", fmt.Errorf("create migration directory: %w", err)
	}

	if err := createRegistryFile(outDir); err != nil {
		return "", fmt.Errorf("create registry file: %w", err)
	}

	path := filepath.Join(outDir, filename)

	// #nosec G304 -- path is generated inside the caller-provided migration output directory.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultMigrationFileMode)
	if err != nil {
		return "", fmt.Errorf("create migration file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(code); err != nil {
		return "", fmt.Errorf("write migration file: %w", err)
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
		comment = defaultMigrationName
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

	src, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format %s template: %w", name, err)
	}
	return src, nil
}

func renderMigrationCodeTemplate(data migrationTemplateData) ([]byte, error) {
	return renderGoTemplate("migration", migrationTemplate, data)
}
