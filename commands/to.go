package commands

import (
	"fmt"
	"github.com/trntv/sshed/host"
	"github.com/trntv/sshed/keychain"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"os"
)

func (cmds *Commands) newToCommand() cli.Command {
	return cli.Command{
		Name:      "to",
		Usage:     "Connects to host",
		ArgsUsage: "<key>",
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "verbose, v",
				Usage: "verbose ssh output",
			},
		},
		BashComplete: func(c *cli.Context) {
			// This will complete if no args are passed
			if c.NArg() > 0 {
				return
			}
			cmds.completeWithServers()
		},
		Action: cmds.toAction,
	}
}

func (cmds *Commands) toAction(c *cli.Context) (err error) {
	var key string
	var srv *host.Host

	if c.NArg() == 0 {
		key, err = cmds.askServerKey()
		if err != nil {
			return err
		}
	} else {
		key = c.Args().First()
	}

	srv = ssh.Config.Get(key)
	if srv == nil {
		return fmt.Errorf(theme.StyleError("host not found"))
	}

	// Interactive "Save Password" flow if missing and no keys are present
	if srv.Password() == "" && srv.PrivateKey() == "" && srv.IdentityFile == "" {
		var save bool
		_ = survey.AskOne(&survey.Confirm{
			Message: "Password not found in Keychain. Do you want to save it now to enable Touch ID?",
			Default: true,
		}, &save, nil)

		if save {
			pwd := ""
			prompt := &survey.Password{
				Message: "Enter password for " + srv.Key + ":",
			}
			err = survey.AskOne(prompt, &pwd, survey.Required)
			if err == nil {
				// Get current record or create new one
				rec, _ := keychain.Get(srv.Key)
				if rec == nil {
					rec = &keychain.Record{}
				}
				rec.Password = pwd
				err = keychain.Put(srv.Key, rec)
				if err != nil {
					return fmt.Errorf(theme.StyleError("failed to save password to keychain: " + err.Error()))
				}
				fmt.Printf("%s\n", theme.StyleSuccess("✓ Securely saved to macOS Keychain. Next time you will be prompted for Touch ID."))
			}
		}
	}

	cmd, err := cmds.createCommand(c, srv, &options{verbose: c.Bool("verbose")}, "")
	if err != nil {
		return err
	}

	return cmds.RunCommand(cmd, srv, os.Stdout, os.Stderr)
}
