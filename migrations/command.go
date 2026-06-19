package migrations

import "fmt"

// RunMigrationsCommand dispatches qrafter migration CLI subcommands.
func RunMigrationsCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: qrafter-migrations <init|revision> [flags]")
	}

	switch args[0] {
	case "init":
		return generateMigrationsConfig(args[1:])
	case "revision":
		return generateMigrationRevision(args[1:])
	default:
		return fmt.Errorf("unknown migrations command %q", args[0])
	}
}
