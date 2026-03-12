package theme

import (
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/AlecAivazis/survey.v1"
)

type Theme struct {
	Name        string
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Background  lipgloss.Color
	Foreground  lipgloss.Color
	Rainbow     []lipgloss.Color
}

var Themes = map[string]Theme{
	"aura": {
		Name:      "Aura Official",
		Primary:   lipgloss.Color("#BD93F9"), // Purple
		Secondary: lipgloss.Color("#8BE9FD"), // Cyan
		Accent:    lipgloss.Color("#FF79C6"), // Pink
		Foreground: lipgloss.Color("#F8F8F2"),
	},
	"deep-space": {
		Name:      "Deep Space",
		Primary:   lipgloss.Color("#624CAB"), // Deep Purple
		Secondary: lipgloss.Color("#2E004B"), // Darker Purple
		Accent:    lipgloss.Color("#000000"), // Pure Black
		Foreground: lipgloss.Color("#E0E0E0"),
	},
	"catppuccin": {
		Name:      "Catppuccin Mocha",
		Primary:   lipgloss.Color("#CBA6F7"), // Mauve
		Secondary: lipgloss.Color("#89B4FA"), // Blue
		Accent:    lipgloss.Color("#F5E0DC"), // Rosewater
		Foreground: lipgloss.Color("#CDD6F4"),
	},
	"tokyo-night": {
		Name:      "Tokyo Night",
		Primary:   lipgloss.Color("#7AA2F7"), // Blue
		Secondary: lipgloss.Color("#BB9AF7"), // Magenta
		Accent:    lipgloss.Color("#7DCFFF"), // Cyan
		Foreground: lipgloss.Color("#C0CAF5"),
	},
	"nord": {
		Name:      "Nordic",
		Primary:   lipgloss.Color("#88C0D0"), // Frost Blue
		Secondary: lipgloss.Color("#81A1C1"), // Blue
		Accent:    lipgloss.Color("#A3BE8C"), // Green
		Foreground: lipgloss.Color("#E5E9F0"),
	},
	"dracula": {
		Name:      "Dracula",
		Primary:   lipgloss.Color("#BD93F9"),
		Secondary: lipgloss.Color("#50FA7B"),
		Accent:    lipgloss.Color("#FF79C6"),
		Foreground: lipgloss.Color("#F8F8F2"),
	},
}

var ActiveTheme Theme

func Init(name string) {
	if t, ok := Themes[name]; ok {
		ActiveTheme = t
	} else {
		ActiveTheme = Themes["aura"]
	}
}

// Styles
func StylePrimary(s string) string {
	return lipgloss.NewStyle().Foreground(ActiveTheme.Primary).Render(s)
}

func StyleSecondary(s string) string {
	return lipgloss.NewStyle().Foreground(ActiveTheme.Secondary).Render(s)
}

func StyleAccent(s string) string {
	return lipgloss.NewStyle().Foreground(ActiveTheme.Accent).Render(s)
}

func StyleHeader(s string) string {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Foreground).
		Background(ActiveTheme.Primary).
		Padding(0, 1).
		Bold(true).
		Render(s)
}

func StyleTag(s string) string {
	return lipgloss.NewStyle().
		Foreground(ActiveTheme.Primary).
		Italic(true).
		Render("#" + s)
}

func StyleSuccess(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true).Render(s)
}

func StyleError(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true).Render(s)
}

func StyleWarning(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true).Render(s)
}

func StyleInfo(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).Render(s)
}

func StyleTitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Background(ActiveTheme.Primary).
		Foreground(ActiveTheme.Foreground).
		Padding(0, 1)
}

// GetSurveyTheme returns a custom theme for survey
func GetSurveyTheme() *SurveyTheme {
	return &SurveyTheme{}
}

type SurveyTheme struct {
	survey.AskOpt
}

// We will implement icons and colors if we want to customize survey further, 
// for now we use the default and just style with Aura colors where possible in the strings.
