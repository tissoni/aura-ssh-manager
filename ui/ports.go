package ui

import (
	"fmt"
	"io"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/trntv/sshed/theme"
	"github.com/trntv/sshed/utils"
)

type portItem struct {
	proc utils.Process
}

func (i portItem) FilterValue() string { return i.proc.Port + " " + i.proc.Name }

type portDelegate struct{}

func (d portDelegate) Height() int                               { return 1 }
func (d portDelegate) Spacing() int                              { return 0 }
func (d portDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d portDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(portItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("Port: %-6s | PID: %-6s | Process: %s", i.proc.Port, i.proc.PID, i.proc.Name)

	var fn func(...string) string
	fn = itemStyle.Render
	if index == m.Index() {
		fn = selectedItemStyle.Render
		str = "➜ " + str
	}

	fmt.Fprintf(w, "%s", fn(str))
}

type PortsModel struct {
	list     list.Model
	quitting bool
	err      error
}

func (m PortsModel) Init() tea.Cmd {
	return nil
}

func (m PortsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "k", "backspace", "delete":
			i, ok := m.list.SelectedItem().(portItem)
			if ok {
				err := utils.KillProcess(i.proc.PID)
				if err != nil {
					m.err = err
					return m, nil
				}
				// Refresh list
				return m, refreshPorts()
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-4)
	case refreshPortsMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		var items []list.Item
		for _, p := range msg.procs {
			items = append(items, portItem{p})
		}
		m.list.SetItems(items)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type refreshPortsMsg struct {
	procs []utils.Process
	err   error
}

func refreshPorts() tea.Cmd {
	return func() tea.Msg {
		procs, err := utils.ListActivePorts()
		return refreshPortsMsg{procs: procs, err: err}
	}
}

func (m PortsModel) View() string {
	if m.err != nil {
		return theme.StyleError(fmt.Sprintf("\n Error: %v", m.err))
	}
	return "\n" + m.list.View() + "\n (k) Kill process | (q) Quit"
}

func ShowPortKiller() error {
	procs, err := utils.ListActivePorts()
	if err != nil {
		return err
	}

	var items []list.Item
	for _, p := range procs {
		items = append(items, portItem{p})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].(portItem).proc.Port < items[j].(portItem).proc.Port
	})

	l := list.New(items, portDelegate{}, 0, 0)
	l.Title = " AURA PORT KILLER (macOS) "
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.Color("#FF5555")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1)

	m := PortsModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
