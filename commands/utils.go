package commands

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/trntv/sshed/theme"
	"github.com/trntv/sshed/ui"
	"github.com/urfave/cli"
)

func (cmds *Commands) newUtilsCommand() cli.Command {
	return cli.Command{
		Name:  "utils",
		Usage: "Security and maintenance utilities",
		Subcommands: []cli.Command{
			{
				Name:  "gen-pwd",
				Usage: "Generate a strong random password",
				Flags: []cli.Flag{
					cli.IntFlag{Name: "length, l", Value: 16, Usage: "Password length"},
				},
				Action: cmds.genPwdAction,
			},
			{
				Name:  "gen-passphrase",
				Usage: "Generate a secure passphrase (CorrectHorseBatteryStaple style)",
				Flags: []cli.Flag{
					cli.IntFlag{Name: "words, w", Value: 4, Usage: "Number of words"},
				},
				Action: cmds.genPassphraseAction,
			},
			{
				Name:  "ports",
				Usage: "Find and kill processes on active ports (macOS)",
				Action: func(c *cli.Context) error {
					return ui.ShowPortKiller()
				},
			},
		},
	}
}

func (cmds *Commands) genPwdAction(c *cli.Context) error {
	length := c.Int("length")
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+"
	
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}

	fmt.Printf("\n%s\n%s\n\n", theme.StyleSecondary("Generated Password:"), theme.StyleSuccess(string(result)))
	return nil
}

func (cmds *Commands) genPassphraseAction(c *cli.Context) error {
	numWords := c.Int("words")
	
	// A small but sufficient wordlist for demonstration. 
	// In a real app we might load a larger file, but 200+ words are fine for now.
	wordlist := []string{
		"apple", "beach", "brain", "bread", "brush", "chair", "chest", "chord", "click", "clock",
		"cloud", "dance", "diary", "drink", "earth", "feast", "field", "flame", "glass", "heart",
		"house", "juice", "light", "money", "music", "night", "ocean", "party", "piano", "pilot",
		"plane", "plant", "radio", "river", "scene", "shirt", "shoes", "smile", "smoke", "sound",
		"space", "spoon", "storm", "table", "teeth", "tiger", "toast", "touch", "train", "truck",
		"voice", "water", "watch", "whale", "wheel", "world", "write", "youth", "zebra", "actor",
		"alarm", "angle", "atlas", "badge", "beast", "beret", "blade", "blast", "blend", "blink",
		"block", "bloom", "board", "boost", "brace", "brave", "brick", "bride", "brief", "brisk",
		"broad", "broom", "brown", "cable", "camel", "canal", "candy", "cargo", "carve", "cause",
		"chain", "chalk", "charm", "chart", "check", "cheek", "chief", "child", "chime", "china",
		"choir", "chunk", "cider", "cigar", "circus", "clasp", "clean", "clear", "climb", "cloak",
		"close", "cloth", "clown", "coach", "coast", "cobra", "color", "comet", "coral", "couch",
		"count", "court", "cover", "crack", "craft", "crane", "crate", "crawl", "crazy", "cream",
		"crest", "cricket", "crime", "crisp", "crook", "cross", "crowd", "crown", "crust", "cube",
		"curve", "cycle", "daily", "dairy", "daisy", "dance", "decor", "delta", "dense", "depth",
		"derby", "desert", "design", "detail", "device", "devil", "digit", "dime", "dinner", "disco",
		"dish", "diver", "dock", "dodge", "dollar", "dolphin", "donor", "donut", "door", "dose",
		"dot", "double", "doubt", "drain", "drama", "dream", "dress", "drift", "drill", "drive",
		"drone", "drop", "drum", "duck", "duel", "duet", "duke", "duly", "dust", "duty", "eagle",
		"early", "earn", "earth", "ease", "east", "echo", "edge", "edit", "egg", "eight", "elder",
		"elect", "elite", "else", "ember", "empty", "entry", "envoy", "equal", "error", "essay",
	}

	var chosen []string
	for i := 0; i < numWords; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(wordlist))))
		chosen = append(chosen, wordlist[num.Int64()])
	}

	passphrase := strings.Join(chosen, "-")

	fmt.Printf("\n%s\n%s\n\n", theme.StyleSecondary("Generated Passphrase:"), theme.StyleSuccess(passphrase))
	return nil
}
