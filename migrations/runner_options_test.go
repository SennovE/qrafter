package migrations

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRevisionOptionsFromArgs(t *testing.T) {
	options, err := revisionOptionsFromArgs([]string{
		"--workdir", "app",
		"--dir", "db/migrations",
		"--config-import", "example.com/app/db/migrations",
		"--comment", "add_users",
		"--go", "gotip",
		"--timeout", "5s",
	})
	if err != nil {
		t.Fatalf("revisionOptionsFromArgs error = %v", err)
	}
	if options.WorkDir != "app" ||
		options.MigrationDir != "db/migrations" ||
		options.ConfigImportPath != "example.com/app/db/migrations" ||
		options.Comment != "add_users" ||
		options.GoBinary != "gotip" ||
		options.Timeout != 5*time.Second {
		t.Fatalf("options = %#v", options)
	}

	for _, args := range [][]string{
		{"--workdir", ""},
		{"--dir", ""},
		{"--go", ""},
		{"--timeout", "0s"},
	} {
		if _, err := revisionOptionsFromArgs(args); err == nil {
			t.Fatalf("revisionOptionsFromArgs(%v) error = nil", args)
		}
	}
}

func TestApplyOptionsFromArgs(t *testing.T) {
	options, err := applyOptionsFromArgs(DirectionDown, []string{
		"--workdir", "app",
		"--dir", "db/migrations",
		"--config-import", "example.com/app/db/migrations",
		"--to", "base",
		"--version-table", "schema_version",
		"--go", "gotip",
		"--timeout", "7s",
	})
	if err != nil {
		t.Fatalf("applyOptionsFromArgs error = %v", err)
	}
	if options.Direction != DirectionDown ||
		options.WorkDir != "app" ||
		options.MigrationDir != "db/migrations" ||
		options.Target != "base" ||
		options.VersionTable != "schema_version" ||
		options.GoBinary != "gotip" ||
		options.Timeout != 7*time.Second {
		t.Fatalf("options = %#v", options)
	}

	for _, args := range [][]string{
		{"--workdir", ""},
		{"--dir", ""},
		{"--go", ""},
		{"--timeout", "0s"},
	} {
		if _, err := applyOptionsFromArgs(DirectionUp, args); err == nil {
			t.Fatalf("applyOptionsFromArgs(%v) error = nil", args)
		}
	}
	if _, err := applyOptionsFromArgs("sideways", nil); err == nil {
		t.Fatal("unsupported direction error = nil")
	}
}

func TestResolveRunnerOptionsWithExplicitImportPath(t *testing.T) {
	ctx := context.Background()

	revision, err := resolveRevisionOptions(ctx, &revisionOptions{
		WorkDir:          ".",
		MigrationDir:     "migrations",
		ConfigImportPath: " example.com/app/migrations ",
		Comment:          "test",
		GoBinary:         "go",
		Timeout:          time.Second,
	})
	if err != nil {
		t.Fatalf("resolve revision options: %v", err)
	}
	if revision.ConfigImportPath != "example.com/app/migrations" || !filepath.IsAbs(revision.WorkDir) {
		t.Fatalf("revision resolved = %#v", revision)
	}

	apply, err := resolveApplyOptions(ctx, &applyOptions{
		WorkDir:          ".",
		MigrationDir:     "migrations",
		ConfigImportPath: " example.com/app/migrations ",
		GoBinary:         "go",
		Timeout:          time.Second,
		Direction:        DirectionUp,
	})
	if err != nil {
		t.Fatalf("resolve apply options: %v", err)
	}
	if apply.ConfigImportPath != "example.com/app/migrations" || !filepath.IsAbs(apply.WorkDir) {
		t.Fatalf("apply resolved = %#v", apply)
	}

	if _, err := resolveRevisionOptions(ctx, nil); err == nil {
		t.Fatal("nil revision options error = nil")
	}
	if _, err := resolveApplyOptions(ctx, nil); err == nil {
		t.Fatal("nil apply options error = nil")
	}
}

