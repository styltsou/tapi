package components

import (
	"github.com/styltsou/tapi/internal/ui/styles"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel is a generic model for prompting user input
type InputModel struct {
	TextInput   textinput.Model
	Title       string
	OnCommitMsg func(value string) tea.Msg
	OnCancelMsg func() tea.Msg
	Width       int
	Height      int
}

func NewInputModel(title string, placeholder string, onCommit func(string) tea.Msg, onCancel func() tea.Msg) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 40

	return InputModel{
		TextInput:   ti,
		Title:       title,
		OnCommitMsg: onCommit,
		OnCancelMsg: onCancel,
	}
}

func (m *InputModel) SetValue(val string) {
	m.TextInput.SetValue(val)
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.OnCommitMsg != nil {
				return m, func() tea.Msg {
					return m.OnCommitMsg(m.TextInput.Value())
				}
			}
		case "esc":
			if m.OnCancelMsg != nil {
				return m, func() tea.Msg {
					return m.OnCancelMsg()
				}
			}
		}
	}

	m.TextInput, cmd = m.TextInput.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	width := 45 // Slightly narrower for better look

	// Ensure the input itself doesn't exceed the width
	m.TextInput.Width = width - 4

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.TextInput.View(),
	)

	solidified := styles.Solidify(content, width, lipgloss.NewStyle())
	return styles.WithBorderTitle(styles.ModalStyle, solidified, m.Title)
}

func (m *InputModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}
