package main

import (
	"git.sr.ht/~humaid/neatnote/cmd"
	"log"
	"os"

	"github.com/urfave/cli"
)

// VERSION specifies the version of neatnote
var VERSION = "0.1.0"

func main() {
	app := cli.NewApp()
	app.Name = "neatnote"
	app.Usage = "a web app to allow University students to post notes in a civil manner."
	app.Version = VERSION
	app.Commands = []cli.Command{
		cmd.CmdStart,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
