package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/ui/styles"
)

// ConfirmationModel handles simple Yes/No confirmation prompts
type ConfirmationModel struct {
	Title       string
	OnConfirm   tea.Msg
	OnCancel    tea.Msg
	Width       int
	Height      int
}

func NewConfirmationModel(title string, onConfirm, onCancel tea.Msg) ConfirmationModel {
	return ConfirmationModel{
		Title:     title,
		OnConfirm: onConfirm,
		OnCancel:  onCancel,
	}
}

func (m ConfirmationModel) Init() tea.Cmd {
	return nil
}

func (m ConfirmationModel) Update(msg tea.Msg) (ConfirmationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			return m, func() tea.Msg { return m.OnConfirm }
		case "n", "N", "esc":
			return m, func() tea.Msg { return m.OnCancel }
		}
	}
	return m, nil
}

func (m ConfirmationModel) View() string {
	title := ""
	if m.Title != "" {
		title = styles.ErrorColorStyle.Bold(true).Render(m.Title) + "\n\n"
	}

	help := styles.DimStyle.Render("y/Enter: confirm  n/Esc: cancel")

	// Determine width based on content
	helpW := lipgloss.Width(help)
	titleW := lipgloss.Width(title)
	width := max(max(20, helpW), titleW) + 4 // Padding/Borders

	// Center content
	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		help,
	)

	bgStyle := lipgloss.NewStyle().Background(styles.DarkGray)
	solidified := styles.Solidify(content, width, bgStyle)

	return styles.ConfirmationModalStyle.Render(solidified)
}

func (m *ConfirmationModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}
