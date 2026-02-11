package ui

import (
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
	width       int
	height      int
	collections []storage.Collection
	cursor      int // 0..len(collections) where last item is "New Collection"
}

func NewWelcomeModel() WelcomeModel {
	return WelcomeModel{}
}

func (m *WelcomeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
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
		case "enter":
			if m.cursor < len(m.collections) {
				// Selected an existing collection
				col := m.collections[m.cursor]
				return m, func() tea.Msg {
					return CollectionSelectedMsg{Collection: col}
				}
			}
			// "New Collection" selected
			return m, func() tea.Msg {
				return PromptForInputMsg{
					Title:       "New Collection",
					Placeholder: "Enter collection name",
					OnCommit: func(val string) tea.Msg {
						return CreateCollectionMsg{Name: val}
					},
				}
			}
		case "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m WelcomeModel) View() string {
	var sb strings.Builder

	// Logo
	logoStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)
	sb.WriteString(logoStyle.Render(asciiLogo))
	sb.WriteString("\n")

	subtitleStyle := lipgloss.NewStyle().
		Foreground(Gray).
		Italic(true)
	sb.WriteString(subtitleStyle.Render("        Terminal API Client"))
	sb.WriteString("\n\n")

	// Section header
	sectionStyle := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)
	sb.WriteString(sectionStyle.Render("  Recent Collections"))
	sb.WriteString("\n\n")

	// Collection items
	normalItem := lipgloss.NewStyle().Foreground(lipgloss.Color("#cccccc")).PaddingLeft(2)
	activeItem := lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(PrimaryColor).
		PaddingLeft(1)

	if len(m.collections) == 0 {
		sb.WriteString(DimStyle.Render("    No collections yet. Create one to get started!"))
		sb.WriteString("\n\n")
	} else {
		for i, col := range m.collections {
			reqCount := fmt.Sprintf("%d requests", len(col.Requests))
			label := fmt.Sprintf("%-30s %s", col.Name, DimStyle.Render(reqCount))
			if i == m.cursor {
				sb.WriteString(activeItem.Render(label))
			} else {
				sb.WriteString(normalItem.Render(label))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// "New Collection" action
	newLabel := "  + New Collection"
	if m.cursor == len(m.collections) {
		sb.WriteString(activeItem.Render(newLabel))
	} else {
		sb.WriteString(normalItem.Render(newLabel))
	}
	sb.WriteString("\n\n")

	// Hints
	hintStyle := lipgloss.NewStyle().Foreground(Gray)
	sb.WriteString(hintStyle.Render("  j/k navigate  •  Enter select  •  q quit"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, sb.String())
}
