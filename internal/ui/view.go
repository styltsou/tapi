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

	// Determine the base view (Welcome screen vs Dashboard)
	var baseView string
	if m.state == uimsg.ViewWelcome || (m.state == uimsg.ViewInput && m.currentCollection == nil) {
		baseView = m.welcome.View()
	} else {
		// 1. Header
		header := m.viewHeader()

		// Tab Bar
		tabBar := m.viewTabBar()

		// 2. Dashboard Content
		dashboard := m.viewDashboard()

		// 3. Status Bar
		bar := m.viewStatusBar()

		// Final Assembly (Base UI)
		baseView = lipgloss.JoinVertical(lipgloss.Left,
			header,
			tabBar,
			dashboard,
			bar,
		)
	}

	// Overlays (Modals, Menus, Command Palette)
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
	} else if m.mode == ModeCommand {
		overlay = styles.CommandPaletteStyle.
			Width(min(60, m.Width-4)).
			Render(m.commandInput.View())
	} else if m.state == uimsg.ViewInput {
		overlay = m.input.View()
	}

	if overlay != "" {
		// Calculate position
		overlayW := lipgloss.Width(overlay)
		overlayH := lipgloss.Height(overlay)
		
		x := (m.Width - overlayW) / 2
		y := (m.Height - overlayH) / 2
		
		// If it's command mode, place it slightly lower
		if m.mode == ModeCommand {
			y = m.Height - overlayH - 1
		}
		
		baseView = placeOverlay(baseView, overlay, x, y)
	}

	// Render notifications in top-right corner
	if len(m.notifications) > 0 {
		var notifStr string
		for i, n := range m.notifications {
			style := styles.NotificationStyle
			if !n.IsError {
				style = style.Background(styles.PrimaryColor)
			}
			notifStr += style.Render(n.Message)
			if i < len(m.notifications)-1 {
				notifStr += "\n"
			}
		}

		notifW := lipgloss.Width(notifStr)
		x := m.Width - notifW - 1
		y := 1 // Top margin
		baseView = placeOverlay(baseView, notifStr, x, y)
	}

	return baseView
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
	// Determine styles based on focus
	sStyle := styles.SidebarStyle
	rStyle := styles.RequestPaneStyle
	respStyle := styles.ResponsePaneStyle

	if m.focusedPane == PaneCollections {
		sStyle = sStyle.BorderForeground(styles.ActiveBorderColor)
	} else {
		sStyle = sStyle.BorderForeground(styles.InactiveBorderColor)
	}

	if m.focusedPane == PaneRequest {
		rStyle = rStyle.BorderForeground(styles.ActiveBorderColor)
	} else {
		rStyle = rStyle.BorderForeground(styles.InactiveBorderColor)
	}

	// Response pane might not have a visible border in this layout (except maybe top if we added it),
	// or we just highlight the header.
	// For now, let's keep it simple.

	var sidebar string
	if m.sidebarVisible {
		sidebar = sStyle.Width(m.collections.Width).Height(m.collections.Height).Render(m.collections.View())
	}
	requestView := m.request.View()
	responseView := m.response.View()

	// Add headers
	// For headers to look right with partial borders, they should probably span the full width
	// minus the border width.
	
	reqHeader := styles.PaneHeaderStyle.Width(m.request.Width - 2).Render("REQUEST") 
	respHeader := styles.PaneHeaderStyle.Width(m.response.Width - 2).Render("RESPONSE")

	request := rStyle.Width(m.request.Width).Height(m.request.Height).Render(
		lipgloss.JoinVertical(lipgloss.Left, reqHeader, requestView),
	)
	response := respStyle.Width(m.response.Width).Height(m.response.Height).Render(
		lipgloss.JoinVertical(lipgloss.Left, respHeader, responseView),
	)

	var dashboard string
	
	// Stack Request and Response vertically
	content := lipgloss.JoinVertical(lipgloss.Left, request, response)

	if m.sidebarVisible {
		dashboard = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)
	} else {
		dashboard = content
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
	
	// Persistent error indicator
	var errorBlock string
	if len(m.loadErrors) > 0 {
		errorBlock = styles.StatusBarInfoStyle.
			Background(styles.ErrorColor).
			Foreground(styles.White).
			Padding(0, 1).
			Render(fmt.Sprintf("! %d Errors", len(m.loadErrors)))
	}

	helpView := m.help.View(m.keys)
	helpBlock := styles.StatusBarInfoStyle.Render(helpView)

	wSoFar := lipgloss.Width(logo) + lipgloss.Width(modeIndicator) + lipgloss.Width(contextBlock) + lipgloss.Width(helpBlock) + lipgloss.Width(errorBlock)
	statusWidth := max(0, m.Width - wSoFar)
	
	statusText := m.statusText
	if statusText == "" {
		statusText = "Ready"
	}
	
	statusStyle := styles.StatusBarInfoStyle.Width(statusWidth)
	if m.statusIsErr {
		statusStyle = statusStyle.Background(styles.ErrorColor).Foreground(styles.White)
	}
	statusBlock := statusStyle.Render(statusText)

	return lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		modeIndicator,
		contextBlock,
		statusBlock,
		errorBlock,
		helpBlock,
	)
}
