package components

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/storage"
)

// CollectionSelectorModel handles the collection selection modal
type CollectionSelectorModel struct {
	Width       int
	Height      int
	Visible     bool
	collections []storage.Collection
	list        list.Model
}

func NewCollectionSelectorModel() CollectionSelectorModel {
	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = styles.ModalSelectedStyle
	d.Styles.SelectedDesc = styles.ModalSelectedStyle.Copy().Foreground(styles.Gray)
	d.Styles.NormalTitle = d.Styles.NormalTitle.Copy().Background(styles.DarkGray)
	d.Styles.NormalDesc = d.Styles.NormalDesc.Copy().Background(styles.DarkGray).Foreground(styles.Gray)

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Select Collection"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = styles.TitleStyle.Copy().Background(styles.DarkGray).Foreground(styles.PrimaryColor)
	
	return CollectionSelectorModel{
		list:    l,
		Visible: false,
	}
}

func (m *CollectionSelectorModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	// Modal size
	m.list.SetSize(width/2, height/2)
}

func (m *CollectionSelectorModel) SetCollections(collections []storage.Collection) {
	m.collections = collections
	
	items := make([]list.Item, len(collections))
	for i, col := range collections {
		items[i] = collectionItem{col}
	}
	m.list.SetItems(items)
}

func (m CollectionSelectorModel) Update(msg tea.Msg) (CollectionSelectorModel, tea.Cmd) {
	switch msg := msg.(type) {
	case uimsg.CollectionsLoadedMsg:
		m.SetCollections(msg.Collections)
		return m, nil

	case tea.KeyMsg:
		if !m.Visible {
			return m, nil
		}

		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "esc":
			// If it's the initial screen, maybe quitting is better? 
			// But generall esc usually closes modals. 
			// If there is no active collection, we might want to enforce selection.
			// For now, let's treat it as hide.
			m.Visible = false
			return m, nil

		case "enter":
			if selected, ok := m.list.SelectedItem().(collectionItem); ok {
				m.Visible = false
				return m, func() tea.Msg {
					return uimsg.CollectionSelectedMsg{Collection: selected.collection}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m CollectionSelectorModel) View() string {
	if !m.Visible {
		return ""
	}
	
	width := m.Width / 2
	bgStyle := lipgloss.NewStyle().Background(styles.DarkGray)
	
	content := styles.Solidify(m.list.View(), width, bgStyle)
	
	return styles.ModalStyle.Render(content)
}

// collectionItem implements list.Item
type collectionItem struct {
	collection storage.Collection
}

func (i collectionItem) Title() string       { return i.collection.Name }
func (i collectionItem) Description() string {
	if i.collection.BaseURL != "" {
		return i.collection.BaseURL
	}
	return fmt.Sprintf("%d requests", len(i.collection.Requests))
}
func (i collectionItem) FilterValue() string { return i.collection.Name }
