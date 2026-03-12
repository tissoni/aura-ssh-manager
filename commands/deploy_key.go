package commands

import (
	"fmt"

	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/trntv/sshed/ui"
	"github.com/urfave/cli"
)

func (cmds *Commands) newDeployKeyCommand() cli.Command {
	return cli.Command{
		Name:      "deploy-key",
		Usage:     "Deploy SSH public keys to remote servers",
		ArgsUsage: "[server]",
		Action: func(c *cli.Context) error {
			var hostKey string
			if c.NArg() > 0 {
				hostKey = c.Args().Get(0)
			} else {
				var err error
				hostKey, err = cmds.askServerKey()
				if err != nil {
					return err
				}
			}

			srv := ssh.Config.Get(hostKey)
			if srv == nil {
				return fmt.Errorf("%s: %s", theme.StyleError("host not found"), hostKey)
			}

			return ui.ShowDeployKey(srv)
		},
	}
}
