package migrations

import (
	"flag"
	"fmt"
)

func revisionCommandOptionsFromArgs(args []string) (string, *configOptions, error) {
	fs := flag.NewFlagSet("revision-command", flag.ContinueOnError)

	var path string
	var options configOptions

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
		&options.DatabaseDSN,
		"dsn",
		"",
		"Go expression used as database DSN",
	)

	if err := fs.Parse(args); err != nil {
		return "", nil, err
	}

	if path == "" {
		return "", nil, fmt.Errorf("missing required flag: --path")
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
