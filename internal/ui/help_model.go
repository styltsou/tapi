// internal/ui/help_model.go
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpOverlayModel displays a full-screen keybinding cheatsheet
type HelpOverlayModel struct {
	visible bool
	width   int
	height  int
}

// helpCategory groups related keybindings
type helpCategory struct {
	title    string
	bindings []helpBinding
}

type helpBinding struct {
	key  string
	desc string
}

func NewHelpOverlayModel() HelpOverlayModel {
	return HelpOverlayModel{}
}

func (m *HelpOverlayModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *HelpOverlayModel) Toggle() {
	m.visible = !m.visible
}

func (m HelpOverlayModel) Update(msg tea.Msg) (HelpOverlayModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc", "q":
			m.visible = false
			return m, nil
		}
	}
	return m, nil
}

func getHelpCategories() []helpCategory {
	return []helpCategory{
		{
			title: "General",
			bindings: []helpBinding{
				{"SPC", "Leader key"},
				{"i / Enter", "Insert mode"},
				{"ESC", "Normal mode"},
				{"?", "This help"},
				{"Ctrl+C", "Quit"},
			},
		},
		{
			title: "Leader Commands (SPC+…)",
			bindings: []helpBinding{
				{"r", "Run request"},
				{"s", "Save request"},
				{"e", "Toggle sidebar"},
				{"v", "Environments"},
				{"c", "Change collection"},
				{"k", "Command menu"},
				{"o", "Preview request"},
				{"y", "Copy as cURL"},
				{"w", "Close tab"},
				{"q", "Quit"},
			},
		},
		{
			title: "Navigation",
			bindings: []helpBinding{
				{"Tab", "Next pane"},
				{"Shift+Tab", "Prev pane"},
				{"gt", "Next tab"},
				{"gT", "Prev tab"},
				{"j / k", "Up / Down"},
			},
		},
		{
			title: "Request Builder",
			bindings: []helpBinding{
				{"Tab", "Next field"},
				{"Shift+Tab", "Prev field"},
				{"Ctrl+A", "Add row"},
				{"Ctrl+D", "Delete row"},
			},
		},
		{
			title: "Response Viewer",
			bindings: []helpBinding{
				{"/", "Search body"},
				{"n / N", "Next / Prev match"},
				{"c", "Copy body"},
			},
		},
	}
}

func (m HelpOverlayModel) View() string {
	if !m.visible {
		return ""
	}

	categories := getHelpCategories()

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7aa2f7")).
		Bold(true).
		Width(16).
		Align(lipgloss.Right)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a9b1d6")).
		PaddingLeft(2)

	catTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0af68")).
		Bold(true).
		MarginTop(1).
		MarginBottom(0)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7D56F4")).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1)

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("⌨  Keyboard Shortcuts"))
	sb.WriteString("\n")

	for _, cat := range categories {
		sb.WriteString(catTitleStyle.Render("── " + cat.title + " ──"))
		sb.WriteString("\n")
		for _, b := range cat.bindings {
			line := lipgloss.JoinHorizontal(lipgloss.Center,
				keyStyle.Render(b.key),
				descStyle.Render(b.desc),
			)
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(Gray).Render("Press ? or ESC to close"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 3).
		Width(min(50, m.width-4)).
		Background(lipgloss.Color("#1a1b26"))

	return boxStyle.Render(sb.String())
}
