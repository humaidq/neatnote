package main

import (
	"git.sr.ht/~humaid/neatnote/cmd"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

// VERSION specifies the version of neatnote
var VERSION = "0.1.0"

func main() {
	app := &cli.App{
		Name:    "neatnote",
		Usage:   "a web app to allow University students to post notes in a civil manner.",
		Version: VERSION,
		Commands: []*cli.Command{
			cmd.CmdStart,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
