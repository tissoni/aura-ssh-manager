package commands

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func (cmds *Commands) newAtCommand() cli.Command {
	return cli.Command{
		Name:      "at",
		Usage:     "Executes commands",
		ArgsUsage: "[key] [command]",
		Action:    cmds.atAction,
		BashComplete: func(c *cli.Context) {
			// This will complete if no args are passed
			if c.NArg() > 0 {
				return
			}
			cmds.completeWithServers()
		},
	}
}
func (cmds *Commands) atAction(c *cli.Context) (err error) {
	arg := c.Args().First()
	var keys []string

	if arg == "" {
		keys, err = cmds.askServersKeys()
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(arg, "@") {
		tag := strings.TrimPrefix(arg, "@")
		allHosts := ssh.Config.GetAll()
		for k, h := range allHosts {
			for _, t := range h.Tags() {
				if t == tag {
					keys = append(keys, k)
					break
				}
			}
		}
		if len(keys) == 0 {
			return fmt.Errorf("no hosts found with tag: %s", tag)
		}
		fmt.Printf("%s: %s\n", theme.StyleSecondary("Targeting hosts with tag"), theme.StylePrimary(strings.Join(keys, ", ")))
	} else {
		keys = []string{arg}
	}

	command := c.Args().Get(1)
	if command == "" {

		err = survey.AskOne(&survey.Input{Message: "Command:"}, &command, nil)
		if err != nil {
			return err
		}

		fmt.Println("")
	}

	var wg sync.WaitGroup
	for _, key := range keys {
		var srv = ssh.Config.Get(key)
		if srv == nil {
			return errors.New("host not found")
		}

		if err != nil {
			return err
		}

		wg.Add(1)
		go (func() {
			defer wg.Done()

			cmd, err := cmds.createCommand(c, srv, &options{}, command)
			if err != nil {
				log.Panicln(err)
			}

			var buf []byte
			w := bytes.NewBuffer(buf)

			err = cmds.RunCommand(cmd, srv, w, os.Stderr)
			if err != nil {
				log.Panicln(err)
			}

			sr, err := ioutil.ReadAll(w)
			if err != nil {
				log.Panicln(err)
			}

			fmt.Printf("%s:\r\n", theme.StyleAccent(srv.Key))
			fmt.Println(string(sr))
		})()
	}

	wg.Wait()

	// macOS Notification with sound and Aura branding
	_ = exec.Command("osascript", "-e", fmt.Sprintf(`display notification "Tasks completed on %d hosts successfully." with title "Aura" subtitle "Batch execution finished" sound name "Crystal"`, len(keys))).Run()

	return err
}
