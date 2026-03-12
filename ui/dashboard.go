package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/trntv/sshed/health"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
)

var (
	quitTextStyle = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	detailStyle   = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))
)

type DashboardModel struct {
	list        list.Model
	input       textinput.Model
	showInput   bool
	choice      string
	quitting    bool
	pingResult  string
	diagnostic  health.Diagnostic
	diagLoading bool
	diagError   string
	cmdOutput   string
	cmdLoading  bool
	sortByMRU   bool
}

func (m DashboardModel) Init() tea.Cmd {
	_ = ssh.LoadState()
	ti := textinput.New()
	ti.Placeholder = "Enter remote command..."
	ti.CharLimit = 156
	ti.Width = 30
	m.input = ti
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}
type diagMsg struct {
	diag health.Diagnostic
	err  error
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				cmd := m.input.Value()
				if cmd != "" {
					m.showInput = false
					m.cmdLoading = true
					m.cmdOutput = ""
					i := m.list.SelectedItem().(UnifiedHostItem)
					return m, runAdhocCmd(i.Key, cmd)
				}
			case "esc":
				m.showInput = false
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "p":
			i, ok := m.list.SelectedItem().(UnifiedHostItem)
			if ok {
				return m, pingCmd(i.Hostname)
			}
		case "d": // Diagnostic
			i, ok := m.list.SelectedItem().(UnifiedHostItem)
			if ok && !i.IsLocal {
				m.diagLoading = true
				m.diagError = ""
				return m, diagnosticCmd(i.Key)
			}
		case "x": // Ad-hoc command
			i, ok := m.list.SelectedItem().(UnifiedHostItem)
			if ok && !i.IsLocal {
				m.showInput = true
				m.input.Focus()
				m.input.SetValue("")
				return m, nil
			}
		case "f":
			// Toggle favorite
			i, ok := m.list.SelectedItem().(UnifiedHostItem)
			if ok && !i.IsLocal {
				ssh.ToggleFavorite(i.Key)
				return m, refreshListCmd(m.sortByMRU)
			}
		case "H":
			// Toggle MRU sorting
			m.sortByMRU = !m.sortByMRU
			return m, refreshListCmd(m.sortByMRU)
		case "enter":
			i, ok := m.list.SelectedItem().(UnifiedHostItem)
			if ok {
				m.choice = i.Key
				ssh.RecordConnection(i.Key)
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width/2, msg.Height-5)
	case tickMsg:
		return m, tea.Batch(tick(), checkHealthCmd(m.list.Items()))
	case healthResultsMsg:
		items := m.list.Items()
		for _, res := range msg {
			for idx, itm := range items {
				i := itm.(UnifiedHostItem)
				if i.Key == res.Key {
					i.Status = res.Status
					items[idx] = i
				}
			}
		}
		m.list.SetItems(items)
		return m, nil
	case diagMsg:
		m.diagLoading = false
		if msg.err != nil {
			m.diagError = msg.err.Error()
		} else {
			m.cmdOutput = "" // Clear ad-hoc if diagnostic loaded
			m.diagnostic = msg.diag
		}
		return m, nil
	case adhocMsg:
		m.cmdLoading = false
		m.diagnostic = health.Diagnostic{} // Clear diagnostic if ad-hoc loaded
		if msg.err != nil {
			m.cmdOutput = "Error: " + msg.err.Error()
		} else {
			m.cmdOutput = msg.out
		}
		return m, nil
	case pingMsg:
		m.pingResult = string(msg)
		return m, nil
	case refreshListMsg:
		m.list.SetItems(msg)
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DashboardModel) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return quitTextStyle.Render("Hasta luego!")
	}

	// Main list
	listView := m.list.View()

	// Detail Pane
	var detailContent string
	selected := m.list.SelectedItem()
	if selected != nil {
		i := selected.(UnifiedHostItem)
		detailContent = fmt.Sprintf("%s\n\n", theme.StyleHeader(" SERVER DETAILS "))
		detailContent += fmt.Sprintf("%s %s\n", theme.StyleSecondary("Key:     "), i.Key)
		detailContent += fmt.Sprintf("%s %s\n", theme.StyleSecondary("Host:    "), i.Hostname)
		if len(i.Tags) > 0 {
			detailContent += fmt.Sprintf("%s %s\n", theme.StyleSecondary("Tags:    "), strings.Join(i.Tags, ", "))
		}
		
		if !i.IsLocal {
			// Commands Help
			detailContent += "\n" + theme.StylePrimary("(p)") + " Ping " + 
			                 theme.StylePrimary("(d)") + " Diag " + 
			                 theme.StylePrimary("(x)") + " Cmd " + 
			                 theme.StylePrimary("(f)") + " Fav\n"

			detailContent += "\n" + theme.StyleHeader(" OUTPUT ") + "\n"
			
			if m.showInput {
				detailContent += "\n" + m.input.View() + "\n" + theme.StyleInfo("press Enter to run, Esc to cancel")
			} else if m.diagLoading || m.cmdLoading {
				detailContent += "\n⏳ Running..."
			} else if m.diagError != "" {
				detailContent += "\n❌ " + theme.StyleError(m.diagError)
			} else if m.diagnostic.Uptime != "" {
				detailContent += fmt.Sprintf("\n%s %s\n", theme.StyleSecondary("Uptime:  "), m.diagnostic.Uptime)
				detailContent += fmt.Sprintf("%s %s\n", theme.StyleSecondary("Load:    "), m.diagnostic.LoadAvg)
				detailContent += fmt.Sprintf("%s %s / %s\n", theme.StyleSecondary("Memory:  "), m.diagnostic.MemUsed, m.diagnostic.MemTotal)
				detailContent += fmt.Sprintf("%s %s / %s\n", theme.StyleSecondary("Disk /:  "), m.diagnostic.DiskUsed, m.diagnostic.DiskTotal)
			} else if m.cmdOutput != "" {
				detailContent += "\n" + theme.StyleInfo(m.cmdOutput)
			} else {
				detailContent += "\nSelect a server and press a key to see details."
			}
		} else {
			detailContent += "\n💻 This is your local machine."
		}
	}

	detailView := detailStyle.Width(m.list.Width()).Height(m.list.Height()).Render(detailContent)

	res := lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView)

	if m.pingResult != "" {
		res += "\n\n" + theme.StylePrimary(" Ping: ") + m.pingResult
	}

	return "\n" + res
}

