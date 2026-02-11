// internal/ui/collections_model.go
package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/storage"
)

// CollectionsModel handles the collections list view
type CollectionsModel struct {
	width       int
	height      int
	collection  storage.Collection
	list        list.Model
}

func NewCollectionsModel() CollectionsModel {
	d := list.NewDefaultDelegate()
	d.ShowDescription = true
	d.SetHeight(2)
	d.Styles.SelectedTitle = SelectedStyle
	d.Styles.SelectedDesc = SelectedStyle.Copy().Foreground(Gray)

	l := list.New([]list.Item{}, d, 0, 0)
	l.Title = "Requests"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = TitleStyle

	return CollectionsModel{
		list: l,
	}
}

func (m *CollectionsModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width-4, height-4)
}

func (m *CollectionsModel) SetCollection(collection storage.Collection) {
	m.collection = collection
	
	// Sets the title to the collection name
	m.list.Title = collection.Name

	var items []list.Item
	// Add "Create New Request" item
	items = append(items, requestItem{isCreate: true})

	for _, req := range collection.Requests {
		items = append(items, requestItem{
			collection: collection,
			request:    req,
		})
	}

	m.list.SetItems(items)
}

func (m CollectionsModel) Update(msg tea.Msg) (CollectionsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		
		switch msg.String() {
		case "enter":
			if selected, ok := m.list.SelectedItem().(requestItem); ok {
				if selected.isCreate {
					return m, func() tea.Msg {
						return PromptForInputMsg{
							Title:       "Create New Request",
							Placeholder: "Enter request name",
							OnCommit: func(val string) tea.Msg {
								return CreateRequestMsg{Name: val}
							},
						}
					}
				}

				return m, func() tea.Msg {
					return RequestSelectedMsg{
						Request: selected.request,
						BaseURL: selected.collection.BaseURL,
					}
				}
			}
		
		case "d": // Delete Request
			if selected, ok := m.list.SelectedItem().(requestItem); ok && !selected.isCreate {
				return m, func() tea.Msg {
					return DeleteRequestMsg{
						CollectionName: selected.collection.Name,
						RequestName:    selected.request.Name,
					}
				}
			}

		case "D": // Delete Collection (Shift+d)
			if selected, ok := m.list.SelectedItem().(requestItem); ok && !selected.isCreate {
				return m, func() tea.Msg {
					return DeleteCollectionMsg{Name: selected.collection.Name}
				}
			}

		case "r": // Rename Collection
			if selected, ok := m.list.SelectedItem().(requestItem); ok && !selected.isCreate {
				oldName := selected.collection.Name
				return m, func() tea.Msg {
					return PromptForInputMsg{
						Title:       "Rename Collection: " + oldName,
						Placeholder: "Enter new collection name",
						OnCommit: func(val string) tea.Msg {
							return RenameCollectionMsg{OldName: oldName, NewName: val}
						},
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m CollectionsModel) View() string {
	return m.list.View()
}

// requestItem implements list.Item interface
type requestItem struct {
	collection storage.Collection
	request    storage.Request
	isCreate   bool
}

func (i requestItem) FilterValue() string {
	if i.isCreate {
		return "New Request Create"
	}
	return i.collection.Name + " " + i.request.Name
}

func (i requestItem) Title() string {
	if i.isCreate {
		return "+ Create New Request"
	}
	return MethodBadge(i.request.Method) + " " + i.request.Name
}

func (i requestItem) Description() string {
	if i.isCreate {
		return "Add a new request to your collection"
	}
	return DimStyle.Render(i.collection.Name)
}
