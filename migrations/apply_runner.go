package migrations

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const applyRunnerTemplate = `package main

import (
	"context"
	"fmt"
	"time"

	configpkg {{.ConfigImportPath}}
	qmig "github.com/SennovE/qrafter/migrations"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), {{.TimeoutNanoseconds}}*time.Nanosecond)
	defer cancel()

	config := configpkg.MigrationConfig
	versionTable := config.VersionTable
	if {{.VersionTable}} != "" {
		versionTable = {{.VersionTable}}
	}

	result, err := qmig.ApplyMigrations(ctx, &qmig.MigrationApplyConfig{
		DriverName:     config.DriverName,
		DataSourceName: config.DataSourceName,
		Dialect:        config.Dialect,
		Registry:       configpkg.Registry,
		Direction:      {{.Direction}},
		Target:         {{.Target}},
		VersionTable:    versionTable,
	})
	if err != nil {
		panic(err)
	}

	if len(result.Applied) == 0 {
		fmt.Printf("already at %s\n", migrationVersionLabel(result.To))
		return
	}

	for _, version := range result.Applied {
		fmt.Printf("%s %s\n", {{.StepVerb}}, version)
	}
	fmt.Printf("migrated from %s to %s\n", migrationVersionLabel(result.From), migrationVersionLabel(result.To))
}

func migrationVersionLabel(version string) string {
	if version == "" {
		return "base"
	}
	return version
}
`

type applyRunnerTemplateData struct {
	ConfigImportPath   string
	Direction          string
	Target             string
	VersionTable       string
	StepVerb           string
	TimeoutNanoseconds int64
}

type resolvedApplyOptions struct {
	WorkDir          string
	ConfigImportPath string
	Target           string
	VersionTable     string
	GoBinary         string
	Timeout          time.Duration
	Direction        MigrationDirection
}

func applyMigrationsCommand(direction MigrationDirection, args []string) error {
	options, err := applyOptionsFromArgs(direction, args)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), options.Timeout)
	defer cancel()

	return runMigrationApply(ctx, options)
}

func runMigrationApply(ctx context.Context, options *applyOptions) error {
	resolved, err := resolveApplyOptions(ctx, options)
	if err != nil {
		return err
	}

	code, err := generateApplyRunnerCode(resolved)
	if err != nil {
		return err
	}

	return runTemporaryMigrationBuild(ctx, resolved.WorkDir, resolved.GoBinary, string(resolved.Direction)+"-", code)
}

func resolveApplyOptions(
	ctx context.Context,
	options *applyOptions,
) (*resolvedApplyOptions, error) {
	if options == nil {
		return nil, fmt.Errorf("migrations: apply options are nil")
	}

	workDir, err := filepathAbs(options.WorkDir)
	if err != nil {
		return nil, err
	}

	resolved := &resolvedApplyOptions{
		WorkDir:      workDir,
		Target:       options.Target,
		VersionTable: options.VersionTable,
		GoBinary:     options.GoBinary,
		Timeout:      options.Timeout,
		Direction:    options.Direction,
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

func filepathAbs(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve workdir: %w", err)
	}
	return absPath, nil
}

func generateApplyRunnerCode(options *resolvedApplyOptions) ([]byte, error) {
	data := applyRunnerTemplateData{
		ConfigImportPath:   strconv.Quote(options.ConfigImportPath),
		Direction:          directionCode(options.Direction),
		Target:             strconv.Quote(options.Target),
		VersionTable:       strconv.Quote(options.VersionTable),
		StepVerb:           strconv.Quote(stepVerb(options.Direction)),
		TimeoutNanoseconds: options.Timeout.Nanoseconds(),
	}
	return renderGoTemplate("migration apply runner", applyRunnerTemplate, data)
}

func directionCode(direction MigrationDirection) string {
	switch direction {
	case DirectionUp:
		return "qmig.DirectionUp"
	case DirectionDown:
		return "qmig.DirectionDown"
	default:
		return "qmig.MigrationDirection(" + strconv.Quote(string(direction)) + ")"
	}
}

func stepVerb(direction MigrationDirection) string {
	if direction == DirectionDown {
		return "reverted"
	}
	return "applied"
}