// Commands

type healthResultsMsg []health.CheckResult

func checkHealthCmd(items []list.Item) tea.Cmd {
	return func() tea.Msg {
		var hosts []struct{ Key, Hostname, Port string }
		for _, itm := range items {
			i := itm.(UnifiedHostItem)
			if !i.IsLocal {
				srv := ssh.Config.Get(i.Key)
				if srv != nil {
					hosts = append(hosts, struct{ Key, Hostname, Port string }{i.Key, srv.Hostname, srv.Port})
				}
			}
		}
		results := make([]health.CheckResult, 0)
		out := health.CheckConcurrent(hosts)
		for res := range out {
			results = append(results, res)
		}
		return healthResultsMsg(results)
	}
}

func diagnosticCmd(key string) tea.Cmd {
	return func() tea.Msg {
		srv := ssh.Config.Get(key)
		if srv == nil {
			return diagMsg{err: fmt.Errorf("host not found")}
		}
		
		// Run diagnostic command
		out, err := ssh.RunRemoteCommand(srv, "uptime; free -m; df -h /")
		if err != nil {
			return diagMsg{err: err}
		}

		// Split output
		parts := strings.Split(out, "\n")
		// This is a bit brittle, a better way would be to run them separately or use markers
		// For now assume uptime is first line, free has "Mem:", df has "/"
		var uptime, free, df string
		for _, line := range parts {
			if strings.Contains(line, "load average:") {
				uptime = line
			}
			if strings.Contains(line, "Mem:") {
				free = line
			}
			if strings.Contains(line, " /") && strings.Contains(line, "%") {
				df = line
			}
		}
		
		return diagMsg{diag: health.ParseDiagnostic(uptime, free, df)}
	}
}

type refreshListMsg []list.Item

func refreshListCmd(sortByMRU bool) tea.Cmd {
	return func() tea.Msg {
		return refreshListMsg(GetSearchableHosts(false, sortByMRU))
	}
}

type adhocMsg struct {
	out string
	err error
}

func runAdhocCmd(key string, command string) tea.Cmd {
	return func() tea.Msg {
		srv := ssh.Config.Get(key)
		if srv == nil {
			return adhocMsg{err: fmt.Errorf("host not found")}
		}
		out, err := ssh.RunRemoteCommand(srv, command)
		return adhocMsg{out: out, err: err}
	}
}

type pingMsg string

func pingCmd(hostname string) tea.Cmd {
	return func() tea.Msg {
		duration, err := health.Ping(hostname)
		if err != nil {
			return pingMsg(fmt.Sprintf("Error: %v", err))
		}
		return pingMsg(fmt.Sprintf("%v", duration))
	}
}

func ShowDashboard() (string, error) {
	_ = ssh.LoadState()
	items := GetSearchableHosts(false, false)

	theme.Init(theme.ActiveTheme.Name)

	l := list.New(items, UnifiedItemDelegate{}, 0, 0)
	l.Title = " AURA PREMIUM DASHBOARD "
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Background(theme.ActiveTheme.Primary).
		Foreground(theme.ActiveTheme.Foreground).
		Padding(0, 1)

	m := DashboardModel{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
	res, err := p.Run()
	if err != nil {
		return "", err
	}

	finalModel := res.(DashboardModel)
	return finalModel.choice, nil
}
