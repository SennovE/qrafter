// Command qrafter-migrations scaffolds helper commands for qrafter migrations.
package main

import (
	"os"

	qmig "github.com/SennovE/qrafter/migrations"
)

func main() {
	if err := qmig.GenerateMigrationsConfig(os.Args[1:]); err != nil {
		panic(err)
	}
}
