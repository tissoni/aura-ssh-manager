package ui

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/trntv/sshed/host"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"gopkg.in/AlecAivazis/survey.v1"
)

func ShowDeployKey(srv *host.Host) error {
	fmt.Println(theme.StyleHeader(" SSH KEY DEPLOYMENT "))
	fmt.Printf("Deploying key to: %s (%s)\n\n", theme.StylePrimary(srv.Key), theme.StyleSecondary(srv.Hostname))

	usr, _ := user.Current()
	defaultKeys := []string{
		filepath.Join(usr.HomeDir, ".ssh", "id_ed25519.pub"),
		filepath.Join(usr.HomeDir, ".ssh", "id_rsa.pub"),
	}

	var options []string
	for _, k := range defaultKeys {
		if _, err := ioutil.ReadFile(k); err == nil {
			options = append(options, fmt.Sprintf("Use default: %s", filepath.Base(k)))
		}
	}

	options = append(options, "Paste custom public key manually", "Generate new keypair and deploy")

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: "Choose deployment method:",
		Options: options,
	}, &choice, nil)

	if err != nil {
		return err
	}

	var pubKey string

	switch {
	case strings.HasPrefix(choice, "Use default"):
		filename := strings.TrimPrefix(choice, "Use default: ")
		path := filepath.Join(usr.HomeDir, ".ssh", filename)
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		pubKey = string(content)

	case choice == "Paste custom public key manually":
		err = survey.AskOne(&survey.Input{
			Message: "Paste your public key:",
		}, &pubKey, survey.Required)
		if err != nil {
			return err
		}

	case choice == "Generate new keypair and deploy":
		var keyName string
		survey.AskOne(&survey.Input{
			Message: "Key name (default: id_aura):",
			Default: "id_aura",
		}, &keyName, nil)

		path := filepath.Join(usr.HomeDir, ".ssh", keyName)
		_, pub, err := ssh.GenerateKeypair(path)
		if err != nil {
			return err
		}
		pubKey = pub
		fmt.Printf("%s Keypair generated at %s\n", theme.StyleSuccess("✓"), path)
	}

	fmt.Printf("%s Deploying to %s...\n", theme.StyleSecondary("➜"), srv.Hostname)
	err = ssh.DeployPublicKey(srv, pubKey)
	if err != nil {
		return fmt.Errorf("deployment failed: %v", err)
	}

	fmt.Printf("\n%s %s\n", theme.StyleSuccess("✓"), "Public key deployed successfully!")
	return nil
}
