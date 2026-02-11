// internal/ui/env_model.go
package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/storage"
)

// EnvModel handles the environment selector modal
type EnvModel struct {
	width        int
	height       int
	visible      bool
	environments []storage.Environment
	list         list.Model
	selected     *storage.Environment
}

func NewEnvModel() EnvModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Environment"
	l.SetShowHelp(false)

	return EnvModel{
		list:    l,
		visible: false,
	}
}

func (m *EnvModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Modal is centered and smaller
	m.list.SetSize(width/2, height/2)
}

func (m *EnvModel) SetEnvironments(envs []storage.Environment) {
	m.environments = envs

	items := make([]list.Item, 0, len(envs)+1)
	// Add "Create New" item first
	items = append(items, envItem{isCreate: true})

	for _, env := range envs {
		items = append(items, envItem{env: env})
	}

	m.list.SetItems(items)
}

func (m EnvModel) Update(msg tea.Msg) (EnvModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.visible = false
			return m, nil

		case "enter":
			if selected, ok := m.list.SelectedItem().(envItem); ok {
				if selected.isCreate {
					m.visible = false
					// Prompt for environment name
					return m, func() tea.Msg {
						return PromptForInputMsg{
							Title:       "Create New Environment",
							Placeholder: "Enter environment name",
							OnCommit: func(val string) tea.Msg {
								return CreateEnvMsg{Name: val}
							},
						}
					}
				}

				m.selected = &selected.env
				m.visible = false

				return m, func() tea.Msg {
					return EnvChangedMsg{NewEnv: selected.env}
				}
			}

		case "e":
			if selected, ok := m.list.SelectedItem().(envItem); ok && !selected.isCreate {
				return m, func() tea.Msg {
					return FocusMsg{Target: ViewEnvEditor, Data: selected.env}
				}
			}
		
		case "d", "x":
			if selected, ok := m.list.SelectedItem().(envItem); ok && !selected.isCreate {
				return m, func() tea.Msg {
					return DeleteEnvMsg{Name: selected.env.Name}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m EnvModel) View() string {
	if !m.visible {
		return ""
	}

	// Centered modal styling
	return ModalStyle.Render(m.list.View())
}

// envItem implements list.Item interface
type envItem struct {
	env      storage.Environment
	isCreate bool
}

func (i envItem) FilterValue() string {
	if i.isCreate {
		return "Create New Environment"
	}
	return i.env.Name
}

func (i envItem) Title() string {
	if i.isCreate {
		return "+ Create New Environment"
	}
	return i.env.Name
}

func (i envItem) Description() string {
	if i.isCreate {
		return "Create a new environment configuration"
	}
	return ""
}
