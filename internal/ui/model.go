package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
	"github.com/styltsou/tapi/internal/storage/exporter"
	"github.com/styltsou/tapi/internal/storage/importer"
)

// Pane represents the focusable areas of the dashboard
type Pane int

const (
	PaneCollections Pane = iota
	PaneRequest
	PaneResponse
)

const (
	minWidth  = 100
	minHeight = 20
)

// InputMode represents Normal or Insert mode (Vim-style)
type InputMode int

const (
	ModeNormal InputMode = iota
	ModeInsert
)



// Model is the main application model that manages state and sub-models
type Model struct {
	state       ViewState
	focusedPane Pane
	request     RequestModel
	response    ResponseModel
	collections CollectionsModel
	welcome     WelcomeModel
	env         EnvModel
	envEditor   EnvEditorModel
	input       InputModel
	menu        CommandMenuModel
	collectionSelector CollectionSelectorModel
	helpOverlay HelpOverlayModel
	help        help.Model
	keys        KeyMap
	width       int
	height      int
	tooSmall    bool
	sidebarVisible bool

	// Vim-style mode
	mode         InputMode
	leaderActive bool
	gPending     bool // for gt/gT vim combos

	// Request tabs
	tabs      []RequestTab
	activeTab int

	// HTTP client
	httpClient *http.Client
	cfg        config.Config

	// Current context
	currentCollection *storage.Collection
	currentEnv        *storage.Environment

	// Status line
	statusText  string
	statusIsErr bool
}

// NewModel creates a new main model with initial state
func NewModel(cfg config.Config) Model {
	ApplyTheme(cfg.Theme)

	return Model{
		state:       ViewWelcome,
		focusedPane: PaneCollections,
		keys:        DefaultKeyMap(),
		httpClient:  http.NewClient(),
		cfg:         cfg,
		sidebarVisible: true,
		mode:         ModeNormal,

		// Initialize sub-models
		request:     NewRequestModel(),
		response:    NewResponseModel(),
		collections: NewCollectionsModel(),
		welcome:     NewWelcomeModel(),
		env:         NewEnvModel(),
		envEditor:   NewEnvEditorModel(),
		input:       NewInputModel("", "", nil, nil),
		menu:        NewCommandMenuModel(),
		collectionSelector: NewCollectionSelectorModel(),
		helpOverlay: NewHelpOverlayModel(),
		help:        help.New(),
	}
}

