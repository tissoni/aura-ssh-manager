package commands

import (
	"fmt"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"sort"
)

func (cmds *Commands) newIdentitiesCommand() cli.Command {
	return cli.Command{
		Name:    "identities",
		Aliases: []string{"id"},
		Usage:   "Lists all saved SSH identities (private keys)",
		Action:  cmds.identitiesAction,
	}
}

func (cmds *Commands) identitiesAction(ctx *cli.Context) error {
	keys := ssh.Config.Keys
	if len(keys) == 0 {
		fmt.Println(theme.StyleSecondary("No SSH identities found in Aura's secure storage."))
		return nil
	}

	fmt.Println(theme.StyleHeader(" SAVED SSH IDENTITIES "))

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("🗝  %s\n", theme.StylePrimary(key))
	}

	fmt.Printf("\n%s\n", theme.StyleSecondary("You can associate these with hosts using 'aura add'."))

	return nil
}
