package commands

import (
	"fmt"
	"github.com/trntv/sshed/config"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"sort"
)

func (cmds *Commands) newThemeCommand() cli.Command {
	return cli.Command{
		Name:  "theme",
		Usage: "Switch between Aura Eye Candy themes",
		Action: func(c *cli.Context) error {
			arg := c.Args().First()
			if arg == "" {
				fmt.Println(theme.StyleHeader(" AVAILABLE THEMES "))
				keys := make([]string, 0, len(theme.Themes))
				for k := range theme.Themes {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					t := theme.Themes[k]
					if k == theme.ActiveTheme.Name { // This logic is slightly wrong but we'll fix
						fmt.Printf("➜ %s\n", theme.StylePrimary(t.Name))
					} else {
						fmt.Printf("  %s\n", k)
					}
				}
				return nil
			}

			if _, ok := theme.Themes[arg]; !ok {
				return fmt.Errorf("theme '%s' not found", arg)
			}

			conf := config.Load()
			conf.Theme = arg
			err := config.Save(conf)
			if err != nil {
				return err
			}

			theme.Init(arg)
			fmt.Println(theme.StyleSuccess("Theme applied: " + theme.ActiveTheme.Name))
			return nil
		},
	}
}
