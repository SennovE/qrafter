package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateConfigCodePostgreSQL(t *testing.T) {
	code, err := generateConfigCode(&configOptions{
		DriverImportPath: "github.com/lib/pq",
		DriverName:       "postgres",
		Dialect:          "postgres",
		DatabaseDSN:      "postgres://localhost/app?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("generate config code: %v", err)
	}

	want := `package migrations

import (
	"github.com/SennovE/qrafter/dialect"
	qmig "github.com/SennovE/qrafter/migrations"
	_ "github.com/lib/pq"
)

// MigrationConfig configures qrafter migration generation for this project.
var MigrationConfig = qmig.MigrationToolConfig{
	DriverName:     "postgres",
	DataSourceName: "postgres://localhost/app?sslmode=disable",
	Introspector:   qmig.NewPostgreSQL(qmig.WithSchemas("public")),
	Dialect:        dialect.PostgreSQL{},
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
	if normalizeGeneratedSourceForTest(string(code)) != normalizeGeneratedSourceForTest(want) {
		t.Fatalf("config code mismatch:\n%s", string(code))
	}
}

func TestGenerateMigrationsConfigWritesFormattedFile(t *testing.T) {
	dir := t.TempDir()

	err := GenerateMigrationsConfig([]string{
		"--dir", dir,
		"--driver-import", "github.com/lib/pq",
		"--driver", "postgres",
		"--dialect", "postgres",
		"--dsn", "postgres://localhost/app?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("generate migrations config: %v", err)
	}

	path := filepath.Join(dir, configFilename)
	// #nosec G304 -- path is the generated config file inside t.TempDir().
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read generated config: %v", err)
	}
	for _, want := range []string{
		`_ "github.com/lib/pq"`,
		`DriverName:     "postgres",`,
		`Introspector:   qmig.NewPostgreSQL(qmig.WithSchemas("public")),`,
		`VersionTable:   qmig.DefaultMigrationVersionTable,`,
	} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("generated config does not contain %q:\n%s", want, string(got))
		}
	}
}

func normalizeGeneratedSourceForTest(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimRight(s, "\n") + "\n"
}
