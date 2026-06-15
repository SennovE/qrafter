// Command qrafter-migrations scaffolds helper commands for qrafter migrations.
package main

import (
	"os"

	qmig "github.com/SennovE/qrafter/migrations"
)

func main() {
	if os.Args[1] == "init" {
		path, options, err := qmig.RevisionCommandOptionsFromArgs(os.Args[2:])
		if err != nil {
			panic(err)
		}
		err = qmig.WriteRevisionCommandFile(path, options)
		if err != nil {
			panic(err)
		}
	}
}
