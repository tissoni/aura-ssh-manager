package ui

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	sshConfig "github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
	"github.com/trntv/sshed/tunnels"
	"golang.org/x/crypto/ssh"
)

var (
	mgr *tunnels.Manager
)

type tunnelItem struct {
	tunnel *tunnels.Tunnel
}

func (i tunnelItem) FilterValue() string { return i.tunnel.Name }

type tunnelDelegate struct{}

func (d tunnelDelegate) Height() int                               { return 1 }
func (d tunnelDelegate) Spacing() int                              { return 0 }
func (d tunnelDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d tunnelDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(tunnelItem)
	if !ok {
		return
	}

	state := "○"
	stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	if i.tunnel.Active {
		state = "●"
		stateStyle = stateStyle.Foreground(lipgloss.Color("#50FA7B"))
	}

	name := i.tunnel.Name
	info := fmt.Sprintf("[%s] %s -> %s", i.tunnel.Type, i.tunnel.LocalAddress, i.tunnel.RemoteAddress)

	str := fmt.Sprintf("%-15s %s", name, info)

	var fn func(...string) string
	fn = itemStyle.Render
	if index == m.Index() {
		fn = selectedItemStyle.Render
		str = "➜ " + str
	}

	fmt.Fprintf(w, "%s %s", stateStyle.Render(state), fn(str))
}

type TunnelModel struct {
	list     list.Model
	quitting bool
}

func (m TunnelModel) Init() tea.Cmd {
	return nil
}

func (m TunnelModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "s":
			i, ok := m.list.SelectedItem().(tunnelItem)
			if ok && !i.tunnel.Active {
				// Get Host config
				h := sshConfig.Config.Get(i.tunnel.HostKey)
				if h != nil {
					// Prepare SSH Config (simplified for now, using existing logic)
					auth := []ssh.AuthMethod{}
					if pwd := h.Password(); pwd != "" {
						auth = append(auth, ssh.Password(pwd))
					}
					
					config := &ssh.ClientConfig{
						User:            h.User,
						Auth:            auth,
						HostKeyCallback: ssh.InsecureIgnoreHostKey(),
						Timeout:         5 * time.Second,
					}
					
					addr := h.Hostname
					if h.Port != "" {
						addr = fmt.Sprintf("%s:%s", addr, h.Port)
					} else {
						addr = fmt.Sprintf("%s:22", addr)
					}

					err := mgr.Start(i.tunnel, config, addr)
					if err != nil {
						// Error handling could be better in UI
						return m, nil
					}
				}
			}
		case "x":
			i, ok := m.list.SelectedItem().(tunnelItem)
			if ok && i.tunnel.Active {
				mgr.Stop(i.tunnel.ID)
			}
		case "d":
			i, ok := m.list.SelectedItem().(tunnelItem)
			if ok {
				if i.tunnel.Active {
					mgr.Stop(i.tunnel.ID)
				}
				mgr.Remove(i.tunnel.ID)
				_ = mgr.Save()
				
				// Remove from local list and refresh
				idx := m.list.Index()
				m.list.RemoveItem(idx)
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m TunnelModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View() + "\n (s) Start | (x) Stop | (d) Delete | (q) Quit"
}

func ShowTunnels(mInstance *tunnels.Manager) error {
	mgr = mInstance
	var items []list.Item
	keys := make([]string, 0, len(mgr.Tunnels))
	for k := range mgr.Tunnels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		items = append(items, tunnelItem{mgr.Tunnels[k]})
	}

	l := list.New(items, tunnelDelegate{}, 0, 0)
	l.Title = " AURA PORT FORWARDING MANAGER "
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Background(theme.ActiveTheme.Primary).
		Foreground(theme.ActiveTheme.Foreground).
		Padding(0, 1)

	m := TunnelModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