func TestDeriveConfigImportPathFromModule(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "db", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o750); err != nil {
		t.Fatalf("mkdir migrations: %v", err)
	}

	module := goModuleInfo{Path: "example.com/app", Dir: root}
	got, err := deriveConfigImportPathFromModule(module, root, filepath.Join("db", "migrations"))
	if err != nil {
		t.Fatalf("derive import path: %v", err)
	}
	if got != "example.com/app/db/migrations" {
		t.Fatalf("import path = %q", got)
	}

	got, err = deriveConfigImportPathFromModule(module, root, root)
	if err != nil {
		t.Fatalf("derive root import path: %v", err)
	}
	if got != "example.com/app" {
		t.Fatalf("root import path = %q", got)
	}

	outside := filepath.Join(t.TempDir(), "migrations")
	if _, err := deriveConfigImportPathFromModule(module, root, outside); err == nil {
		t.Fatal("outside module error = nil")
	}
}

func TestGenerateRunnerCodeAndTemporaryBuild(t *testing.T) {
	code, err := generateRevisionRunnerCode(&resolvedRevisionOptions{
		ConfigImportPath: "example.com/app/migrations",
		MigrationDir:     "db/migrations",
		Comment:          "add_users",
		Timeout:          3 * time.Second,
	})
	if err != nil {
		t.Fatalf("generate revision runner: %v", err)
	}
	for _, want := range []string{
		`configpkg "example.com/app/migrations"`,
		`qmig.MakeMigration`,
		`"add_users"`,
		`"db/migrations"`,
	} {
		if !strings.Contains(string(code), want) {
			t.Fatalf("revision runner missing %q:\n%s", want, string(code))
		}
	}

	tmp := t.TempDir()
	// #nosec G306 -- test file should be readable by the test process
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte("module example.com/tmp\n\ngo 1.18\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	goBinary := fakeGoBinary(t)
	if err := runTemporaryMigrationBuild(
		context.Background(),
		tmp,
		goBinary,
		"test-",
		[]byte("package main\n\nfunc main() {}\n"),
	); err != nil {
		t.Fatalf("run temporary build: %v", err)
	}
}

func TestCurrentGoModuleAndDerivedImportPath(t *testing.T) {
	root := t.TempDir()
	goBinary := fakeGoBinaryOutput(t, "example.com/app\n"+root+"\n", "", 0)

	module, err := currentGoModule(context.Background(), goBinary, root)
	if err != nil {
		t.Fatalf("currentGoModule error = %v", err)
	}
	if module.Path != "example.com/app" || module.Dir != root {
		t.Fatalf("module = %#v", module)
	}

	badOutput := fakeGoBinaryOutput(t, "only-one-line\n", "", 0)
	if _, err := currentGoModule(context.Background(), badOutput, root); err == nil {
		t.Fatal("unexpected module output error = nil")
	}

	failing := fakeGoBinaryOutput(t, "", "no module", 1)
	if _, err := currentGoModule(context.Background(), failing, root); err == nil {
		t.Fatal("failing go list error = nil")
	}
}

func fakeGoBinary(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "go.bat")
		// #nosec G306 -- test file should be readable by the test process
		if err := os.WriteFile(path, []byte("@echo off\r\nexit /b 0\r\n"), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}
		return path
	}

	path := filepath.Join(dir, "go")
	// #nosec G306 -- test file should be readable by the test process
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write fake go: %v", err)
	}
	return path
}

func fakeGoBinaryOutput(t *testing.T, stdout, stderr string, exitCode int) string {
	t.Helper()

	dir := t.TempDir()
	if runtime.GOOS == "windows" {
		path := filepath.Join(dir, "go.bat")
		script := "@echo off\r\n"
		for _, line := range strings.Split(strings.TrimSuffix(stdout, "\n"), "\n") {
			if line != "" {
				script += "echo " + line + "\r\n"
			}
		}
		for _, line := range strings.Split(strings.TrimSuffix(stderr, "\n"), "\n") {
			if line != "" {
				script += "echo " + line + " 1>&2\r\n"
			}
		}
		script += fmt.Sprintf("exit /b %d\r\n", exitCode)
		// #nosec G306 -- test file should be readable by the test process
		if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
			t.Fatalf("write fake go: %v", err)
		}
		return path
	}

	path := filepath.Join(dir, "go")
	script := "#!/bin/sh\n"
	if stdout != "" {
		script += "printf '%s' " + shellQuote(stdout) + "\n"
	}
	if stderr != "" {
		script += "printf '%s' " + shellQuote(stderr) + " >&2\n"
	}
	script += fmt.Sprintf("exit %d\n", exitCode)
	// #nosec G306 -- test file should be readable by the test process
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake go: %v", err)
	}
	return path
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
