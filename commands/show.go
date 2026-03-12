package commands

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"os/exec"
	"strings"
)

func (cmds *Commands) newShowCommand() cli.Command {
	return cli.Command{
		Name:      "show",
		Usage:     "Shows host",
		ArgsUsage: "<key>",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "copy, c",
				Usage: "copy ssh command to clipboard",
			},
		},
		Action:    cmds.showAction,
		BashComplete: func(c *cli.Context) {
			// This will complete if no args are passed
			if c.NArg() > 0 {
				return
			}
			cmds.completeWithServers()
		},
	}
}

func (cmds *Commands) showAction(c *cli.Context) (err error) {
	var key string

	if c.NArg() == 0 {
		key, err = cmds.askServerKey()
		if err != nil {
			return err
		}
	} else {
		key = c.Args().First()
	}

	srv := ssh.Config.Get(key)
	if srv == nil {
		return errors.New("host not found")
	}


	fmt.Printf("%s: %s\n", theme.StyleSuccess("Hostname"), theme.StylePrimary(srv.Hostname))
	fmt.Printf("%s: %s\n", theme.StyleSuccess("Port"), theme.StylePrimary(srv.Port))
	fmt.Printf("%s: %s\n", theme.StyleSuccess("User"), theme.StylePrimary(srv.User))
	if srv.IdentityFile != "" {
		fmt.Printf("%s: %s\n", theme.StyleSuccess("IdentityFile"), theme.StylePrimary(srv.IdentityFile))
	}

	if c.Bool("copy") {
		userHost := srv.Hostname
		if srv.User != "" {
			userHost = srv.User + "@" + srv.Hostname
		}
		sshCmd := fmt.Sprintf("ssh %s", userHost)
		if srv.Port != "" {
			sshCmd += " -p " + srv.Port
		}
		if srv.IdentityFile != "" {
			sshCmd += " -i " + srv.IdentityFile
		}

		cp := exec.Command("pbcopy")
		cp.Stdin = strings.NewReader(sshCmd)
		_ = cp.Run()
		fmt.Println(theme.StyleSuccess("✓ SSH command copied to clipboard!"))
	}

	return nil
}
