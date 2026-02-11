package ui

import (
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SuggestionModel handles the autocomplete popup
type SuggestionModel struct {
	list    list.Model
	visible bool
	width   int
	height  int
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
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	// Make it compact
	l.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(2)

	return SuggestionModel{
		list:    l,
		visible: false,
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
	m.visible = true
	m.list.Select(0)
}

func (m *SuggestionModel) Hide() {
	m.visible = false
}

func (m *SuggestionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Suggestion box is small
	m.list.SetSize(30, 10)
}

func (m SuggestionModel) Update(msg tea.Msg) (SuggestionModel, tea.Cmd) {
	if !m.visible {
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
	if !m.visible {
		return ""
	}
	// Use a distinct style for suggestion popup, maybe reusing ModalStyle but smaller/different position?
	// For now, ModalStyle is fine.
	return ModalStyle.Render(m.list.View())
}
