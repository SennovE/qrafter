package migrations

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateApplyRunnerCodeImportsConfigPackage(t *testing.T) {
	code, err := generateApplyRunnerCode(&resolvedApplyOptions{
		ConfigImportPath: "example.com/app/migrations",
		Target:           "head",
		VersionTable:     "schema_version",
		Timeout:          15 * time.Second,
		Direction:        DirectionUp,
	})
	if err != nil {
		t.Fatalf("generate apply runner code: %v", err)
	}

	wantSnippets := []string{
		`configpkg "example.com/app/migrations"`,
		`versionTable := config.VersionTable`,
		`if "schema_version" != "" {`,
		`Registry:       configpkg.Registry,`,
		`Direction:      qmig.DirectionUp,`,
		`Target:         "head",`,
		`VersionTable:    versionTable,`,
	}
	for _, snippet := range wantSnippets {
		if !strings.Contains(string(code), snippet) {
			t.Fatalf("generated runner does not contain %q:\n%s", snippet, string(code))
		}
	}
}

func TestGenerateApplyRunnerCodeDownDirection(t *testing.T) {
	code, err := generateApplyRunnerCode(&resolvedApplyOptions{
		ConfigImportPath: "example.com/app/migrations",
		Target:           "base",
		Timeout:          time.Second,
		Direction:        DirectionDown,
	})
	if err != nil {
		t.Fatalf("generate apply runner code: %v", err)
	}

	for _, snippet := range []string{
		`Direction:      qmig.DirectionDown,`,
		`Target:         "base",`,
		`"reverted"`,
		`if "" != "" {`,
	} {
		if !strings.Contains(string(code), snippet) {
			t.Fatalf("generated down runner does not contain %q:\n%s", snippet, string(code))
		}
	}

	if got := directionCode("custom"); got != `qmig.MigrationDirection("custom")` {
		t.Fatalf("custom direction code = %q", got)
	}
}
