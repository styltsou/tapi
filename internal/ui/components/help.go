// internal/ui/help_model.go
package components

import (
	"github.com/styltsou/tapi/internal/ui/styles"

	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpOverlayModel displays a full-screen keybinding cheatsheet
type HelpOverlayModel struct {
	Visible bool
	Width   int
	Height  int
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
	m.Width = width
	m.Height = height
}

func (m *HelpOverlayModel) Toggle() {
	m.Visible = !m.Visible
}

func (m HelpOverlayModel) Update(msg tea.Msg) (HelpOverlayModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc", "q":
			m.Visible = false
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
				{"m", "Menu"},
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
	if !m.Visible {
		return ""
	}

	categories := getHelpCategories()

	bg := styles.DarkGray
	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7aa2f7")).
		Background(bg).
		Bold(true).
		Width(16).
		Align(lipgloss.Right)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#a9b1d6")).
		Background(bg).
		PaddingLeft(2)

	catTitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#e0af68")).
		Background(bg).
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
			// Add background to the whole line to be safe
			sb.WriteString(lipgloss.NewStyle().Background(bg).Render(line) + "\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.Gray).Background(bg).Render("Press ? or ESC to close"))

	width := min(50, m.Width-4)
	bgStyle := lipgloss.NewStyle().Background(styles.DarkGray)
	
	content := styles.Solidify(sb.String(), width, bgStyle)
	
	return styles.ModalStyle.Render(content)
}
