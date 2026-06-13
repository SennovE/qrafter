package main

import (
	"os"
	
	qmig "github.com/SennovE/qrafter/migrations"
)

func main() {
	switch os.Args[1] {
	case "init":
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
