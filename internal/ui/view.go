package ui

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	// Check if terminal is too small
	if m.tooSmall {
		return lipgloss.Place(m.Width, m.Height,
			lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("Terminal is too small.\nPlease resize to at least %dx%d.", minWidth, minHeight),
		)
	}

	// Welcome screen
	if m.state == uimsg.ViewWelcome {
		return m.welcome.View()
	}

	// 1. Header
	header := m.viewHeader()

	// Tab Bar
	tabBar := m.viewTabBar()

	// 2. Dashboard Content
	dashboard := m.viewDashboard()

	// 3. Status Bar
	bar := m.viewStatusBar()

	// Final Assembly
	fullView := lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabBar,
		dashboard,
		bar,
	)

	// Modals (Overlays)
	var overlay string
	if m.helpOverlay.Visible {
		overlay = m.helpOverlay.View()
	} else if m.menu.Visible {
		overlay = m.menu.View()
	} else if m.env.Visible {
		overlay = m.env.View()
	} else if m.state == uimsg.ViewEnvEditor {
		overlay = m.envEditor.View()
	} else if m.collectionSelector.Visible {
		overlay = m.collectionSelector.View()
	}

	if overlay != "" {
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("#111111")),
		)
	}

	// Input prompt floats on top of the current view (transparent overlay)
	if m.state == uimsg.ViewInput {
		inputOverlay := m.input.View()
		return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, inputOverlay,
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
			lipgloss.WithWhitespaceChars("░"),
		)
	}

	return fullView
}


// ========================================
// View Helper Functions
// ========================================



func (m Model) viewHeader() string {
	headerText := " TAPI "
	if m.currentCollection != nil {
		headerText += " • " + m.currentCollection.Name
	}
	header := styles.TitleStyle.Render(headerText)
	if m.currentEnv != nil {
		header += " " + styles.StatusStyle.Render("Env: "+m.currentEnv.Name)
	}
	header += "\n"
	return header
}

func (m Model) viewTabBar() string {
	var tabBar string
	if len(m.tabs) > 0 {
		var tabs []string
		for i, tab := range m.tabs {
			style := lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(lipgloss.Color("#555555")).
				Background(lipgloss.Color("#1a1b26"))
			
			if i == m.activeTab {
				style = style.
					Foreground(lipgloss.Color("#ffffff")).
					Background(lipgloss.Color("#7D56F4")).
					Bold(true)
			}
			tabs = append(tabs, style.Render(tab.Label))
		}
		tabBar = lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
		tabBar += "\n"
	}
	return tabBar
}

func (m Model) viewDashboard() string {
	sStyle, rStyle, respStyle := styles.InactivePaneStyle, styles.InactivePaneStyle, styles.InactivePaneStyle
	switch m.focusedPane {
	case PaneCollections:
		sStyle = styles.ActivePaneStyle
	case PaneRequest:
		rStyle = styles.ActivePaneStyle
	case PaneResponse:
		respStyle = styles.ActivePaneStyle
	}

	var sidebar string
	if m.sidebarVisible {
		sidebar = sStyle.Width(m.collections.Width).Height(m.collections.Height).Render(m.collections.View())
	}
	request := rStyle.Width(m.request.Width).Height(m.request.Height).Render(m.request.View())
	response := respStyle.Width(m.response.Width).Height(m.response.Height).Render(m.response.View())

	var dashboard string
	if m.sidebarVisible {
		dashboard = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, request, response)
	} else {
		dashboard = lipgloss.JoinHorizontal(lipgloss.Top, request, response)
	}
	
	// Apply Main Layout padding
	return styles.MainLayoutStyle.Render(dashboard)
}

func (m Model) viewStatusBar() string {
	logo := styles.StatusBarLogoStyle.Render(" TAPI ")

	// Mode indicator
	var modeIndicator string
	if m.leaderActive {
		modeIndicator = lipgloss.NewStyle().Background(lipgloss.Color("#e0af68")).Foreground(lipgloss.Color("#1a1b26")).Bold(true).Padding(0, 1).Render("LEADER")
	} else if m.mode == ModeInsert {
		modeIndicator = lipgloss.NewStyle().Background(lipgloss.Color("#9ece6a")).Foreground(lipgloss.Color("#1a1b26")).Bold(true).Padding(0, 1).Render("INSERT")
	} else {
		modeIndicator = lipgloss.NewStyle().Background(lipgloss.Color("#7aa2f7")).Foreground(lipgloss.Color("#1a1b26")).Bold(true).Padding(0, 1).Render("NORMAL")
	}

	ctx := " No Env "
	if m.currentEnv != nil {
		ctx = " " + m.currentEnv.Name + " "
	}
	contextBlock := styles.StatusBarContextStyle.Render(ctx)
	
	helpView := m.help.View(m.keys)
	helpBlock := styles.StatusBarInfoStyle.Render(helpView)
	
	wSoFar := lipgloss.Width(logo) + lipgloss.Width(contextBlock) + lipgloss.Width(helpBlock)
	statusWidth := max(0, m.Width - wSoFar - 4) // Adjust for padding
	
	statusText := m.statusText
	if statusText == "" {
		statusText = "Ready"
	}
	
	statusStyle := styles.StatusBarInfoStyle.Width(statusWidth)
	if m.statusIsErr {
		statusStyle = statusStyle.Background(styles.ErrorColor).Foreground(styles.White)
	}
	statusBlock := statusStyle.Render(statusText)
	
	// Persistent error indicator
	var errorBlock string
	if len(m.loadErrors) > 0 {
		errorBlock = styles.StatusBarInfoStyle.
			Background(styles.ErrorColor).
			Foreground(styles.White).
			Padding(0, 1).
			Render(fmt.Sprintf("! %d Errors", len(m.loadErrors)))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		modeIndicator,
		contextBlock,
		statusBlock,
		errorBlock,
		helpBlock,
	)
}




