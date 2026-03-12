package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/urfave/cli"
)

func (cmds *Commands) newSCPCommand() cli.Command {
	return cli.Command{
		Name:      "scp",
		Usage:     "Transfer files using Aura host configuration",
		ArgsUsage: "[source] [dest]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				fmt.Println(theme.StyleHeader(" AURA SCP USAGE "))
				fmt.Println("Transfer files using Aura's host configuration and Touch ID security.")
				fmt.Println("\n" + theme.StyleSecondary("Examples:"))
				fmt.Println("  ./aura scp local-file.txt do:/root/          " + theme.StyleTag("Upload file"))
				fmt.Println("  ./aura scp do:/root/backup.tar.gz ./         " + theme.StyleTag("Download file"))
				fmt.Println("  ./aura scp -r ./folder do:/var/www/          " + theme.StyleTag("Upload directory"))
				return nil
			}

			src := c.Args().Get(0)
			dst := c.Args().Get(1)

			// Detect which part is the remote host
			// Pattern: host:path or path host:path
			
			var hostKey, remotePath string
			var isUpload bool

			if strings.Contains(dst, ":") {
				parts := strings.SplitN(dst, ":", 2)
				hostKey = parts[0]
				remotePath = parts[1]
				isUpload = true
			} else if strings.Contains(src, ":") {
				parts := strings.SplitN(src, ":", 2)
				hostKey = parts[0]
				remotePath = parts[1]
				isUpload = false
			} else {
				return fmt.Errorf("one of the arguments must be in [host]:[path] format")
			}

			srv := ssh.Config.Get(hostKey)
			if srv == nil {
				return fmt.Errorf(theme.StyleError("host not found: " + hostKey))
			}

			// Construct SCP command
			scpArgs := []string{"-P", srv.Port}
			if srv.Port == "" {
				scpArgs[0] = "-P"
				scpArgs[1] = "22"
			}
			if srv.IdentityFile != "" {
				idFile := ssh.ConvertTilde(srv.IdentityFile)
				// If it points to a .pub file, try to use the private key instead
				if strings.HasSuffix(idFile, ".pub") {
					idFile = strings.TrimSuffix(idFile, ".pub")
				}
				scpArgs = append(scpArgs, "-i", idFile)
			}

			remoteAddr := srv.Hostname
			if srv.User != "" {
				remoteAddr = srv.User + "@" + srv.Hostname
			}

			if isUpload {
				scpArgs = append(scpArgs, src, remoteAddr+":"+remotePath)
			} else {
				scpArgs = append(scpArgs, remoteAddr+":"+remotePath, dst)
			}

			fmt.Printf("%s %s\n", theme.StyleSecondary("Executing SCP:"), theme.StylePrimary("scp "+strings.Join(scpArgs, " ")))

			// We need to use the same PTY injection for SCP if password is required
			cmd := exec.Command("scp", scpArgs...)

			// Note: We reuse RunCommand because it handles PTY and password injection
			// RunCommand expects *exec.Cmd and *host.Host
			return cmds.RunCommand(cmd, srv, os.Stdout, os.Stderr)
		},
	}
}
