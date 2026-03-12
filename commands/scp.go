package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/trntv/sshed/ui"
	"github.com/urfave/cli"
)

func (cmds *Commands) newSCPCommand() cli.Command {
	return cli.Command{
		Name:      "scp",
		Usage:     "Transfer files using Aura host configuration",
		ArgsUsage: "[source] [dest]",
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				scpArgs, srv, msg, err := ui.ShowFileCopyPicker()
				if err != nil {
					return err
				}

				if msg != "" {
					fmt.Println(msg)
				}

				cmd := exec.Command("scp", scpArgs...)
				// Use RunCommand if we have a primary host for authentication, 
				// otherwise run directly (e.g. for local-to-local if ever supported)
				if srv != nil {
					return cmds.RunCommand(cmd, srv, os.Stdout, os.Stderr)
				}
				return cmd.Run()
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
				return fmt.Errorf("host not found: %s", hostKey)
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
				idFile = strings.TrimSuffix(idFile, ".pub")
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
