package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/trntv/sshed/host"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"gopkg.in/AlecAivazis/survey.v1"
)

// --- Copy Model (File Picker) ---

type CopyModel struct {
	filepicker   filepicker.Model
	selectedFile string
	quitting     bool
	err          error
}

func (m CopyModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m CopyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
		return m, tea.Quit
	}

	return m, cmd
}

func (m CopyModel) View() string {
	if m.quitting {
		return ""
	}
	if m.selectedFile != "" {
		return theme.StyleSuccess("Selected: ") + m.selectedFile + "\n"
	}
	return "\n" + theme.StyleHeader(" PICK A FILE OR DIRECTORY ") + "\n\n" + m.filepicker.View()
}

// --- Host Selection ---

func askHostUnified(message string) (string, *host.Host, error) {
	return SearchHosts(message, true)
}

// --- Main Picker ---

func askPath(isLocal bool, message string, isDirOnly bool) (string, error) {
	if isLocal {
		fp := filepicker.New()
		fp.AllowedTypes = []string{}
		fp.DirAllowed = true
		if isDirOnly {
			fp.FileAllowed = false
		}
		fp.CurrentDirectory, _ = os.Getwd()

		m := CopyModel{filepicker: fp}
		p := tea.NewProgram(m, tea.WithAltScreen())

		res, err := p.Run()
		if err != nil {
			return "", err
		}

		finalModel := res.(CopyModel)
		if finalModel.quitting || finalModel.selectedFile == "" {
			return "", fmt.Errorf("cancelled")
		}
		return finalModel.selectedFile, nil
	}

	var path string
	err := survey.AskOne(&survey.Input{
		Message: message,
		Default: "~/",
	}, &path, survey.Required)
	return path, err
}

func ShowFileCopyPicker() ([]string, *host.Host, string, error) {
	srcKey, srcHost, err := SearchHosts("SELECT SOURCE HOST", true)
	if err != nil {
		return nil, nil, "", err
	}
	fmt.Printf("%s Source Host: %s\n", theme.StyleSuccess("✓"), theme.StyleSecondary(srcKey))

	dstKey, dstHost, err := SearchHosts("SELECT DESTINATION HOST", true)
	if err != nil {
		return nil, nil, "", err
	}
	fmt.Printf("%s Destination Host: %s\n", theme.StyleSuccess("✓"), theme.StyleSecondary(dstKey))

	if srcKey == "LOCAL" && dstKey == "LOCAL" {
		return nil, nil, "", fmt.Errorf("source and destination cannot both be local (just use cp!)")
	}

	srcPath, err := askPath(srcKey == "LOCAL", "Source path on "+srcKey+":", false)
	if err != nil {
		return nil, nil, "", err
	}
	fmt.Printf("%s Source: %s\n", theme.StyleSuccess("✓"), theme.StyleSecondary(srcPath))

	var isRecursive bool
	if srcKey == "LOCAL" {
		info, err := os.Stat(srcPath)
		if err == nil && info.IsDir() {
			isRecursive = true
		}
	} else {
		survey.AskOne(&survey.Confirm{
			Message: "Is the source a directory? (Copy recursively)",
			Default: false,
		}, &isRecursive, nil)
	}

	dstPath, err := askPath(dstKey == "LOCAL", "Destination path on "+dstKey+":", dstKey == "LOCAL")
	if err != nil {
		return nil, nil, "", err
	}
	fmt.Printf("%s Destination: %s\n", theme.StyleSuccess("✓"), theme.StyleSecondary(dstPath))

	// Construct SCP command arguments
	var args []string

	args = append(args, "-O") // Legacy protocol for better compatibility
	if isRecursive {
		args = append(args, "-r")
	}

	if srcKey != "LOCAL" && dstKey != "LOCAL" {
		args = append(args, "-3")
	}

	addSCPPeerArgs := func(h *host.Host) {
		if h.Port != "" && h.Port != "22" {
			args = append(args, "-P", h.Port)
		}
		if h.IdentityFile != "" {
			idFile := strings.TrimSuffix(ssh.ConvertTilde(h.IdentityFile), ".pub")
			args = append(args, "-i", idFile)
		}
	}

	var primaryHost *host.Host
	if srcHost != nil {
		primaryHost = srcHost
		addSCPPeerArgs(srcHost)
	}
	if dstHost != nil {
		if primaryHost == nil {
			primaryHost = dstHost
		}
		addSCPPeerArgs(dstHost)
	}

	formatRemote := func(h *host.Host, path string) string {
		addr := h.Hostname
		if h.User != "" {
			addr = h.User + "@" + addr
		}
		return addr + ":" + path
	}

	var finalSrc, finalDst string
	if srcKey == "LOCAL" {
		finalSrc = srcPath
	} else {
		finalSrc = formatRemote(srcHost, srcPath)
	}

	if dstKey == "LOCAL" {
		finalDst = dstPath
	} else {
		finalDst = formatRemote(dstHost, dstPath)
	}

	args = append(args, finalSrc, finalDst)

	msg := fmt.Sprintf("\n%s Transferring %s ➜ %s...", theme.StyleSecondary("➜"), srcKey, dstKey)
	
	return args, primaryHost, msg, nil
}
