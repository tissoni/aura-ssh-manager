package commands

import (
	"fmt"
	"github.com/trntv/sshed/snippets"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"os"
	"sort"
)

func (cmds *Commands) newSnippetsCommand() cli.Command {
	return cli.Command{
		Name:    "snippets",
		Aliases: []string{"snip"},
		Usage:   "Manage and execute command snippets",
		Subcommands: []cli.Command{
			{
				Name:   "ls",
				Usage:  "List all snippets",
				Action: cmds.snippetsListAction,
			},
			{
				Name:   "add",
				Usage:  "Add a new snippet",
				Action: cmds.snippetsAddAction,
			},
			{
				Name:   "rm",
				Usage:  "Remove a snippet",
				Action: cmds.snippetsRemoveAction,
			},
			{
				Name:      "run",
				Usage:     "Run a snippet on a host",
				ArgsUsage: "[host] [snippet-name]",
				Action:    cmds.snippetsRunAction,
			},
		},
	}
}

func (cmds *Commands) snippetsListAction(c *cli.Context) error {
	s := snippets.Load()
	if len(s.Items) == 0 {
		fmt.Println(theme.StyleSecondary("No snippets found."))
		return nil
	}

	fmt.Println(theme.StyleHeader(" AURA SNIPPETS "))
	keys := make([]string, 0, len(s.Items))
	for k := range s.Items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		snip := s.Items[k]
		fmt.Printf("➜ %s: %s\n", theme.StylePrimary(snip.Name), theme.StyleSecondary(snip.Command))
	}
	return nil
}

func (cmds *Commands) snippetsAddAction(c *cli.Context) error {
	qs := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Snippet Name (slug):",
			},
			Validate: survey.Required,
		},
		{
			Name: "command",
			Prompt: &survey.Input{
				Message: "Command:",
			},
			Validate: survey.Required,
		},
	}

	answers := struct {
		Name    string
		Command string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}

	s := snippets.Load()
	s.Add(answers.Name, answers.Command)
	err = s.Save()
	if err != nil {
		return err
	}

	fmt.Println(theme.StyleSuccess("✓ Snippet saved successfully."))
	return nil
}

func (cmds *Commands) snippetsRemoveAction(c *cli.Context) error {
	s := snippets.Load()
	options := make([]string, 0, len(s.Items))
	for k := range s.Items {
		options = append(options, k)
	}
	sort.Strings(options)

	var selected string
	err := survey.AskOne(&survey.Select{
		Message: "Choose snippet to remove:",
		Options: options,
	}, &selected, nil)
	if err != nil {
		return err
	}

	s.Remove(selected)
	_ = s.Save()
	fmt.Println(theme.StyleSuccess("✓ Snippet removed."))
	return nil
}

func (cmds *Commands) snippetsRunAction(c *cli.Context) error {
	hostKey := c.Args().Get(0)
	snipName := c.Args().Get(1)

	if hostKey == "" {
		var err error
		hostKey, err = cmds.askServerKey()
		if err != nil {
			return err
		}
	}

	s := snippets.Load()
	if snipName == "" {
		options := make([]string, 0, len(s.Items))
		for k := range s.Items {
			options = append(options, k)
		}
		sort.Strings(options)

		err := survey.AskOne(&survey.Select{
			Message: "Choose snippet to run:",
			Options: options,
		}, &snipName, nil)
		if err != nil {
			return err
		}
	}

	snip, ok := s.Items[snipName]
	if !ok {
		return fmt.Errorf("snippet '%s' not found", snipName)
	}

	// Reuse existing 'at' logic essentially
	// For simplicity, we can just call RunCommand after creating it
	// But 'at' action is more robust. We'll implement a clean run here.
	
	srv := ssh.Config.Get(hostKey)
	if srv == nil {
		return fmt.Errorf("host not found: %s", hostKey)
	}

	fmt.Printf("%s %s %s %s\n", theme.StyleSecondary("Running snippet"), theme.StylePrimary(snip.Name), theme.StyleSecondary("on"), theme.StyleSecondary(srv.Hostname))

	cmd, err := cmds.createCommand(c, srv, &options{}, snip.Command)
	if err != nil {
		return err
	}

	return cmds.RunCommand(cmd, srv, os.Stdout, os.Stderr)
}