// Init initializes the model and returns initial commands
func (m Model) Init() tea.Cmd {
	logger.Logger.Info("Initializing TUI")

	return tea.Batch(
		tea.EnterAltScreen,
		loadCollectionsCmd(),
		loadEnvsCmd(),
	)
}

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Check minimum dimensions
		if m.width < minWidth || m.height < minHeight {
			m.tooSmall = true
			return m, nil
		}
		m.tooSmall = false

		// Dashboard layout calculation
		sidebarWidth := 0
		if m.sidebarVisible {
			sidebarWidth = 30
			if m.width < 120 {
				sidebarWidth = 25
			}
		}
		
		remainingWidth := max(0, m.width-sidebarWidth)
		requestWidth := remainingWidth / 2
		responseWidth := max(0, remainingWidth-requestWidth)

		contentHeight := max(0, msg.Height-4) // header(1) + status(1) + help(1) + spacing(1)

		m.collections.SetSize(sidebarWidth, contentHeight)
		m.request.SetSize(requestWidth, contentHeight)
		m.response.SetSize(responseWidth, contentHeight)
		m.welcome.SetSize(msg.Width, msg.Height)
		
		m.env.SetSize(msg.Width, msg.Height)
		m.envEditor.SetSize(msg.Width, msg.Height)
		m.menu.SetSize(msg.Width, msg.Height)
		m.collectionSelector.SetSize(msg.Width, msg.Height)
		m.helpOverlay.SetSize(msg.Width, msg.Height)

		logger.Logger.Debug("Window resized", "width", msg.Width, "height", msg.Height)
		return m, nil

	case ErrMsg:
		logger.Logger.Error("Application error", "error", msg.Err)
		return m, showStatusCmd("Error: "+msg.Err.Error(), true)

	case tea.KeyMsg:
		// Always allow Ctrl+C to quit
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// If help overlay is open, route to it
		if m.helpOverlay.visible {
			newHelp, helpCmd := m.helpOverlay.Update(msg)
			m.helpOverlay = newHelp
			return m, helpCmd
		}

		// If a modal is open, route to it
		if m.menu.visible || m.env.visible || m.state == ViewEnvEditor || m.collectionSelector.visible || m.state == ViewInput {
			// Let modals handle their own keys (handled below in routing section)
			break
		}

		// Welcome screen has its own key handling
		if m.state == ViewWelcome {
			newWelcome, welcomeCmd := m.welcome.Update(msg)
			m.welcome = newWelcome
			return m, welcomeCmd
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
				return m.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			case "c":
				m.collectionSelector.visible = true
				return m, nil
			case "v":
				m.env.visible = !m.env.visible
				return m, nil
			case "r":
				// Trigger request execution from request model
				req, targetedURL := m.request.buildRequest()
				return m, func() tea.Msg {
					return ExecuteRequestMsg{Request: req, BaseURL: m.request.baseURL, TargetedURL: targetedURL}
				}
			case "s":
				req, _ := m.request.buildRequest()
				return m, func() tea.Msg {
					return SaveRequestMsg{Request: req}
				}
			case "p":
				m.focusedPane = PaneRequest
				return m, nil
			case "o":
				m.request.preview = !m.request.preview
				return m, nil
			case "k":
				m.menu.visible = true
				m.env.visible = false
				m.collectionSelector.visible = false
				return m, nil
			case "y":
				// Copy as cURL
				req, targetedURL := m.request.buildRequest()
				if targetedURL != "" {
					req.URL = targetedURL
				}
				curlCmd := exporter.ExportCurl(req, m.request.baseURL)
				err := clipboard.WriteAll(curlCmd)
				if err != nil {
					return m, showStatusCmd("Failed to copy cURL", true)
				}
				return m, showStatusCmd("cURL copied to clipboard", false)
			case "w":
				// Close current tab
				if len(m.tabs) > 0 {
					m.closeTab(m.activeTab)
				}
				return m, nil
			case "q":
				return m, tea.Quit
			default:
				// Unknown chord, ignore
				return m, nil
			}
		}

		// --- Normal Mode ---
		if m.mode == ModeNormal {
			switch msg.String() {
			case " ":
				m.leaderActive = true
				m.gPending = false
				return m, nil
			case "?":
				m.helpOverlay.Toggle()
				return m, nil
			case "g":
				if !m.gPending {
					m.gPending = true
					return m, nil
				}
				// gg — ignore double g
				m.gPending = false
				return m, nil
			case "t":
				if m.gPending {
					// gt — next tab
					m.gPending = false
					if len(m.tabs) > 1 {
						m.saveCurrentTab()
						m.activeTab = (m.activeTab + 1) % len(m.tabs)
						m.loadActiveTab()
					}
					return m, nil
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
					return m, nil
				}
			case "i", "enter":
				// Enter Insert mode
				m.mode = ModeInsert
				return m, nil
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
				return m, nil
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
				return m, nil
			}
			// In Normal mode, route j/k/arrows etc. to focused pane for navigation
		}

		// --- Insert Mode ---
		if m.mode == ModeInsert {
			if msg.String() == "esc" {
				m.mode = ModeNormal
				return m, nil
			}
			// Route all other keys to focused sub-model for text editing
		}
	
	case CopyAsCurlMsg:
		// Delegate to leader key handler logic
		req, targetedURL := m.request.buildRequest()
		if targetedURL != "" {
			req.URL = targetedURL
		}
		curlCmd := exporter.ExportCurl(req, m.request.baseURL)
		if err := clipboard.WriteAll(curlCmd); err != nil {
			return m, showStatusCmd("Failed to copy cURL", true)
		}
		return m, showStatusCmd("cURL copied to clipboard", false)

	case ToggleSidebarMsg:
		m.sidebarVisible = !m.sidebarVisible
		if m.sidebarVisible {
			m.focusedPane = PaneCollections
		} else {
			m.focusedPane = PaneRequest
		}
		// Recalculate layout immediately
		return m.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	case FocusRequestMsg:
		m.focusedPane = PaneRequest
		return m, nil

	case FocusMsg:
		m.state = msg.Target
		if msg.Target == ViewEnvEditor {
			if env, ok := msg.Data.(storage.Environment); ok {
				m.envEditor.SetEnvironment(env)
			}
		}
		return m, nil

	case BackMsg:
		if m.currentCollection == nil {
			// No collection selected, go back to welcome
			m.state = ViewWelcome
		} else {
			m.state = ViewCollectionList
		}
		m.focusedPane = PaneCollections
		return m, nil

	case ExecuteRequestMsg:
		m.focusedPane = PaneResponse
		m.response.SetLoading(true)
		finalReq := m.applyCurrentEnv(msg.Request)
		// Inject default headers from config (don't override request-specific headers)
		for k, v := range m.cfg.DefaultHeaders {
			if _, exists := finalReq.Headers[k]; !exists {
				if finalReq.Headers == nil {
					finalReq.Headers = make(map[string]string)
				}
				finalReq.Headers[k] = v
			}
		}
		// If TargetedURL is present, we temporarily override the URL in the request
		// or pass it explicitly.
		if msg.TargetedURL != "" {
			finalReq.URL = msg.TargetedURL
		}
		return m, executeRequestCmd(m.httpClient, finalReq, msg.BaseURL)

	case ResponseReadyMsg:
		m.response.SetResponse(msg.Response, msg.Request)
		m.response.SetLoading(false)
		// Save response to current tab
		if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
			m.tabs[m.activeTab].Response = msg.Response
		}
		return m, nil

	case CollectionSelectedMsg:
		m.state = ViewCollectionList
		m.currentCollection = &msg.Collection
		m.collections.SetCollection(msg.Collection)
		m.focusedPane = PaneRequest
		return m.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

	case CollectionsLoadedMsg:
		m.welcome.SetCollections(msg.Collections)
		// Also forward to collectionSelector
		newSelector, selCmd := m.collectionSelector.Update(msg)
		m.collectionSelector = newSelector
		return m, selCmd

	case RequestSelectedMsg:
		// Check if this request is already open in a tab
		for i, tab := range m.tabs {
			if tab.Request.Name == msg.Request.Name && tab.BaseURL == msg.BaseURL {
				// Switch to existing tab
				m.saveCurrentTab()
				m.activeTab = i
				m.loadActiveTab()
				m.focusedPane = PaneRequest
				return m, nil
			}
		}
		// Save current tab state before opening new one
		m.saveCurrentTab()
		// Open new tab
		label := msg.Request.Method + " " + msg.Request.Name
		if len(label) > 20 {
			label = label[:20] + "…"
		}
		m.tabs = append(m.tabs, RequestTab{
			Request:  msg.Request,
			BaseURL:  msg.BaseURL,
			Response: nil,
			Label:    label,
		})
		m.activeTab = len(m.tabs) - 1
		m.request.LoadRequest(msg.Request, msg.BaseURL)
		// Clear response for new tab
		w, h := m.response.width, m.response.height
		m.response = NewResponseModel()
		m.response.SetSize(w, h)
		m.focusedPane = PaneRequest
		return m, nil

	case CreateEnvMsg:
		newEnv := storage.Environment{Name: msg.Name, Variables: make(map[string]string)}
		m.state = ViewEnvEditor
		m.envEditor.SetEnvironment(newEnv)
		return m, nil

	case DeleteEnvMsg:
		return m, func() tea.Msg {
			return ConfirmActionMsg{
				Title:     "Delete environment: " + msg.Name + "?",
				OnConfirm: confirmedDeleteEnvMsg{Name: msg.Name},
			}
		}

	case confirmedDeleteEnvMsg:
		return m, deleteEnvCmd(msg.Name)

	case PromptForInputMsg:
		m.state = ViewInput
		m.input.title = msg.Title
		m.input.textInput.Placeholder = msg.Placeholder
		m.input.textInput.SetValue("")
		m.input.onCommitMsg = msg.OnCommit
		m.input.onCancelMsg = func() tea.Msg { return BackMsg{} }
		return m, m.input.Init()

	case CreateRequestMsg:
		targetCol := "My Collection"
		if m.currentCollection != nil {
			targetCol = m.currentCollection.Name
		}
		// Create a basic GET request by default
		newReq := storage.Request{
			Name:    msg.Name,
			Method:  "GET",
			URL:     "https://httpbin.org/get",
			Headers: make(map[string]string),
		}
		return m, createRequestCmd(targetCol, newReq)

	case DeleteRequestMsg:
		return m, deleteRequestCmd(msg.CollectionName, msg.RequestName)

	case DeleteCollectionMsg:
		return m, deleteCollectionCmd(msg.Name)

	case CreateCollectionMsg:
		col := storage.Collection{Name: msg.Name, Requests: []storage.Request{}}
		if err := storage.SaveCollection(col); err != nil {
			m.statusText = "Error: " + err.Error()
			m.statusIsErr = true
			return m, nil
		}
		// Select the newly created collection and go to workspace
		return m, func() tea.Msg {
			return CollectionSelectedMsg{Collection: col}
		}

	case RenameCollectionMsg:
		return m, renameCollectionCmd(msg.OldName, msg.NewName)

	case EnvChangedMsg:
		m.currentEnv = &msg.NewEnv
		m.request.SetVariables(msg.NewEnv.Variables)
		return m, showStatusCmd(fmt.Sprintf("Env: %s", msg.NewEnv.Name), false)

	case StatusMsg:
		m.statusText = msg.Message
		m.statusIsErr = msg.IsError
		// Auto-dismiss status after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return ClearStatusMsg{}
		})

	case ClearStatusMsg:
		m.statusText = ""
		m.statusIsErr = false
		return m, nil

	case ImportCollectionMsg:
		importPath := expandTilde(msg.Path)
		collections, err := importer.ImportFromFile(importPath)
		if err != nil {
			return m, showStatusCmd("Import failed: "+err.Error(), true)
		}
		for _, col := range collections {
			if saveErr := storage.SaveCollection(col); saveErr != nil {
				return m, showStatusCmd("Import save failed: "+saveErr.Error(), true)
			}
		}
		// Reload and select the first imported collection
		if len(collections) > 0 {
			return m, tea.Batch(
				showStatusCmd(fmt.Sprintf("Imported %d collection(s)", len(collections)), false),
				loadCollectionsCmd(),
				func() tea.Msg {
					return CollectionSelectedMsg{Collection: collections[0]}
				},
			)
		}
		return m, showStatusCmd("Import successful", false)

	case ExportCollectionMsg:
		if m.currentCollection == nil {
			return m, showStatusCmd("No collection selected to export", true)
		}
		exportPath := expandTilde(msg.DestPath)
		if err := storage.ExportCollection(m.currentCollection.Name, exportPath); err != nil {
			return m, showStatusCmd("Export failed: "+err.Error(), true)
		}
		return m, showStatusCmd("Exported to "+exportPath, false)

	case ConfirmActionMsg:
		m.state = ViewInput
		confirmAction := msg.OnConfirm
		m.input = NewInputModel(
			msg.Title,
			"Type 'yes' to confirm",
			func(val string) tea.Msg {
				if val == "yes" {
					return confirmAction
				}
				return BackMsg{}
			},
			func() tea.Msg { return BackMsg{} },
		)
		m.input.SetSize(m.width, m.height)
		return m, m.input.Init()

	case DuplicateRequestMsg:
		return m, duplicateRequestCmd(msg.CollectionName, msg.RequestName)

	case RequestSaveResponseMsg:
		m.state = ViewInput
		m.input = NewInputModel(
			"Save Response Body",
			"filename.json",
			func(val string) tea.Msg {
				return SaveResponseBodyMsg{Filename: val, Body: msg.Body}
			},
			func() tea.Msg { return BackMsg{} },
		)
		m.input.SetSize(m.width, m.height)
		return m, m.input.Init()

	case SaveResponseBodyMsg:
		m.state = ViewCollectionList
		m.focusedPane = PaneResponse
		return m, saveResponseBodyCmd(expandTilde(msg.Filename), msg.Body)
	}

	// Route updates based on focus and state
	if m.menu.visible {
		newMenu, menuCmd := m.menu.Update(msg)
		m.menu = newMenu
		cmds = append(cmds, menuCmd)
	} else if m.env.visible {
		newEnv, envCmd := m.env.Update(msg)
		m.env = newEnv
		cmds = append(cmds, envCmd)
	} else if m.state == ViewEnvEditor {
		newEditor, editCmd := m.envEditor.Update(msg)
		m.envEditor = newEditor
		cmds = append(cmds, editCmd)
	} else if m.collectionSelector.visible {
		newSelector, selCmd := m.collectionSelector.Update(msg)
		m.collectionSelector = newSelector
		cmds = append(cmds, selCmd)
	} else if m.state == ViewInput {
		newInput, inputCmd := m.input.Update(msg)
		m.input = newInput
		cmds = append(cmds, inputCmd)
	} else {
		// Dashboard updates
		switch m.focusedPane {
		case PaneCollections:
			newCollections, colCmd := m.collections.Update(msg)
			m.collections = newCollections
			cmds = append(cmds, colCmd)
		case PaneRequest:
			newRequest, reqCmd := m.request.Update(msg)
			m.request = newRequest
			cmds = append(cmds, reqCmd)
		case PaneResponse:
			newResponse, respCmd := m.response.Update(msg)
			m.response = newResponse
			cmds = append(cmds, respCmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Check if terminal is too small
	if m.tooSmall {
		return lipgloss.Place(m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			fmt.Sprintf("Terminal is too small.\nPlease resize to at least %dx%d.", minWidth, minHeight),
		)
	}

	// Welcome screen
	if m.state == ViewWelcome {
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
	if m.helpOverlay.visible {
		overlay = m.helpOverlay.View()
	} else if m.menu.visible {
		overlay = m.menu.View()
	} else if m.env.visible {
		overlay = m.env.View()
	} else if m.state == ViewEnvEditor {
		overlay = m.envEditor.View()
	} else if m.collectionSelector.visible {
		overlay = m.collectionSelector.View()
	}

	if overlay != "" {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceBackground(lipgloss.Color("#111111")),
		)
	}

	// Input prompt floats on top of the current view (transparent overlay)
	if m.state == ViewInput {
		inputOverlay := m.input.View()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, inputOverlay,
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
			lipgloss.WithWhitespaceChars("░"),
		)
	}

	return fullView
}

// applyCurrentEnv applies the current environment variables to a request
func (m *Model) applyCurrentEnv(req storage.Request) storage.Request {
	if m.currentEnv == nil {
		return req
	}

	req.URL = storage.Substitute(req.URL, m.currentEnv.Variables)
	req.Body = storage.Substitute(req.Body, m.currentEnv.Variables)
	for k, v := range req.Headers {
		req.Headers[k] = storage.Substitute(v, m.currentEnv.Variables)
	}
	return req
}

// confirmedDeleteEnvMsg is the internal message sent after user confirms env deletion
type confirmedDeleteEnvMsg struct {
	Name string
}

// expandTilde replaces a leading ~ with the user's home directory
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}

