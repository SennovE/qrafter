// Command qrafter-migrations runs qrafter migration tools.
package main

import (
	"os"

	qmig "github.com/SennovE/qrafter/migrations"
)

func main() {
	if err := qmig.RunMigrationsCommand(os.Args[1:]); err != nil {
		panic(err)
	}
}
