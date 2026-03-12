package ui

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/trntv/sshed/health"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct {
	key            string
	hostname       string
	healthCheckURL string
	status         health.Status
	httpStatus     health.Status
}

func (i item) FilterValue() string { return i.key }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.key)

	statusDot := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("●")
	if i.status == health.StatusUp {
		statusDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("●")
	} else if i.status == health.StatusDown {
		statusDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("●")
	}

	httpIndicator := ""
	if i.healthCheckURL != "" {
		httpIcon := "🌐"
		httpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		if i.httpStatus == health.StatusUp {
			httpStyle = httpStyle.Foreground(lipgloss.Color("#50FA7B"))
		} else if i.httpStatus == health.StatusDown {
			httpStyle = httpStyle.Foreground(lipgloss.Color("#FF5555"))
		}
		httpIndicator = " " + httpStyle.Render(httpIcon)
	}

	var fn func(...string) string
	fn = itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render(s...)
		}
		str = "➜ " + str
	}

	fmt.Fprintf(w, "%s%s %s", statusDot, httpIndicator, fn(str))
}

type Model struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type tickMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "f":
			// Switch to tunnels (handled in ShowDashboard loop or caller)
			m.choice = "SWITCH_TO_TUNNELS"
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i.key
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := msg.Width, msg.Height
		m.list.SetSize(h, v-5)
	case tickMsg:
		// Refresh health
		items := m.list.Items()
		for idx, itm := range items {
			i := itm.(item)
			srv := ssh.Config.Get(i.key)
			if srv != nil {
				i.status = health.Check(srv.Hostname, srv.Port)
				if i.healthCheckURL != "" {
					i.httpStatus = health.CheckHTTP(i.healthCheckURL)
				}
				items[idx] = i
			}
		}
		m.list.SetItems(items)
		return m, tick()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.choice != "" {
		return ""
	}
	if m.quitting {
		return quitTextStyle.Render("Hasta luego!")
	}
	return "\n" + m.list.View()
}

func ShowDashboard() (string, error) {
	hosts := ssh.Config.GetAll()
	var items []list.Item
	keys := make([]string, 0, len(hosts))
	for k := range hosts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		srv := hosts[k]
		items = append(items, item{
			key:            k,
			hostname:       srv.Hostname,
			healthCheckURL: srv.HealthCheckURL,
			status:         health.StatusDown,
			httpStatus:     health.StatusDown,
		})
	}

	theme.Init(theme.ActiveTheme.Name) // Ensure theme is loaded

	l := list.New(items, itemDelegate{}, 0, 0)
	l.Title = " AURA DASHBOARD "
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Background(theme.ActiveTheme.Primary).
		Foreground(theme.ActiveTheme.Foreground).
		Padding(0, 1)

	m := Model{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())
	res, err := p.Run()
	if err != nil {
		return "", err
	}

	finalModel := res.(Model)
	return finalModel.choice, nil
}
