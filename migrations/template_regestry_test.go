package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendMigrationToRegistryAddsFirstEntry(t *testing.T) {
	dir := writeRegistryConfig(t, `package migrations

import qmig "github.com/SennovE/qrafter/migrations"

var Registry = []qmig.Migration{
}
`)

	if err := appendMigrationToRegistry(dir, "20260620120000"); err != nil {
		t.Fatalf("append migration to registry: %v", err)
	}

	got := readRegistryConfig(t, dir)
	want := `var Registry = []qmig.Migration{
	qmig.New("20260620120000", Up20260620120000, Down20260620120000),
}`
	if !strings.Contains(got, want) {
		t.Fatalf("registry does not contain first entry:\n%s", got)
	}
}

func TestAppendMigrationToRegistryKeepsExistingEntries(t *testing.T) {
	dir := writeRegistryConfig(t, `package migrations

import qmig "github.com/SennovE/qrafter/migrations"

var Registry = []qmig.Migration{
	qmig.New("20260620110000", Up20260620110000, Down20260620110000),
}
`)

	if err := appendMigrationToRegistry(dir, "20260620120000"); err != nil {
		t.Fatalf("append migration to registry: %v", err)
	}

	got := readRegistryConfig(t, dir)
	for _, want := range []string{
		`qmig.New("20260620110000", Up20260620110000, Down20260620110000),`,
		`qmig.New("20260620120000", Up20260620120000, Down20260620120000),`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("registry does not contain %q:\n%s", want, got)
		}
	}
	if !strings.Contains(got, "var Registry = []qmig.Migration{\n") {
		t.Fatalf("registry was not formatted:\n%s", got)
	}
}

func TestAppendMigrationToRegistryReportsMissingRegistry(t *testing.T) {
	dir := writeRegistryConfig(t, `package migrations
`)

	if err := appendMigrationToRegistry(dir, "20260620120000"); err == nil {
		t.Fatal("expected error for missing Registry declaration")
	}
}

func writeRegistryConfig(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, configFilename)
	if err := os.WriteFile(path, []byte(content), defaultFileMode); err != nil {
		t.Fatalf("write registry config: %v", err)
	}
	return dir
}

func readRegistryConfig(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, configFilename)
	// #nosec G304 -- path is the generated registry file inside t.TempDir().
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read registry config: %v", err)
	}
	return string(got)
}
