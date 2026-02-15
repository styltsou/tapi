package components

import (
	"strings"

	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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


func NewCommandMenuModel() CommandMenuModel {
	items := []list.Item{
		commandItem{title: "Change Collection", desc: "Switch to a different collection", action: uimsg.OpenCollectionSelectorMsg{}},
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
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = styles.ModalSelectedStyle
	d.Styles.SelectedDesc = styles.ModalSelectedStyle.Copy().Foreground(styles.Gray)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Copy()
	d.Styles.NormalDesc = d.Styles.NormalDesc.Copy().Foreground(styles.Gray)

	l := list.New(items, d, 0, 0)
	l.Title = "Menu"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Paginator.ActiveDot = styles.PrimaryColorStyle.Render("■ ")
	l.Paginator.InactiveDot = styles.DimStyle.Render("□ ")
	l.Styles.Title = styles.TitleStyle.Copy().Background(styles.DarkGray).Foreground(styles.PrimaryColor)
	l.Styles.StatusBar = l.Styles.StatusBar.Copy().Background(styles.DarkGray)
	l.KeyMap.Quit.SetKeys()

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

	menuWidth := min(50, m.Width-4)
	m.list.SetShowTitle(false)
	
	// Solidify the list view content without a background
	content := strings.TrimRight(m.list.View(), "\n\r ")
	content = styles.Solidify(content, menuWidth, lipgloss.NewStyle())
	
	return styles.WithBorderTitle(styles.ModalStyle, content, "Menu")
}
