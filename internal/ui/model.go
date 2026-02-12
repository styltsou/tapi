package ui

import (
	"github.com/styltsou/tapi/internal/ui/commands"
	"github.com/styltsou/tapi/internal/ui/components"
	"github.com/styltsou/tapi/internal/ui/keys"
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
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
	state       uimsg.ViewState
	focusedPane Pane
	request     components.RequestModel
	response    components.ResponseModel
	collections components.CollectionsModel
	welcome     components.WelcomeModel
	env         components.EnvModel
	envEditor   components.EnvEditorModel
	input       components.InputModel
	menu        components.CommandMenuModel
	collectionSelector components.CollectionSelectorModel
	helpOverlay components.HelpOverlayModel
	help        help.Model
	keys        keys.KeyMap
	Width       int
	Height      int
	tooSmall    bool
	sidebarVisible bool

	// Vim-style mode
	mode         InputMode
	leaderActive bool
	gPending     bool // for gt/gT vim combos

	// Request tabs
	tabs      []components.RequestTab
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

	// Collection Load Errors
	loadErrors []error
}

// NewModel creates a new main model with initial state
func NewModel(cfg config.Config) Model {
	styles.ApplyTheme(cfg.Theme)

	return Model{
		state:       uimsg.ViewWelcome,
		focusedPane: PaneCollections,
		keys:        keys.DefaultKeyMap(),
		httpClient:  http.NewClient(),
		cfg:         cfg,
		sidebarVisible: true,
		mode:         ModeNormal,

		// Initialize sub-models
		request:     components.NewRequestModel(),
		response:    components.NewResponseModel(),
		collections: components.NewCollectionsModel(),
		welcome:     components.NewWelcomeModel(),
		env:         components.NewEnvModel(),
		envEditor:   components.NewEnvEditorModel(),
		input:       components.NewInputModel("", "", nil, nil),
		menu:        components.NewCommandMenuModel(),
		collectionSelector: components.NewCollectionSelectorModel(),
		helpOverlay: components.NewHelpOverlayModel(),
		help:        help.New(),
	}
}

// Init initializes the model and returns initial commands
func (m Model) Init() tea.Cmd {
	logger.Logger.Info("Initializing TUI")

	return tea.Batch(
		tea.EnterAltScreen,
		commands.LoadCollectionsCmd(),
		commands.LoadEnvsCmd(),
	)
}

