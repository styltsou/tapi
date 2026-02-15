package components

import (
	"strings"

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
	d.Styles.NormalTitle = d.Styles.NormalTitle.Copy()
	d.Styles.NormalDesc = d.Styles.NormalDesc.Copy().Foreground(styles.Gray)

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Select Collection"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Paginator.ActiveDot = styles.PrimaryColorStyle.Render("■ ")
	l.Paginator.InactiveDot = styles.DimStyle.Render("□ ")
	l.Styles.Title = styles.TitleStyle.Copy().Background(styles.DarkGray).Foreground(styles.PrimaryColor)
	
	return CollectionSelectorModel{
		list:    l,
		Visible: false,
	}
}

func (m *CollectionSelectorModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.updateListSize()
}

func (m *CollectionSelectorModel) updateListSize() {
	numItems := len(m.list.Items())
	if numItems == 0 {
		numItems = 4 // Fallback for initial state
	}

	// Each item in default delegate is 3 lines (2 for text + 1 spacing)
	neededHeight := numItems * 3
	if numItems > 0 {
		neededHeight -= 1 // No spacer after the last item
	}

	// Constraint: At least 3 collections + New (4 items total = 11 lines)
	// But not taller than half the screen
	finalHeight := max(11, min(neededHeight, m.Height/2))
	m.list.SetSize(m.Width/3, finalHeight)
}

func (m *CollectionSelectorModel) SetCollections(collections []storage.Collection) {
	m.collections = collections
	
	items := make([]list.Item, len(collections))
	for i, col := range collections {
		items[i] = collectionItem{col}
	}
	// Add New Collection option at the end
	items = append(items, newCollectionOption{})
	m.list.SetItems(items)
	m.updateListSize()
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
			m.Visible = false
			return m, nil

		case "enter":
			selected := m.list.SelectedItem()
			if item, ok := selected.(collectionItem); ok {
				m.Visible = false
				return m, func() tea.Msg {
					return uimsg.CollectionSelectedMsg{Collection: item.collection}
				}
			}
			if _, ok := selected.(newCollectionOption); ok {
				m.Visible = false
				return m, func() tea.Msg {
					return uimsg.PromptForInputMsg{
						Title:       "New Collection",
						Placeholder: "Enter collection name",
						OnCommit:    func(val string) tea.Msg { return uimsg.CreateCollectionMsg{Name: val} },
					}
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
	
	width := m.Width / 3
	
	// List View (Hide default title)
	m.list.SetShowTitle(false)
	
	content := strings.TrimRight(m.list.View(), "\n\r ")
	content = styles.Solidify(content, width, lipgloss.NewStyle())
	
	return styles.WithBorderTitle(styles.SelectorModalStyle, content, "Collections")
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

// newCollectionOption implements list.Item
type newCollectionOption struct{}

func (i newCollectionOption) Title() string       { return "+ New collection" }
func (i newCollectionOption) Description() string { return "Create a fresh collection" }
func (i newCollectionOption) FilterValue() string { return "new collection" }
