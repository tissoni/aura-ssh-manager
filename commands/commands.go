package commands

import (
	"fmt"
	"github.com/creack/pty"
	"github.com/trntv/sshed/host"
	"github.com/trntv/sshed/keychain"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/trntv/sshed/tunnels"
	"github.com/trntv/sshed/ui"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Commands struct {
	bin string
}

type options struct {
	verbose bool
}

func RegisterCommands(app *cli.App) {
	commands := &Commands{}

	beforeFunc := app.Before
	app.Before = func(context *cli.Context) error {

		err := beforeFunc(context)
		if err != nil {
			return err
		}

		commands.bin = context.String("bin")

		return nil
	}

	app.Commands = []cli.Command{
		commands.newShowCommand(),
		commands.newListCommand(),
		commands.newAddCommand(),
		commands.newRemoveCommand(),
		commands.newToCommand(),
		commands.newAtCommand(),
		commands.newConfigCommand(),
		commands.newThemeCommand(),
		commands.newIdentitiesCommand(),
		commands.newSnippetsCommand(),
		commands.newDashboardCommand(),
		commands.newTunnelsCommand(),
		commands.newBackupCommand(),
		commands.newRestoreCommand(),
		commands.newSCPCommand(),
		commands.newUtilsCommand(),
	}
}

func (cmds *Commands) completeWithServers() {
	hosts := ssh.Config.GetAll()
	for key := range hosts {
		fmt.Println(key)
	}
}

func (cmds *Commands) askPassword() string {
	key := ""
	prompt := &survey.Password{
		Message: "Please type your password:",
	}
	survey.AskOne(prompt, &key, nil)

	return key
}

func (cmds *Commands) askServerKey() (string, error) {
	var key string
	options := make([]string, 0)
	srvs := ssh.Config.GetAll()
	for key := range srvs {
		options = append(options, key)
	}

	sort.Strings(options)
	prompt := &survey.Select{
		Message:  "Choose server:",
		Options:  options,
		PageSize: 16,
	}
	err := survey.AskOne(prompt, &key, survey.Required)

	return key, err
}

func (cmds *Commands) askServersKeys() ([]string, error) {
	var keys []string
	options := make([]string, 0)
	srvs := ssh.Config.GetAll()
	for _, h := range srvs {
		options = append(options, h.Key)
	}

	sort.Strings(options)
	prompt := &survey.MultiSelect{
		Message:  "Choose servers:",
		Options:  options,
		PageSize: 16,
	}
	err := survey.AskOne(prompt, &keys, survey.Required)

	return keys, err
}

func (cmds *Commands) createCommand(c *cli.Context, srv *host.Host, options *options, command string) (cmd *exec.Cmd, err error) {
	var username string
	if srv.User == "" {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		username = u.Username
	} else {
		username = srv.User
	}

	var args = make([]string, 0)
	args = append(args, cmds.bin)
	args = append(args, fmt.Sprintf("-F %s", ssh.Config.Path))

	// Intelligent Keep-Alive
	args = append(args, "-o ServerAliveInterval=60")
	args = append(args, "-o ServerAliveCountMax=3")

	if pk := srv.PrivateKey(); pk != "" {
		tf, err := ioutil.TempFile("", "")
		defer os.Remove(tf.Name())
		defer tf.Close()

		if err != nil {
			return nil, err
		}

		_, err = tf.Write([]byte(pk))
		if err != nil {
			return nil, err
		}

		err = tf.Chmod(os.FileMode(0600))
		if err != nil {
			return nil, err
		}

		srv.IdentityFile = tf.Name()
	}

	if srv.User != "" {
		args = append(args, fmt.Sprintf("%s@%s", username, srv.Hostname))
	} else {
		args = append(args, fmt.Sprintf("%s", srv.Hostname))
	}

	if srv.Port != "" {
		args = append(args, fmt.Sprintf("-p %s", srv.Port))
	}

	if srv.IdentityFile != "" {
		idFile := ssh.ConvertTilde(srv.IdentityFile)
		// If it points to a .pub file, try to use the private key instead
		if strings.HasSuffix(idFile, ".pub") {
			idFile = strings.TrimSuffix(idFile, ".pub")
		}
		args = append(args, fmt.Sprintf("-i %s", idFile))
	}

	if options.verbose == true {
		args = append(args, "-v")
	}

	if command != "" {
		args = append(args, command)
	}

	if options.verbose == true {
		fmt.Printf("%s: %s\n", theme.StyleSuccess("Executing"), theme.StyleSecondary(strings.Join(args, " ")))
	}

	cmd = exec.Command("sh", "-c", strings.Join(args, " "))

	return cmd, err
}

func (cmds *Commands) RunCommand(cmd *exec.Cmd, srv *host.Host, stdout io.Writer, stderr io.Writer) error {
	// Prepare Logging
	usr, _ := user.Current()
	logDir := filepath.Join(usr.HomeDir, "Aura", "logs")
	_ = os.MkdirAll(logDir, 0755)
	logFile, _ := os.Create(filepath.Join(logDir, fmt.Sprintf("%s_%s.log", srv.Key, time.Now().Format("2006-01-02_15-04-05"))))
	if logFile != nil {
		defer logFile.Close()
		multiStdout := io.MultiWriter(stdout, logFile)
		stdout = multiStdout
	}

	password := srv.Password()
	if password == "" {
		cmd.Stderr = stderr
		cmd.Stdin = os.Stdin
		cmd.Stdout = stdout
		return cmd.Run()
	}

	// Touch ID Verification
	fmt.Printf("%s\n", theme.StyleSecondary("Authenticating via Touch ID..."))
	if !keychain.VerifyTouchID() {
		return fmt.Errorf(theme.StyleError("Touch ID authentication failed"))
	}

	f, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer f.Close()

	// Copy output to stdout and handle robust password injection
	go func() {
		var buf [1024]byte
		injected := false
		for {
			n, err := f.Read(buf[:])
			if n > 0 {
				chunk := buf[:n]
				_, _ = stdout.Write(chunk)

				// Detect password prompt and inject
				if !injected && strings.Contains(strings.ToLower(string(chunk)), "password:") {
					// Tiny delay to ensure the PTY is ready to receive
					time.Sleep(10 * time.Millisecond)
					_, _ = fmt.Fprintf(f, "%s\n", password)
					injected = true
				}

				// Detect known_hosts error
				if strings.Contains(string(chunk), "Host key verification failed") {
					go func() {
						// Small delay to let the error output finish
						time.Sleep(100 * time.Millisecond)
						fmt.Printf("\n%s\n", theme.StyleError("! Aura detected a Host Key Verification failure."))
						var fix bool
						_ = survey.AskOne(&survey.Confirm{
							Message: "Do you want Aura to fix your known_hosts automatically for " + srv.Hostname + "?",
							Default: true,
						}, &fix, nil)

						if fix {
							_ = exec.Command("ssh-keygen", "-R", srv.Hostname).Run()
							fmt.Printf("%s\n", theme.StyleSuccess("✓ known_hosts updated. Please try connecting again."))
						}
					}()
				}
			}
			if err != nil {
				break
			}
		}
	}()

	return cmd.Wait()
}

func (cmds *Commands) newTunnelsCommand() cli.Command {
	return cli.Command{
		Name:  "tunnels",
		Usage: "Manager SSH Port Forwarding Tunnels",
		Action: func(c *cli.Context) error {
			mgr := tunnels.NewManager()
			err := mgr.Load()
			if err != nil {
				return err
			}
			return ui.ShowTunnels(mgr)
		},
		Subcommands: []cli.Command{
			{
				Name:   "add",
				Usage:  "Add a new port forwarding tunnel",
				Action: cmds.addTunnelAction,
			},
		},
	}
}

func (cmds *Commands) addTunnelAction(c *cli.Context) error {
	mgr := tunnels.NewManager()
	_ = mgr.Load()

	var qs = []*survey.Question{
		{
			Name:     "name",
			Prompt:   &survey.Input{Message: "Tunnel Name:"},
			Validate: survey.Required,
		},
		{
			Name:     "local",
			Prompt:   &survey.Input{Message: "Local Address (e.g. 127.0.0.1:8080):"},
			Validate: survey.Required,
		},
		{
			Name:     "remote",
			Prompt:   &survey.Input{Message: "Remote Address (e.g. 127.0.0.1:80):"},
			Validate: survey.Required,
		},
		{
			Name: "type",
			Prompt: &survey.Select{
				Message: "Tunnel Type:",
				Options: []string{"local", "remote", "dynamic"},
				Default: "local",
			},
		},
	}

	answers := struct {
		Name   string
		Local  string
		Remote string
		Type   string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		return err
	}

	hostKey, err := cmds.askServerKey()
	if err != nil {
		return err
	}

	id := strings.ToLower(strings.ReplaceAll(answers.Name, " ", "-"))
	mgr.Add(&tunnels.Tunnel{
		ID:            id,
		Name:          answers.Name,
		HostKey:       hostKey,
		LocalAddress:  answers.Local,
		RemoteAddress: answers.Remote,
		Type:          tunnels.TunnelType(answers.Type),
	})

	err = mgr.Save()
	if err != nil {
		return err
	}

	fmt.Printf("\n%s %s\n", theme.StyleSuccess("✓"), "Tunnel added successfully!")
	return nil
}
