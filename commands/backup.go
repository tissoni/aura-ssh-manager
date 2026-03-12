package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/trntv/sshed/config"
	"github.com/trntv/sshed/crypto"
	"github.com/trntv/sshed/keychain"
	"github.com/trntv/sshed/snippets"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
)

type BackupPackage struct {
	Config   *config.Config            `json:"config"`
	Snippets *snippets.Snippets        `json:"snippets"`
	Keychain map[string]*keychain.Record `json:"keychain"`
}

func (cmds *Commands) newBackupCommand() cli.Command {
	return cli.Command{
		Name:  "backup",
		Usage: "Export all Aura data to an encrypted file",
		Action: func(c *cli.Context) error {
			dest := c.Args().First()
			if dest == "" {
				dest = "aura_backup.bin"
			}

			pwd := ""
			err := survey.AskOne(&survey.Password{
				Message: "Enter a master password to encrypt the backup:",
			}, &pwd, survey.Required)
			if err != nil {
				return err
			}

			// Gather data
			conf := config.Load()
			pkg := &BackupPackage{
				Config:   &conf,
				Snippets: snippets.Load(),
				Keychain: make(map[string]*keychain.Record),
			}

			// Keychain records are per-host
			hosts := ssh.Config.GetAll()
			for k := range hosts {
				rec, err := keychain.Get(k)
				if err == nil && rec != nil {
					pkg.Keychain[k] = rec
				}
			}

			data, err := json.Marshal(pkg)
			if err != nil {
				return err
			}

			encrypted, err := crypto.Encrypt(data, pwd)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(dest, encrypted, 0600)
			if err != nil {
				return err
			}

			fmt.Printf("%s\n", theme.StyleSuccess("✓ Backup created successfully at "+dest))
			return nil
		},
	}
}

func (cmds *Commands) newRestoreCommand() cli.Command {
	return cli.Command{
		Name:  "restore",
		Usage: "Import Aura data from an encrypted backup file",
		Action: func(c *cli.Context) error {
			src := c.Args().First()
			if src == "" {
				return fmt.Errorf("please specify a backup file to restore from")
			}

			pwd := ""
			err := survey.AskOne(&survey.Password{
				Message: "Enter the master password used for this backup:",
			}, &pwd, survey.Required)
			if err != nil {
				return err
			}

			encrypted, err := ioutil.ReadFile(src)
			if err != nil {
				return err
			}

			decrypted, err := crypto.Decrypt(encrypted, pwd)
			if err != nil {
				return fmt.Errorf("decryption failed: %v", err)
			}

			pkg := &BackupPackage{}
			err = json.Unmarshal(decrypted, pkg)
			if err != nil {
				return err
			}

			// Confirmation
			var confirm bool
			_ = survey.AskOne(&survey.Confirm{
				Message: "This will overwrite your current config, snippets, and keychain records. Are you sure?",
				Default: false,
			}, &confirm, nil)

			if !confirm {
				return nil
			}

			// Restore Config
			if pkg.Config != nil {
				config.Save(*pkg.Config)
			}

			// Restore Snippets
			if pkg.Snippets != nil {
				pkg.Snippets.Save()
			}

			// Restore Keychain
			for k, rec := range pkg.Keychain {
				keychain.Put(k, rec)
			}

			fmt.Printf("%s\n", theme.StyleSuccess("✓ Restore completed successfully. Please restart Aura."))
			return nil
		},
	}
}
