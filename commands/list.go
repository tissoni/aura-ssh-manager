package commands

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/trntv/sshed/health"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"sort"
)

func (cmds *Commands) newListCommand() cli.Command {
	return cli.Command{
		Name:    "ls",
		Aliases: []string{"list"},
		Usage:   "Lists all hosts",
		Action:  cmds.listAction,
	}
}

func (cmds *Commands) listAction(ctx *cli.Context) error {
	hosts := ssh.Config.GetAll()
	if len(hosts) == 0 {
		fmt.Println(theme.StyleError("Servers list is empty"))
		return nil
	}

	fmt.Println(theme.StyleHeader(" CONFIGURED HOSTS "))

	keys := make([]string, 0, len(hosts))
	for key := range hosts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		srv := hosts[key]
		
		status := health.Check(srv.Hostname, srv.Port)
		statusDot := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render("●")
		if status == health.StatusUp {
			statusDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("●")
		} else if status == health.StatusDown {
			statusDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("●")
		}

		tags := ""
		if len(srv.Tags()) > 0 {
			for _, t := range srv.Tags() {
				tags += theme.StyleTag(t) + " "
			}
		}
		fmt.Printf("%s %s %s\n", statusDot, theme.StylePrimary(key), tags)
	}

	return nil
}
