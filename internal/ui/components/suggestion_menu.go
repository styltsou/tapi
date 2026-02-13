package components

import (
	"github.com/styltsou/tapi/internal/ui/styles"

	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SuggestionModel handles the autocomplete popup
type SuggestionModel struct {
	list    list.Model
	Visible bool
	Width   int
	Height  int
	filter  string
	
	onSelect func(string) tea.Msg
	onCancel func() tea.Msg
}

type suggestionItem struct {
	text string
	desc string
}

func (i suggestionItem) Title() string       { return i.text }
func (i suggestionItem) Description() string { return i.desc }
func (i suggestionItem) FilterValue() string { return i.text }

func NewSuggestionModel() SuggestionModel {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = styles.ModalSelectedStyle
	d.Styles.SelectedDesc = styles.ModalSelectedStyle.Copy().Foreground(styles.Gray)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Copy().Background(styles.DarkGray)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Copy().Background(styles.DarkGray).Foreground(styles.Gray)

	l := list.New([]list.Item{}, d, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(2).Background(styles.DarkGray)

	return SuggestionModel{
		list:    l,
		Visible: false,
	}
}

func (m *SuggestionModel) Show(options map[string]string, filter string, onSelect func(string) tea.Msg, onCancel func() tea.Msg) {
	var items []list.Item
	
	// Sort keys for stability
	keys := make([]string, 0, len(options))
	for k := range options {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		items = append(items, suggestionItem{text: k, desc: options[k]})
	}
	
	m.list.SetItems(items)
	m.filter = filter
	
	// If filter is provided, key presses usually handle it, but we might want to pre-filter?
	// Bubbletea list filtering is interactive. We just show.
	
	m.onSelect = onSelect
	m.onCancel = onCancel
	m.Visible = true
	m.list.Select(0)
}

func (m *SuggestionModel) Hide() {
	m.Visible = false
}

func (m *SuggestionModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	// Suggestion box is small
	m.list.SetSize(30, 10)
}

func (m SuggestionModel) Update(msg tea.Msg) (SuggestionModel, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			if m.onCancel != nil {
				return m, func() tea.Msg { return m.onCancel() }
			}
			return m, nil
		
		case "enter":
			if i, ok := m.list.SelectedItem().(suggestionItem); ok {
				m.Hide()
				if m.onSelect != nil {
					return m, func() tea.Msg { return m.onSelect(i.text) }
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SuggestionModel) View() string {
	if !m.Visible {
		return ""
	}
	
	width := 30 // Fixed width in SetSize
	bgStyle := lipgloss.NewStyle().Background(styles.DarkGray)
	
	content := styles.Solidify(m.list.View(), width, bgStyle)
	
	return styles.ModalStyle.Render(content)
}
