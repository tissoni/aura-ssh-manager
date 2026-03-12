package commands

import (
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/ui"
	"github.com/urfave/cli"
	"os"
)

func (cmds *Commands) newDashboardCommand() cli.Command {
	return cli.Command{
		Name:    "dash",
		Aliases: []string{"d"},
		Usage:   "Aura Dashboard - Interactive TUI",
		Action: func(c *cli.Context) error {
			choice, err := ui.ShowDashboard()
			if err != nil {
				return err
			}

			if choice != "" {
				srv := ssh.Config.Get(choice)
				if srv != nil {
					cmd, err := cmds.createCommand(c, srv, &options{}, "")
					if err != nil {
						return err
					}
					return cmds.RunCommand(cmd, srv, os.Stdout, os.Stderr)
				}
			}
			return nil
		},
	}
}
