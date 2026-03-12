package ui

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/trntv/sshed/health"
	"github.com/trntv/sshed/host"
	"github.com/trntv/sshed/ssh"
	"github.com/trntv/sshed/theme"
)

// Shared styles
var (
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
)

type UnifiedHostItem struct {
	Key            string
	Hostname       string
	HealthCheckURL string
	Tags           []string
	Status         health.Status
	HTTPStatus     health.Status
	LastConnected  time.Time
	IsFavorite     bool
	IsLocal        bool
}

func (i UnifiedHostItem) Title() string       { return i.Key }
func (i UnifiedHostItem) Description() string { return i.Hostname }
func (i UnifiedHostItem) FilterValue() string {
	return i.Key + " " + i.Hostname + " " + strings.Join(i.Tags, " ")
}

type UnifiedItemDelegate struct{}

func (d UnifiedItemDelegate) Height() int                               { return 1 }
func (d UnifiedItemDelegate) Spacing() int                              { return 0 }
func (d UnifiedItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d UnifiedItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(UnifiedHostItem)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Key)
	if i.IsLocal {
		str = "💻 " + i.Key
	}

	// Status Dot
	statusDot := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("●")
	if !i.IsLocal {
		if i.Status == health.StatusUp {
			statusDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("●")
		} else if i.Status == health.StatusDown {
			statusDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("●")
		}
	}

	// Favorite star
	fav := ""
	if i.IsFavorite {
		fav = theme.StyleWarning("⭐ ")
	}

	// Selection rendering
	var fn func(...string) string
	fn = itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render(s...)
		}
		str = "➜ " + str
	}

	fmt.Fprintf(w, "%s %s%s", statusDot, fav, fn(str))
}

type UnifiedHostPickerModel struct {
	List     list.Model
	Choice   string
	Quitting bool
}

func (m UnifiedHostPickerModel) Init() tea.Cmd { return nil }
func (m UnifiedHostPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.List.SelectedItem().(UnifiedHostItem)
			if ok {
				m.Choice = i.Key
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.List.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m UnifiedHostPickerModel) View() string {
	if m.Quitting {
		return ""
	}
	return "\n" + m.List.View()
}

func SearchHosts(message string, includeLocal bool) (string, *host.Host, error) {
	_ = ssh.LoadState()
	items := GetSearchableHosts(includeLocal, false)

	l := list.New(items, UnifiedItemDelegate{}, 20, 10) // Small default height
	l.Title = message
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.StyleTitle()

	m := UnifiedHostPickerModel{List: l}
	p := tea.NewProgram(m) // No AltScreen
	res, err := p.Run()
	if err != nil {
		return "", nil, err
	}

	finalModel := res.(UnifiedHostPickerModel)
	if finalModel.Quitting || finalModel.Choice == "" {
		return "", nil, fmt.Errorf("cancelled")
	}

	if finalModel.Choice == "LOCAL" {
		return "LOCAL", nil, nil
	}

	return finalModel.Choice, ssh.Config.Get(finalModel.Choice), nil
}

func GetSearchableHosts(includeLocal bool, sortByMRU bool) []list.Item {
	hosts := ssh.Config.GetAll()
	var items []UnifiedHostItem
	
	if includeLocal {
		items = append(items, UnifiedHostItem{
			Key:      "LOCAL",
			Hostname: "Local Machine",
			IsLocal:  true,
		})
	}

	for k, srv := range hosts {
		item := UnifiedHostItem{
			Key:            k,
			Hostname:       srv.Hostname,
			HealthCheckURL: srv.HealthCheckURL,
			Tags:           srv.Tags(),
			Status:         health.StatusDown,
		}
		if ssh.CurrentState != nil {
			item.IsFavorite = ssh.CurrentState.Favorites[k]
			item.LastConnected = ssh.CurrentState.LastConnected[k]
		}
		items = append(items, item)
	}

	// Sorting
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsLocal != items[j].IsLocal {
			return items[i].IsLocal
		}
		if items[i].IsFavorite != items[j].IsFavorite {
			return items[i].IsFavorite
		}
		if sortByMRU {
			return items[i].LastConnected.After(items[j].LastConnected)
		}
		return items[i].Key < items[j].Key
	})

	listItems := make([]list.Item, len(items))
	for i, v := range items {
		listItems[i] = v
	}
	return listItems
}
