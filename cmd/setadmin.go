package cmd

import (
	"errors"
	"fmt"
	"git.sr.ht/~humaid/neatnote/models"
	"git.sr.ht/~humaid/neatnote/modules/settings"
	"github.com/urfave/cli/v2"
)

// CmdSetAdmin represents a command-line command
// which sets a user to become an admin.
var CmdSetAdmin = &cli.Command{
	Name:  "set-admin",
	Usage: "Gives a user admin status",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "status",
			Value: "true",
			Usage: "admin status",
		},
	},
	Action: setAdmin,
}

func setAdmin(c *cli.Context) error {
	if c.NArg() < 1 {
		return errors.New("Need username to set")
	}
	settings.LoadConfig()
	engine := models.SetupEngine()
	defer engine.Close()
	u, err := models.GetUser(c.Args().Get(0))
	if err != nil {
		return err
	}
	fmt.Printf("Current admin status: %t\n", u.IsAdmin)

	if c.String("status") == "true" {
		u.IsAdmin = true
		fmt.Println("Setting the user to admin")
	} else if c.String("status") == "false" {
		u.IsAdmin = false
		fmt.Println("Setting the user to non-admin")
	} else {
		return errors.New("Unknown status. Must be either true or false.")
	}

	return models.UpdateUserAdmin(u)
}
