package migrations

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const revisionRunnerTemplate = `package main

import (
	"context"
	"fmt"
	"time"

	configpkg "{{.ConfigImportPath}}"
	qmig "github.com/SennovE/qrafter/migrations"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), {{.TimeoutNanoseconds}}*time.Nanosecond)
	defer cancel()

	config := configpkg.MigrationConfig
	path, err := qmig.MakeMigration(ctx, "{{.Comment}}", "{{.MigrationDir}}", &config)
	if err != nil {
		panic(err)
	}

	fmt.Printf("created %s\n", path)
}
`

type revisionRunnerTemplateData struct {
	ConfigImportPath   string
	MigrationDir       string
	Comment            string
	TimeoutNanoseconds int64
}

type resolvedRevisionOptions struct {
	WorkDir          string
	MigrationDir     string
	ConfigImportPath string
	Comment          string
	GoBinary         string
	Timeout          time.Duration
}

type goModuleInfo struct {
	Path string
	Dir  string
}

func generateMigrationRevision(args []string) error {
	options, err := revisionOptionsFromArgs(args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	return runMigrationRevision(ctx, options)
}

func runMigrationRevision(ctx context.Context, options *revisionOptions) error {
	resolved, err := resolveRevisionOptions(ctx, options)
	if err != nil {
		return err
	}

	code, err := generateRevisionRunnerCode(resolved)
	if err != nil {
		return err
	}

	tmpRoot := filepath.Join(resolved.WorkDir, ".qrafter", "tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return fmt.Errorf("create temporary migration root: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpRoot)
		_ = os.Remove(filepath.Dir(tmpRoot))
	}()

	tempDir, err := os.MkdirTemp(tmpRoot, "revision-")
	if err != nil {
		return fmt.Errorf("create temporary migration build directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	if err := createFile(filepath.Join(tempDir, "main.go"), code); err != nil {
		return fmt.Errorf("create temporary migration build file: %w", err)
	}

	return runTemporaryRevisionBuild(ctx, resolved, tempDir)
}

func resolveRevisionOptions(
	ctx context.Context,
	options *revisionOptions,
) (*resolvedRevisionOptions, error) {
	if options == nil {
		return nil, fmt.Errorf("migrations: revision options are nil")
	}

	workDir, err := filepath.Abs(options.WorkDir)
	if err != nil {
		return nil, fmt.Errorf("resolve workdir: %w", err)
	}

	resolved := &resolvedRevisionOptions{
		WorkDir:      workDir,
		MigrationDir: options.MigrationDir,
		Comment:      options.Comment,
		GoBinary:     options.GoBinary,
		Timeout:      options.Timeout,
	}

	if strings.TrimSpace(options.ConfigImportPath) != "" {
		resolved.ConfigImportPath = strings.TrimSpace(options.ConfigImportPath)
		return resolved, nil
	}

	configImportPath, err := deriveConfigImportPath(ctx, options.GoBinary, workDir, options.MigrationDir)
	if err != nil {
		return nil, err
	}
	resolved.ConfigImportPath = configImportPath
	return resolved, nil
}

func deriveConfigImportPath(
	ctx context.Context,
	goBinary string,
	workDir string,
	migrationDir string,
) (string, error) {
	module, err := currentGoModule(ctx, goBinary, workDir)
	if err != nil {
		return "", err
	}
	return deriveConfigImportPathFromModule(module, workDir, migrationDir)
}

func deriveConfigImportPathFromModule(
	module goModuleInfo,
	workDir string,
	migrationDir string,
) (string, error) {
	absMigrationDir := migrationDir
	if !filepath.IsAbs(absMigrationDir) {
		absMigrationDir = filepath.Join(workDir, migrationDir)
	}
	absMigrationDir, err := filepath.Abs(absMigrationDir)
	if err != nil {
		return "", fmt.Errorf("resolve migrations directory: %w", err)
	}

	rel, err := filepath.Rel(module.Dir, absMigrationDir)
	if err != nil {
		return "", fmt.Errorf("resolve migrations import path: %w", err)
	}
	if rel == "." {
		return module.Path, nil
	}
	if filepath.IsAbs(rel) || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("migrations directory %s is outside module %s", absMigrationDir, module.Dir)
	}
	return module.Path + "/" + filepath.ToSlash(rel), nil
}

func currentGoModule(ctx context.Context, goBinary, workDir string) (goModuleInfo, error) {
	// #nosec G204 -- goBinary is an explicit CLI option for selecting the Go tool.
	cmd := exec.CommandContext(ctx, goBinary, "list", "-m", "-f", "{{.Path}}\n{{.Dir}}")
	cmd.Dir = workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return goModuleInfo{}, fmt.Errorf("read current go module: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 2 || strings.TrimSpace(lines[0]) == "" || strings.TrimSpace(lines[1]) == "" {
		return goModuleInfo{}, fmt.Errorf("unexpected go module output: %q", string(out))
	}
	return goModuleInfo{
		Path: strings.TrimSpace(lines[0]),
		Dir:  strings.TrimSpace(lines[1]),
	}, nil
}

func generateRevisionRunnerCode(options *resolvedRevisionOptions) ([]byte, error) {
	data := revisionRunnerTemplateData{
		ConfigImportPath:   options.ConfigImportPath,
		MigrationDir:       options.MigrationDir,
		Comment:            options.Comment,
		TimeoutNanoseconds: options.Timeout.Nanoseconds(),
	}
	return renderGoTemplate("migration revision runner", revisionRunnerTemplate, data)
}

func runTemporaryRevisionBuild(
	ctx context.Context,
	options *resolvedRevisionOptions,
	tempDir string,
) error {
	rel, err := filepath.Rel(options.WorkDir, tempDir)
	if err != nil {
		return fmt.Errorf("resolve temporary build package: %w", err)
	}

	pkg := "./" + filepath.ToSlash(rel)
	// #nosec G204 -- goBinary is an explicit CLI option for selecting the Go tool.
	cmd := exec.CommandContext(ctx, options.GoBinary, "run", pkg)
	cmd.Dir = options.WorkDir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run temporary migration build: %w", err)
	}
	return nil
}
