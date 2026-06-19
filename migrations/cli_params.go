package migrations

import (
	"flag"
	"fmt"
	"time"
)

const (
	defaultMigrationDirectory = "./migrations"
	defaultMigrationComment   = "migration"
	defaultRevisionTimeout    = 30 * time.Second
	defaultGoBinary           = "go"
)

func configOptionsFromArgs(args []string) (string, *configOptions, error) {
	fs := flag.NewFlagSet("migrations-init", flag.ContinueOnError)

	var path string
	var options configOptions

	fs.StringVar(
		&path,
		"dir",
		defaultMigrationDirectory,
		"directory for generated qrafter config file",
	)

	fs.StringVar(
		&options.DriverImportPath,
		"driver-import",
		"",
		"Go import path for the database/sql driver",
	)

	fs.StringVar(
		&options.DriverName,
		"driver",
		"",
		"database/sql driver name",
	)

	fs.StringVar(
		&options.Dialect,
		"dialect",
		"",
		"database dialect, for example postgres, mysql, sqlite",
	)

	fs.StringVar(
		&options.DatabaseDSN,
		"dsn",
		"",
		"Go expression used as database DSN",
	)

	if err := fs.Parse(args); err != nil {
		return "", nil, err
	}

	if path == "" {
		return "", nil, fmt.Errorf("missing required flag: --dir")
	}

	if options.Dialect == "" {
		return "", nil, fmt.Errorf("missing required flag: --dialect")
	}

	if options.DriverName == "" {
		return "", nil, fmt.Errorf("missing required flag: --driver")
	}

	if options.DriverImportPath == "" {
		return "", nil, fmt.Errorf("missing required flag: --driver-import")
	}

	if options.DatabaseDSN == "" {
		return "", nil, fmt.Errorf("missing required flag: --dsn")
	}

	return path, &options, nil
}

type revisionOptions struct {
	WorkDir          string
	MigrationDir     string
	ConfigImportPath string
	Comment          string
	GoBinary         string
	Timeout          time.Duration
}

func revisionOptionsFromArgs(args []string) (*revisionOptions, error) {
	fs := flag.NewFlagSet("migrations-revision", flag.ContinueOnError)

	options := revisionOptions{
		WorkDir:      ".",
		MigrationDir: defaultMigrationDirectory,
		Comment:      defaultMigrationComment,
		GoBinary:     defaultGoBinary,
		Timeout:      defaultRevisionTimeout,
	}

	fs.StringVar(
		&options.WorkDir,
		"workdir",
		options.WorkDir,
		"project directory containing go.mod",
	)

	fs.StringVar(
		&options.MigrationDir,
		"dir",
		options.MigrationDir,
		"directory containing qrafter_config.go and generated migrations",
	)

	fs.StringVar(
		&options.ConfigImportPath,
		"config-import",
		"",
		"Go import path for package containing MigrationConfig",
	)

	fs.StringVar(
		&options.Comment,
		"comment",
		options.Comment,
		"migration comment",
	)

	fs.StringVar(
		&options.GoBinary,
		"go",
		options.GoBinary,
		"go binary used for the temporary build",
	)

	fs.DurationVar(
		&options.Timeout,
		"timeout",
		options.Timeout,
		"timeout for the temporary build and schema diff",
	)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if options.WorkDir == "" {
		return nil, fmt.Errorf("missing required flag: --workdir")
	}
	if options.MigrationDir == "" {
		return nil, fmt.Errorf("missing required flag: --dir")
	}
	if options.GoBinary == "" {
		return nil, fmt.Errorf("missing required flag: --go")
	}
	if options.Timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}

	return &options, nil
}

type applyOptions struct {
	WorkDir          string
	MigrationDir     string
	ConfigImportPath string
	Target           string
	VersionTable     string
	GoBinary         string
	Timeout          time.Duration
	Direction        MigrationDirection
}

func applyOptionsFromArgs(direction MigrationDirection, args []string) (*applyOptions, error) {
	fs := flag.NewFlagSet("migrations-"+string(direction), flag.ContinueOnError)

	options := applyOptions{
		WorkDir:      ".",
		MigrationDir: defaultMigrationDirectory,
		GoBinary:     defaultGoBinary,
		Timeout:      defaultRevisionTimeout,
		Direction:    direction,
	}

	fs.StringVar(
		&options.WorkDir,
		"workdir",
		options.WorkDir,
		"project directory containing go.mod",
	)

	fs.StringVar(
		&options.MigrationDir,
		"dir",
		options.MigrationDir,
		"directory containing qrafter_config.go and generated migrations",
	)

	fs.StringVar(
		&options.ConfigImportPath,
		"config-import",
		"",
		"Go import path for package containing MigrationConfig and Registry",
	)

	fs.StringVar(
		&options.Target,
		"to",
		"",
		"target migration version, head, or base",
	)

	fs.StringVar(
		&options.VersionTable,
		"version-table",
		"",
		"database table used to store the current migration version; defaults to MigrationConfig.VersionTable",
	)

	fs.StringVar(
		&options.GoBinary,
		"go",
		options.GoBinary,
		"go binary used for the temporary build",
	)

	fs.DurationVar(
		&options.Timeout,
		"timeout",
		options.Timeout,
		"timeout for the temporary build and migration execution",
	)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if options.WorkDir == "" {
		return nil, fmt.Errorf("missing required flag: --workdir")
	}
	if options.MigrationDir == "" {
		return nil, fmt.Errorf("missing required flag: --dir")
	}
	if options.GoBinary == "" {
		return nil, fmt.Errorf("missing required flag: --go")
	}
	if options.Timeout <= 0 {
		return nil, fmt.Errorf("timeout must be positive")
	}
	if options.Direction != DirectionUp && options.Direction != DirectionDown {
		return nil, fmt.Errorf("unsupported migration direction %q", options.Direction)
	}

	return &options, nil
}
