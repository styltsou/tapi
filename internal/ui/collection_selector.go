package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/storage"
)

// CollectionSelectorModel handles the collection selection modal
type CollectionSelectorModel struct {
	width       int
	height      int
	visible     bool
	collections []storage.Collection
	list        list.Model
}

func NewCollectionSelectorModel() CollectionSelectorModel {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Collection"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	// Apply custom styles if needed, but ModalStyle will wrap it
	l.Styles.Title = TitleStyle
	
	return CollectionSelectorModel{
		list:    l,
		visible: true, // Start visible as per plan
	}
}

func (m *CollectionSelectorModel) SetSize(width, height int) {
	m.width = width
	m.height = height
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
	case CollectionsLoadedMsg:
		m.SetCollections(msg.Collections)
		return m, nil

	case tea.KeyMsg:
		if !m.visible {
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
			m.visible = false
			return m, nil

		case "enter":
			if selected, ok := m.list.SelectedItem().(collectionItem); ok {
				m.visible = false
				return m, func() tea.Msg {
					return CollectionSelectedMsg{Collection: selected.collection}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m CollectionSelectorModel) View() string {
	if !m.visible {
		return ""
	}
	return ModalStyle.Render(m.list.View())
}

// collectionItem implements list.Item
type collectionItem struct {
	collection storage.Collection
}

func (i collectionItem) Title() string       { return i.collection.Name }
func (i collectionItem) Description() string { return "" }
func (i collectionItem) FilterValue() string { return i.collection.Name }
