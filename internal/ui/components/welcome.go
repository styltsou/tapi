package components

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/storage"
)

const asciiLogo = `
 ████████╗ █████╗ ██████╗ ██╗
 ╚══██╔══╝██╔══██╗██╔══██╗██║
    ██║   ███████║██████╔╝██║
    ██║   ██╔══██║██╔═══╝ ██║
    ██║   ██║  ██║██║     ██║
    ╚═╝   ╚═╝  ╚═╝╚═╝     ╚═╝`

// WelcomeModel handles the splash / welcome screen
type WelcomeModel struct {
	Width       int
	Height      int
	collections []storage.Collection
	cursor      int // 0..len(collections) where last item is "New Collection"
}

func NewWelcomeModel() WelcomeModel {
	return WelcomeModel{}
}

func (m *WelcomeModel) SetSize(width, height int) {
	m.Width = width
	m.Height = height
}

func (m *WelcomeModel) SetCollections(collections []storage.Collection) {
	m.collections = collections
	m.cursor = 0
}

func (m WelcomeModel) itemCount() int {
	return len(m.collections) + 1 // +1 for "New Collection"
}

func (m WelcomeModel) Update(msg tea.Msg) (WelcomeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if m.cursor < m.itemCount()-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "enter", "c":
			if msg.String() == "c" || m.cursor == len(m.collections) {
				// "New Collection" selected
				return m, func() tea.Msg {
					return uimsg.PromptForInputMsg{
						Title:       "New Collection",
						Placeholder: "Enter collection name",
						OnCommit: func(val string) tea.Msg {
							return uimsg.CreateCollectionMsg{Name: val}
						},
					}
				}
			}

			if m.cursor < len(m.collections) {
				// Selected an existing collection
				col := m.collections[m.cursor]
				return m, func() tea.Msg {
					return uimsg.CollectionSelectedMsg{Collection: col}
				}
			}
		case "d":
			if m.cursor < len(m.collections) {
				col := m.collections[m.cursor]
				return m, func() tea.Msg {
					return uimsg.ConfirmActionMsg{
						Title: "Delete collection: " + styles.ErrorColorStyle.Render(col.Name),
						OnConfirm: uimsg.DeleteCollectionMsg{Name: col.Name},
					}
				}
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m WelcomeModel) View() string {
	var (
		logo          string
		subtitle      string
		sectionHeader string
		content       string
		hints         string
	)

	// Logo
	logoStyle := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true)
	logo = logoStyle.Render(strings.TrimSpace(asciiLogo))

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.Gray).
		Italic(true)
	subtitle = subtitleStyle.Render("Terminal API Client")

	// Section header
	sectionStyle := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true).
		MarginBottom(1)
	sectionHeader = sectionStyle.Render("Recent Collections")

	// Calculate max width for collection list to ensure consistent alignment
	const minListWidth = 40
	listWidth := minListWidth
	for _, col := range m.collections {
		reqCount := fmt.Sprintf("%d requests", len(col.Requests))
		w := len(col.Name) + len(reqCount) + 4 // +4 for spacing
		if w > listWidth {
			listWidth = w
		}
	}

	// Also consider "New Collection" width with "c" hint
	newCollWidth := len("+ New Collection") + 1 + 4 // +1 for "c", +4 for spacing
	if newCollWidth > listWidth {
		listWidth = newCollWidth
	}

	// Collection items
	var listItems []string
	
	// Create styles with dynamic width
	normalItem := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cccccc")).
		PaddingLeft(2).
		Width(listWidth).
		MarginBottom(1)
		
	activeItem := lipgloss.NewStyle().
		Foreground(styles.PrimaryColor).
		Bold(true).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(styles.PrimaryColor).
		PaddingLeft(1).
		Width(listWidth).
		MarginBottom(1)

	if len(m.collections) == 0 {
		listItems = append(listItems, styles.DimStyle.Render("No collections yet. Create one to get started!"))
	} else {
		for i, col := range m.collections {
			reqCount := fmt.Sprintf("%d requests", len(col.Requests))
			// Use spaces to push reqCount to the right, or just simple spacing
			// For a true "space-between" effect we need to calculate padding manually or use lipgloss.PlaceHorizontal
			
			// Simple approach: Name ... (gap) ... Count
			labelRaw := col.Name
			countRaw := styles.DimStyle.Render(reqCount)
			
			// Calculate available space for padding
			// We effectively want: "Name <space> Count"
			// But since we want the whole block centered, we just format it nicely.
			// Let's use the listWidth we calculated.
			
			availableSpace := listWidth - len(col.Name) - lipgloss.Width(reqCount) - 3 // -3 for padding/border diff
			if availableSpace < 1 {
				availableSpace = 1
			}
			
			label := fmt.Sprintf("%s%s%s", labelRaw, strings.Repeat(" ", availableSpace), countRaw)

			if i == m.cursor {
				listItems = append(listItems, activeItem.Render(label))
			} else {
				listItems = append(listItems, normalItem.Render(label))
			}
		}
	}

	// "New Collection" action
	const newLabelRaw = "+ New Collection"
	const shortHintRaw = "c"
	shortHint := styles.DimStyle.Render(shortHintRaw)

	availableSpaceNew := listWidth - len(newLabelRaw) - lipgloss.Width(shortHintRaw) - 3
	if availableSpaceNew < 1 {
		availableSpaceNew = 1
	}
	newLabel := fmt.Sprintf("%s%s%s", newLabelRaw, strings.Repeat(" ", availableSpaceNew), shortHint)

	if m.cursor == len(m.collections) {
		listItems = append(listItems, activeItem.Render(newLabel))
	} else {
		listItems = append(listItems, normalItem.Render(newLabel))
	}

	content = lipgloss.JoinVertical(lipgloss.Left, listItems...)

	// Hints
	hintStyle := lipgloss.NewStyle().Foreground(styles.Gray).MarginTop(2)
	hints = hintStyle.Render("j/k navigate  •  d delete  •  q quit")

	// Assemble the full view
	// We want to center everything horizontally
	
	// Utility to center a block horizontally in the available width
	center := func(s string) string {
		return lipgloss.PlaceHorizontal(m.Width, lipgloss.Center, s)
	}

	// Join all parts vertically with some spacing
	ui := lipgloss.JoinVertical(lipgloss.Center,
		center(logo),
		center(subtitle),
		"\n", // Spacer
		center(sectionHeader),
		center(content),
		center(hints),
	)

	// Place the entire UI vertically centered in the terminal
	return lipgloss.PlaceVertical(m.Height, lipgloss.Center, ui)
}
