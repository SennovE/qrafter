package migrations

import (
	"errors"
	"flag"
	"fmt"
)

const migrationsHelpText = `Usage:
  qrafter-migrations <command> [flags]

Commands:
  init      create qrafter_config.go in the migrations directory
  revision  generate a new migration from the database/schema diff
  up        apply registered migrations forward
  down      revert registered migrations
  help      show this help or command-specific help

Examples:
  qrafter-migrations init --dir ./migrations --driver-import github.com/lib/pq --driver postgres --dialect postgres --dsn postgres://localhost:5432/app?sslmode=disable
  qrafter-migrations revision --dir ./migrations --comment create_users
  qrafter-migrations up --dir ./migrations --to head
  qrafter-migrations down --dir ./migrations --to base

Use "qrafter-migrations help <command>" for command-specific flags.
`

const (
	commandInit     = "init"
	commandRevision = "revision"
	commandUp       = string(DirectionUp)
	commandDown     = string(DirectionDown)
	commandHelp     = "help"
	commandHelpLong = "--help"
	commandHelpFlag = "-h"
)

var migrationCommandHelp = map[string]string{
	commandInit: `Usage:
  qrafter-migrations init --dir ./migrations --driver-import <import> --driver <name> --dialect <dialect> --dsn <dsn>

Flags:
  --dir            directory for qrafter_config.go
  --driver-import  Go import path for the database/sql driver
  --driver         database/sql driver name
  --dialect        database dialect, for example postgres
  --dsn            database DSN
`,
	commandRevision: `Usage:
  qrafter-migrations revision [flags]

Flags:
  --workdir        project directory containing go.mod
  --dir            directory containing qrafter_config.go and generated migrations
  --config-import  Go import path for package containing MigrationConfig
  --comment        migration file comment
  --go             go binary used for the temporary build
  --timeout        timeout for the temporary build and schema diff
`,
	commandUp: `Usage:
  qrafter-migrations up [flags]

Flags:
  --workdir        project directory containing go.mod
  --dir            directory containing qrafter_config.go and generated migrations
  --config-import  Go import path for package containing MigrationConfig and Registry
  --to             target migration version or head
  --version-table  table used to store current migration version
  --go             go binary used for the temporary build
  --timeout        timeout for the temporary build and migration execution
`,
	commandDown: `Usage:
  qrafter-migrations down [flags]

Flags:
  --workdir        project directory containing go.mod
  --dir            directory containing qrafter_config.go and generated migrations
  --config-import  Go import path for package containing MigrationConfig and Registry
  --to             target migration version or base
  --version-table  table used to store current migration version
  --go             go binary used for the temporary build
  --timeout        timeout for the temporary build and migration execution
`,
}

// RunMigrationsCommand dispatches qrafter migration CLI subcommands.
func RunMigrationsCommand(args []string) error {
	if len(args) == 0 {
		return runMigrationsHelp(nil)
	}

	if isHelpArg(args[0]) {
		return runMigrationsHelp(args[1:])
	}

	var err error
	switch args[0] {
	case commandInit:
		err = generateMigrationsConfig(args[1:])
	case commandRevision:
		err = generateMigrationRevision(args[1:])
	case commandUp:
		err = applyMigrationsCommand(DirectionUp, args[1:])
	case commandDown:
		err = applyMigrationsCommand(DirectionDown, args[1:])
	default:
		return fmt.Errorf("unknown migrations command %q", args[0])
	}
	if errors.Is(err, flag.ErrHelp) {
		return nil
	}
	return err
}

func runMigrationsHelp(args []string) error {
	if len(args) == 0 {
		fmt.Print(migrationsHelpText)
		return nil
	}

	text, ok := migrationCommandHelp[args[0]]
	if !ok {
		return fmt.Errorf("unknown migrations command %q", args[0])
	}
	fmt.Print(text)
	return nil
}

func isHelpArg(arg string) bool {
	return arg == commandHelp || arg == commandHelpFlag || arg == commandHelpLong
}
