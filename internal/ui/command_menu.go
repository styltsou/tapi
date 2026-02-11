package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandMenuModel handles the spotlight-like searchable command menu
type CommandMenuModel struct {
	list    list.Model
	visible bool
	width   int
	height  int
}

type commandItem struct {
	title, desc string
	action      tea.Msg
}

func (i commandItem) Title() string       { return i.title }
func (i commandItem) Description() string { return i.desc }
func (i commandItem) FilterValue() string { return i.title }

// OpenCollectionSelectorMsg signals to open the collection selector
type OpenCollectionSelectorMsg struct{}

func NewCommandMenuModel() CommandMenuModel {
	items := []list.Item{
		commandItem{title: "Toggle Sidebar", desc: "Show/Hide the collection sidebar", action: ToggleSidebarMsg{}},
		commandItem{title: "Change Collection", desc: "Switch to a different collection", action: OpenCollectionSelectorMsg{}},
		commandItem{title: "Import Collection", desc: "Import from Postman, Insomnia, or cURL", action: PromptForInputMsg{
			Title:       "Import Collection",
			Placeholder: "Path to file (Postman JSON, Insomnia JSON, or cURL)",
			OnCommit:    func(val string) tea.Msg { return ImportCollectionMsg{Path: val} },
		}},
		commandItem{title: "Export Collection", desc: "Export current collection as YAML", action: PromptForInputMsg{
			Title:       "Export Collection",
			Placeholder: "Destination path (e.g. ~/my-collection.yaml)",
			OnCommit:    func(val string) tea.Msg { return ExportCollectionMsg{DestPath: val} },
		}},
		commandItem{title: "Environments", desc: "Manage environment variables", action: FocusMsg{Target: ViewEnvironments}},
		commandItem{title: "Copy as cURL", desc: "Copy current request as a cURL command", action: CopyAsCurlMsg{}},
		commandItem{title: "Quit", desc: "Exit the application", action: tea.Quit()},
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = SelectedStyle
	d.Styles.SelectedDesc = SelectedStyle.Copy().Foreground(Gray)

	l := list.New(items, d, 0, 0)
	l.Title = "Commands"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	return CommandMenuModel{
		list: l,
	}
}

func (m *CommandMenuModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Constrain menu to reasonable size
	menuWidth := min(50, width-4)
	menuHeight := min(14, height/2)
	m.list.SetSize(menuWidth, menuHeight)
}

func (m CommandMenuModel) Update(msg tea.Msg) (CommandMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.visible = false
			return m, nil
		case "enter":
			if selected, ok := m.list.SelectedItem().(commandItem); ok {
				m.visible = false
				return m, func() tea.Msg {
					return selected.action
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m CommandMenuModel) View() string {
	if !m.visible {
		return ""
	}

	return ModalStyle.Render(m.list.View())
}
