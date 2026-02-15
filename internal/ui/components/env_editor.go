package components

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/storage"
)

// EnvEditorModel handles editing the variables of an environment
type EnvEditorModel struct {
	env     storage.Environment
	inputs  []VarInput
	focused int
	Width   int
	Height  int
}

type VarInput struct {
	key   textinput.Model
	value textinput.Model
}

func NewEnvEditorModel() EnvEditorModel {
	return EnvEditorModel{
		inputs: []VarInput{},
	}
}

func (m *EnvEditorModel) SetEnvironment(env storage.Environment) {
	m.env = env
	m.inputs = []VarInput{}

	for k, v := range env.Variables {
		ki := textinput.New()
		ki.SetValue(k)
		ki.Width = 20

		vi := textinput.New()
		vi.SetValue(v)
		vi.Width = 40

		m.inputs = append(m.inputs, VarInput{key: ki, value: vi})
	}
	
	// Add one empty row if none exist
	if len(m.inputs) == 0 {
		m.AddRow()
	}

	m.focused = 0
	m.updateFocus()
}

func (m *EnvEditorModel) AddRow() {
	ki := textinput.New()
	ki.Placeholder = "Key"
	ki.Width = 20

	vi := textinput.New()
	vi.Placeholder = "Value"
	vi.Width = 40

	m.inputs = append(m.inputs, VarInput{key: ki, value: vi})
}

func (m *EnvEditorModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m EnvEditorModel) Update(msg tea.Msg) (EnvEditorModel, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return uimsg.BackMsg{} }
		
		case "ctrl+s":
			// Save variables back to env
			newVars := make(map[string]string)
			for _, input := range m.inputs {
				if input.key.Value() != "" {
					newVars[input.key.Value()] = input.value.Value()
				}
			}
			m.env.Variables = newVars
			return m, func() tea.Msg { return SaveEnvMsg{Env: m.env} }

		case "ctrl+a":
			m.AddRow()
			m.focused = (len(m.inputs) - 1) * 2
			m.updateFocus()
			return m, nil

		case "ctrl+d":
			row := m.focused / 2
			if row < len(m.inputs) && len(m.inputs) > 1 {
				m.inputs = append(m.inputs[:row], m.inputs[row+1:]...)
				if m.focused >= len(m.inputs)*2 {
					m.focused = len(m.inputs)*2 - 1
				}
				m.updateFocus()
			} else if len(m.inputs) == 1 {
				// Clear the last row instead of deleting if it's the only one?
				m.inputs[0].key.SetValue("")
				m.inputs[0].value.SetValue("")
				m.focused = 0
				m.updateFocus()
			}
			return m, nil

		case "tab":
			m.focused++
			if m.focused >= len(m.inputs)*2 {
				m.focused = 0
			}
			m.updateFocus()
			return m, nil

		case "shift+tab":
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs)*2 - 1
			}
			m.updateFocus()
			return m, nil

		case "up":
			if m.focused >= 2 {
				m.focused -= 2
				m.updateFocus()
			}
			return m, nil

		case "down":
			if m.focused < (len(m.inputs)-1)*2 {
				m.focused += 2
				m.updateFocus()
			}
			return m, nil
		
		case "left", "right":
			if m.focused%2 == 0 {
				m.focused++ // go to value
			} else {
				m.focused-- // go to key
			}
			m.updateFocus()
			return m, nil
		}
	}

	// Update focused input
	row := m.focused / 2
	isKey := m.focused%2 == 0

	if row < len(m.inputs) {
		var cmd tea.Cmd
		if isKey {
			m.inputs[row].key, cmd = m.inputs[row].key.Update(msg)
		} else {
			m.inputs[row].value, cmd = m.inputs[row].value.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *EnvEditorModel) updateFocus() {
	for i := range m.inputs {
		m.inputs[i].key.Blur()
		m.inputs[i].value.Blur()
	}

	row := m.focused / 2
	isKey := m.focused%2 == 0

	if row < len(m.inputs) {
		if isKey {
			m.inputs[row].key.Focus()
		} else {
			m.inputs[row].value.Focus()
		}
	}
}

func (m EnvEditorModel) View() string {
	var sb strings.Builder
	
	bg := styles.DarkGray
	sb.WriteString(styles.TitleStyle.Render(" Editing: " + m.env.Name + " "))
	sb.WriteString("\n\n")

	// Header row
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		styles.HeaderStyle.Copy().Background(bg).Width(24).Render("VARIABLE"),
		styles.HeaderStyle.Copy().Background(bg).Width(44).Render("VALUE"),
	)
	sb.WriteString(header + "\n")

	for i, input := range m.inputs {
		prefix := "  "
		if i == m.focused/2 {
			prefix = "> "
		}

		// Ensure prefix and inputs have background
		prefixStyle := styles.ModalSelectedStyle.Copy().Background(bg)
		dimStyle := styles.DimStyle.Copy().Background(bg)

		row := lipgloss.JoinHorizontal(lipgloss.Center,
			prefixStyle.Render(prefix),
			input.key.View(),
			dimStyle.Render(" : "),
			input.value.View(),
		)
		sb.WriteString(lipgloss.NewStyle().Background(bg).Render(row) + "\n")
	}

	sb.WriteString("\n")
	width := 72 // Variable(24) + Value(44) + " : "(3) + prefix(1) etc... let's be generous
	bgStyle := lipgloss.NewStyle().Background(styles.DarkGray)
	
	content := styles.Solidify(sb.String(), width, bgStyle)
	
	return styles.ModalStyle.Render(content)
}

func (m *EnvEditorModel) SetCursorMode(mode cursor.Mode) tea.Cmd {
	for i := range m.inputs {
		m.inputs[i].key.Cursor.SetMode(mode)
		m.inputs[i].value.Cursor.SetMode(mode)
	}
	if mode == cursor.CursorBlink {
		return textinput.Blink
	}
	return nil
}

type SaveEnvMsg struct {
	Env storage.Environment
}
