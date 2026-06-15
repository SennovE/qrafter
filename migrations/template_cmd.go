package migrations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const revisionCommandTemplate = `package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	_ "{{.DriverImportPath}}"
)

var (
	dsn string = "{{.DatabaseDSN}}"
	outDir string = "{{.MigrationDirectory}}"
	renderer dialect.Renderer = {{.DialectRenderer}}
	databaseIntrospector qmig.Introspector = {{.DatabaseIntrospector}}
	driverName string = "{{.DriverName}}"
)

func main() {
	comment := flag.String("comment", "migration", "migration comment")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	path, err := qmig.MakeMigration(ctx, *comment, outDir, &qmig.MigrationToolConfig{
		DriverName:     driverName,
		DataSourceName: dsn,
		Introspector:   databaseIntrospector,
		Dialect:        renderer,
		Desired:        desiredSchema(renderer),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("created %s\n", path)
}

func desiredSchema(d dialect.Renderer) qmig.Schema {
	var schema qmig.Schema
	// qmig.RegisterTable[YourTable](&schema, d)
	return schema
}

`

// RevisionCommandOptions configures generated revision command source code.
type RevisionCommandOptions struct {
	DriverImportPath   string
	DriverName         string
	Dialect            string
	MigrationDirectory string
	DatabaseDSN        string
}

type revisionCommandTemplateData struct {
	DriverImportPath     string
	DriverName           string
	MigrationDirectory   string
	DatabaseDSN          string
	DialectRenderer      string
	DatabaseIntrospector string
}

// WriteRevisionCommandFile generates a revision command file at path.
func WriteRevisionCommandFile(path string, options *RevisionCommandOptions) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("migrations: revision command path is required")
	}
	code, err := generateRevisionCommandCode(options)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create revision command directory: %w", err)
	}
	// #nosec G304 -- path is a caller-provided destination for generated Go source.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, defaultMigrationFileMode)
	if err != nil {
		return fmt.Errorf("create revision command file: %w", err)
	}
	defer func() { _ = file.Close() }()
	if _, err := file.Write(code); err != nil {
		return fmt.Errorf("write revision command file: %w", err)
	}
	return nil
}

func generateRevisionCommandCode(options *RevisionCommandOptions) ([]byte, error) {
	dialectRenderer, databaseIntrospector := dialectOptions(options.Dialect)
	data := revisionCommandTemplateData{
		DriverImportPath:     options.DriverImportPath,
		DriverName:           options.DriverName,
		MigrationDirectory:   options.MigrationDirectory,
		DatabaseDSN:          options.DatabaseDSN,
		DialectRenderer:      dialectRenderer,
		DatabaseIntrospector: databaseIntrospector,
	}
	return renderGoTemplate("revision command", revisionCommandTemplate, data)
}

func dialectOptions(d string) (renderer, introspector string) {
	d = strings.ToLower(d)
	switch d {
	case "postgres", "pg", "postgresql", "pgx":
		return "dialect.PostgreSQL{}", `qmig.NewPostgreSQL(qmig.WithSchemas("public"))`
	}
	return "nil /* TODO: set dialect renderer manually */", "nil /* TODO: set database introspector manually */"
}