// ========================================
// View Helper Functions
// ========================================



func (m Model) viewHeader() string {
	headerText := " TAPI "
	if m.currentCollection != nil {
		headerText += " • " + m.currentCollection.Name
	}
	header := TitleStyle.Render(headerText)
	if m.currentEnv != nil {
		header += " " + StatusStyle.Render("Env: "+m.currentEnv.Name)
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
	sStyle, rStyle, respStyle := InactivePaneStyle, InactivePaneStyle, InactivePaneStyle
	switch m.focusedPane {
	case PaneCollections:
		sStyle = ActivePaneStyle
	case PaneRequest:
		rStyle = ActivePaneStyle
	case PaneResponse:
		respStyle = ActivePaneStyle
	}

	var sidebar string
	if m.sidebarVisible {
		sidebar = sStyle.Width(m.collections.width).Height(m.collections.height).Render(m.collections.View())
	}
	request := rStyle.Width(m.request.width).Height(m.request.height).Render(m.request.View())
	response := respStyle.Width(m.response.width).Height(m.response.height).Render(m.response.View())

	var dashboard string
	if m.sidebarVisible {
		dashboard = lipgloss.JoinHorizontal(lipgloss.Top, sidebar, request, response)
	} else {
		dashboard = lipgloss.JoinHorizontal(lipgloss.Top, request, response)
	}
	
	// Apply Main Layout padding
	return MainLayoutStyle.Render(dashboard)
}

func (m Model) viewStatusBar() string {
	logo := StatusBarLogoStyle.Render(" TAPI ")

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
	contextBlock := StatusBarContextStyle.Render(ctx)
	
	helpView := m.help.View(m.keys)
	helpBlock := StatusBarInfoStyle.Render(helpView)
	
	wSoFar := lipgloss.Width(logo) + lipgloss.Width(contextBlock) + lipgloss.Width(helpBlock)
	statusWidth := max(0, m.width - wSoFar - 4) // Adjust for padding
	
	statusText := m.statusText
	if statusText == "" {
		statusText = "Ready"
	}
	
	statusStyle := StatusBarInfoStyle.Width(statusWidth)
	if m.statusIsErr {
		statusStyle = statusStyle.Background(ErrorColor).Foreground(White)
	}
	statusBlock := statusStyle.Render(statusText)
	
	return lipgloss.JoinHorizontal(lipgloss.Top,
		logo,
		modeIndicator,
		contextBlock,
		statusBlock,
		helpBlock,
	)
}




