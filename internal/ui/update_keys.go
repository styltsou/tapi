package ui

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/storage/exporter"
	"github.com/styltsou/tapi/internal/ui/commands"
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
)

func (m Model) handleKeyMsg(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	// Always allow Ctrl+C to quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit, true
	}

	// If help overlay is open, route to it
	if m.helpOverlay.Visible {
		newHelp, helpCmd := m.helpOverlay.Update(msg)
		m.helpOverlay = newHelp
		return m, helpCmd, true
	}

	// If a modal is open, route to it
	if m.menu.Visible || m.env.Visible || m.state == uimsg.ViewEnvEditor || m.collectionSelector.Visible || m.state == uimsg.ViewInput {
		// Let modals handle their own keys (handled below in routing section of generic Update)
		return m, nil, false
	}

	// Welcome screen has its own key handling
	if m.state == uimsg.ViewWelcome {
		newWelcome, welcomeCmd := m.welcome.Update(msg)
		m.welcome = newWelcome
		return m, welcomeCmd, true
	}

	// --- Leader Key Handling (Normal mode only) ---
	if m.mode == ModeNormal && m.leaderActive {
		m.leaderActive = false
		switch msg.String() {
		case "e":
			m.sidebarVisible = !m.sidebarVisible
			if m.sidebarVisible {
				m.focusedPane = PaneCollections
			} else {
				m.focusedPane = PaneRequest
			}
			newM, cmd := m.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height})
			return newM.(Model), cmd, true
		case "c":
			m.collectionSelector.Visible = true
			return m, nil, true
		case "v":
			m.env.Visible = !m.env.Visible
			return m, nil, true
		case "r":
			// Trigger request execution from request model
			req, targetedURL := m.request.BuildRequest()
			return m, func() tea.Msg {
				return uimsg.ExecuteRequestMsg{Request: req, BaseURL: m.request.BaseURL, TargetedURL: targetedURL}
			}, true
		case "s":
			req, _ := m.request.BuildRequest()
			return m, func() tea.Msg {
				return uimsg.SaveRequestMsg{Request: req}
			}, true
		case "p":
			m.focusedPane = PaneRequest
			return m, nil, true
		case "o":
			m.request.Preview = !m.request.Preview
			return m, nil, true
		case "k":
			m.menu.Visible = true
			m.env.Visible = false
			m.collectionSelector.Visible = false
			return m, nil, true
		case "y":
			// Copy as cURL
			req, targetedURL := m.request.BuildRequest()
			if targetedURL != "" {
				req.URL = targetedURL
			}
			curlCmd := exporter.ExportCurl(req, m.request.BaseURL)
			err := clipboard.WriteAll(curlCmd)
			if err != nil {
				return m, commands.ShowStatusCmd("Failed to copy cURL", true), true
			}
			return m, commands.ShowStatusCmd("cURL copied to clipboard", false), true
		case "w":
			// Close current tab
			if len(m.tabs) > 0 {
				m.closeTab(m.activeTab)
			}
			return m, nil, true
		case "q":
			return m, tea.Quit, true
		default:
			// Unknown chord, ignore
			return m, nil, true
		}
	}

	// --- Normal Mode ---
	if m.mode == ModeNormal {
		switch msg.String() {
		case " ":
			m.leaderActive = true
			m.gPending = false
			return m, nil, true
		case "?":
			m.helpOverlay.Toggle()
			return m, nil, true
		case "g":
			if !m.gPending {
				m.gPending = true
				return m, nil, true
			}
			// gg — ignore double g
			m.gPending = false
			return m, nil, true
		case "t":
			if m.gPending {
				// gt — next tab
				m.gPending = false
				if len(m.tabs) > 1 {
					m.saveCurrentTab()
					m.activeTab = (m.activeTab + 1) % len(m.tabs)
					m.loadActiveTab()
				}
				return m, nil, true
			}
		case "T":
			if m.gPending {
				// gT — prev tab
				m.gPending = false
				if len(m.tabs) > 1 {
					m.saveCurrentTab()
					m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
					m.loadActiveTab()
				}
				return m, nil, true
			}
		case "i", "enter":
			// Enter Insert mode
			m.mode = ModeInsert
			return m, nil, true
		case "tab":
			// Cycle focus between panes
			if m.sidebarVisible {
				switch m.focusedPane {
				case PaneCollections:
					m.focusedPane = PaneRequest
				case PaneRequest:
					m.focusedPane = PaneResponse
				case PaneResponse:
					m.focusedPane = PaneCollections
				}
			} else {
				if m.focusedPane == PaneRequest {
					m.focusedPane = PaneResponse
				} else {
					m.focusedPane = PaneRequest
				}
			}
			return m, nil, true
		case "shift+tab":
			if m.sidebarVisible {
				switch m.focusedPane {
				case PaneCollections:
					m.focusedPane = PaneResponse
				case PaneRequest:
					m.focusedPane = PaneCollections
				case PaneResponse:
					m.focusedPane = PaneRequest
				}
			} else {
				if m.focusedPane == PaneRequest {
					m.focusedPane = PaneResponse
				} else {
					m.focusedPane = PaneRequest
				}
			}
			return m, nil, true
		}
		// In Normal mode, route j/k/arrows etc. to focused pane for navigation
		// This happens in the main update loop's final routing section
	}

	// --- Insert Mode ---
	if m.mode == ModeInsert {
		if msg.String() == "esc" {
			m.mode = ModeNormal
			return m, nil, true
		}
		// Route all other keys to focused sub-model for text editing
		// This happens in the main update loop's final routing section
	}
	
	return m, nil, false
}
