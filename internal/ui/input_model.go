package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// InputModel is a generic model for prompting user input
type InputModel struct {
	textInput   textinput.Model
	title       string
	onCommitMsg func(value string) tea.Msg
	onCancelMsg func() tea.Msg
	width       int
	height      int
}

func NewInputModel(title string, placeholder string, onCommit func(string) tea.Msg, onCancel func() tea.Msg) InputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40

	return InputModel{
		textInput:   ti,
		title:       title,
		onCommitMsg: onCommit,
		onCancelMsg: onCancel,
	}
}

func (m *InputModel) SetValue(val string) {
	m.textInput.SetValue(val)
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (InputModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.onCommitMsg != nil {
				return m, func() tea.Msg {
					return m.onCommitMsg(m.textInput.Value())
				}
			}
		case tea.KeyEsc:
			if m.onCancelMsg != nil {
				return m, func() tea.Msg {
					return m.onCancelMsg()
				}
			}
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m InputModel) View() string {
	// Minimal floating prompt — LazyVim command-line style
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, 2).
		Width(50)

	titleLine := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render(m.title)

	hintLine := DimStyle.Render("Enter: confirm  Esc: cancel")

	content := lipgloss.JoinVertical(lipgloss.Left,
		titleLine,
		m.textInput.View(),
		hintLine,
	)

	return inputStyle.Render(content)
}

func (m *InputModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}
