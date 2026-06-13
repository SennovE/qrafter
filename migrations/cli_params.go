package migrations

import (
	"flag"
	"fmt"
)

func RevisionCommandOptionsFromArgs(args []string) (string, RevisionCommandOptions, error) {
	fs := flag.NewFlagSet("revision-command", flag.ContinueOnError)

	var path string
	var options RevisionCommandOptions

	fs.StringVar(
		&path,
		"path",
		"",
		"Path for generated cmd file",
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
		&options.MigrationDirectory,
		"dir",
		"./migrations",
		"directory for generated migration files",
	)

	fs.StringVar(
		&options.DatabaseDSN,
		"dsn",
		"",
		"Go expression used as database DSN",
	)

	if err := fs.Parse(args); err != nil {
		return "", RevisionCommandOptions{}, err
	}

	if path == "" {
		return "", RevisionCommandOptions{}, fmt.Errorf("missing required flag: --path")
	}

	if options.Dialect == "" {
		return "", RevisionCommandOptions{}, fmt.Errorf("missing required flag: --dialect")
	}

	if options.DriverName == "" {
		return "", RevisionCommandOptions{}, fmt.Errorf("missing required flag: --driver")
	}

	if options.DriverImportPath == "" {
		return "", RevisionCommandOptions{}, fmt.Errorf("missing required flag: --driver-import")
	}

	if options.DatabaseDSN == "" {
		return "", RevisionCommandOptions{}, fmt.Errorf("missing required flag: --dsn")
	}

	return path, options, nil
}
