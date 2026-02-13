// internal/ui/env_model.go
package components

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/storage"
)

// EnvModel handles the environment selector modal
type EnvModel struct {
	Width        int
	Height       int
	Visible      bool
	environments []storage.Environment
	list         list.Model
	selected     *storage.Environment
}

func NewEnvModel() EnvModel {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = styles.ModalSelectedStyle
	d.Styles.SelectedDesc = styles.ModalSelectedStyle.Copy().Foreground(styles.Gray)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Copy().Background(styles.DarkGray)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Copy().Background(styles.DarkGray).Foreground(styles.Gray)

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Select Environment"
	l.SetShowHelp(false)
	l.Styles.Title = styles.TitleStyle.Copy().Background(styles.DarkGray).Foreground(styles.PrimaryColor)


	return EnvModel{
		list:    l,
		Visible: false,
	}
}

func (m *EnvModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
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
			m.Visible = false
			return m, nil

		case "enter":
			if selected, ok := m.list.SelectedItem().(envItem); ok {
				if selected.isCreate {
					m.Visible = false
					// Prompt for environment name
					return m, func() tea.Msg {
						return uimsg.PromptForInputMsg{
							Title:       "Create New Environment",
							Placeholder: "Enter environment name",
							OnCommit: func(val string) tea.Msg {
								return uimsg.CreateEnvMsg{Name: val}
							},
						}
					}
				}

				m.selected = &selected.env
				m.Visible = false

				return m, func() tea.Msg {
					return uimsg.EnvChangedMsg{NewEnv: selected.env}
				}
			}

		case "e":
			if selected, ok := m.list.SelectedItem().(envItem); ok && !selected.isCreate {
				return m, func() tea.Msg {
					return uimsg.FocusMsg{Target: uimsg.ViewEnvEditor, Data: selected.env}
				}
			}
		
		case "d", "x":
			if selected, ok := m.list.SelectedItem().(envItem); ok && !selected.isCreate {
				return m, func() tea.Msg {
					return uimsg.DeleteEnvMsg{Name: selected.env.Name}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m EnvModel) View() string {
	if !m.Visible {
		return ""
	}

	width := m.Width / 2
	bgStyle := lipgloss.NewStyle().Background(styles.DarkGray)
	
	content := styles.Solidify(m.list.View(), width, bgStyle)
	
	return styles.ModalStyle.Render(content)
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
