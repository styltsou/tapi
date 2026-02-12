package components

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandMenuModel handles the spotlight-like searchable command menu
type CommandMenuModel struct {
	list    list.Model
	Visible bool
	Width   int
	Height  int
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
		commandItem{title: "Toggle Sidebar", desc: "Show/Hide the collection sidebar", action: uimsg.ToggleSidebarMsg{}},
		commandItem{title: "Change Collection", desc: "Switch to a different collection", action: OpenCollectionSelectorMsg{}},
		commandItem{title: "Import Collection", desc: "Import from Postman, Insomnia, or cURL", action: uimsg.PromptForInputMsg{
			Title:       "Import Collection",
			Placeholder: "Path to file (Postman JSON, Insomnia JSON, or cURL)",
			OnCommit:    func(val string) tea.Msg { return uimsg.ImportCollectionMsg{Path: val} },
		}},
		commandItem{title: "Export Collection", desc: "Export current collection as YAML", action: uimsg.PromptForInputMsg{
			Title:       "Export Collection",
			Placeholder: "Destination path (e.g. ~/my-collection.yaml)",
			OnCommit:    func(val string) tea.Msg { return uimsg.ExportCollectionMsg{DestPath: val} },
		}},
		commandItem{title: "Environments", desc: "Manage environment variables", action: uimsg.FocusMsg{Target: uimsg.ViewEnvironments}},
		commandItem{title: "Copy as cURL", desc: "Copy current request as a cURL command", action: uimsg.CopyAsCurlMsg{}},
		commandItem{title: "Quit", desc: "Exit the application", action: tea.Quit()},
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = styles.SelectedStyle
	d.Styles.SelectedDesc = styles.SelectedStyle.Copy().Foreground(styles.Gray)

	l := list.New(items, d, 0, 0)
	l.Title = "Commands"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.TitleStyle

	return CommandMenuModel{
		list: l,
	}
}

func (m *CommandMenuModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
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
			m.Visible = false
			return m, nil
		case "enter":
			if selected, ok := m.list.SelectedItem().(commandItem); ok {
				m.Visible = false
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
	if !m.Visible {
		return ""
	}

	return styles.ModalStyle.Render(m.list.View())
}
