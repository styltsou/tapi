// Package ui
package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/http"

	"github.com/styltsou/tapi/internal/storage"
)

// ViewState represents the different views in the TUI
type ViewState int

const (
	ViewWelcome ViewState = iota
	ViewCollectionList
	ViewRequestBuilder
	ViewResponse
	ViewEnvironments
	ViewEnvEditor
	ViewInput
	ViewHelp
)

// String returns the string representation of ViewState
func (v ViewState) String() string {
	switch v {
	case ViewWelcome:
		return "Welcome"
	case ViewCollectionList:
		return "Collection List"
	case ViewRequestBuilder:
		return "Request Builder"
	case ViewResponse:
		return "Response"
	case ViewEnvironments:
		return "Environments"
	case ViewEnvEditor:
		return "Environment Editor"
	case ViewHelp:
		return "Help"
	default:
		return "Unknown"
	}
}

// ========================================
// Error Messages
// ========================================

// ErrMsg is a global error carrier for handling errors across the application
type ErrMsg struct {
	Err error
}

// ========================================
// Request Execution Messages
// ========================================

// ExecuteRequestMsg is sent by RequestBuilder to MainModel to execute a request
type ExecuteRequestMsg struct {
	Request     storage.Request
	BaseURL     string // Base URL from the collection
	TargetedURL string // The actual URL to execute (after substitutions)
}

// ResponseReadyMsg is sent by HTTP Client to MainModel when a response is received
type ResponseReadyMsg struct {
	Response *http.ProcessedResponse
	Request  storage.Request // Include original request for context
}

// ========================================
// Storage Messages
// ========================================

// SaveRequestMsg is sent when user presses Ctrl+S to save a request
type SaveRequestMsg struct {
	Request storage.Request
}

// SaveCollectionMsg is sent to save an entire collection
type SaveCollectionMsg struct {
	Collection storage.Collection
}

// DeleteRequestMsg is sent to delete a request
type DeleteRequestMsg struct {
	CollectionName string
	RequestName    string
}

// ========================================
// Environment Messages
// ========================================

// EnvChangedMsg is sent when user selects a new environment
type EnvChangedMsg struct {
	NewEnv storage.Environment
}

// CreateEnvMsg is sent when user wants to create a new environment
type CreateEnvMsg struct {
	Name string
}

// DeleteEnvMsg is sent when user wants to delete an environment
type DeleteEnvMsg struct {
	Name string
}

// Request Management Messages
type CreateRequestMsg struct {
	Name string
}

// DeleteRequestMsg is likely already defined above.
// Collection Management Messages
type CreateCollectionMsg struct {
	Name string
}	

type DeleteCollectionMsg struct {
	Name string
}

type RenameCollectionMsg struct {
	OldName string
	NewName string
}

// PromptForInputMsg triggers the input modal
type PromptForInputMsg struct {
	Title       string
	Placeholder string
	OnCommit    func(string) tea.Msg
}

// ========================================
// Navigation Messages
// ========================================

// FocusMsg is sent to switch between main views
type FocusMsg struct {
	Target ViewState
	Data   interface{}
}

// BackMsg is sent to go back to the previous view
type BackMsg struct{}

// ========================================
// Collection Messages
// ========================================

// CollectionSelectedMsg is sent when a collection is selected from the list
type CollectionSelectedMsg struct {
	Collection storage.Collection
}

// CollectionsLoadedMsg is sent when collections are finished loading from storage
type CollectionsLoadedMsg struct {
	Collections []storage.Collection
}

// RequestSelectedMsg is sent when a request is selected from a collection
type RequestSelectedMsg struct {
	Request storage.Request
	BaseURL string // Base URL from parent collection
}

// ========================================
// UI Update Messages
// ========================================

// StatusMsg is sent to update the status bar
type StatusMsg struct {
	Message string
	IsError bool
}

// EnvsLoadedMsg is sent when environments are loaded
type EnvsLoadedMsg struct {
	Envs []storage.Environment
}

// LoadingMsg is sent to show/hide loading indicator
type LoadingMsg struct {
	IsLoading bool
	Message   string
}
// ToggleSidebarMsg toggles the visibility of the sidebar
type ToggleSidebarMsg struct{}

// FocusRequestMsg sets the focus to the request pane
type FocusRequestMsg struct{}

// SaveResponseBodyMsg contains the filename and body to save
type SaveResponseBodyMsg struct {
	Filename string
	Body     []byte
}

// RequestSaveResponseMsg is sent when user wants to save response body to file
type RequestSaveResponseMsg struct {
	Body []byte
}

// SuggestionSelectedMsg contains the selected variable name
type SuggestionSelectedMsg struct {
	VarName string
}

// ========================================
// Import/Export Messages
// ========================================

// ImportCollectionMsg triggers importing a collection from a file
type ImportCollectionMsg struct {
	Path string
}

// ExportCollectionMsg triggers exporting the current collection to a file
type ExportCollectionMsg struct {
	DestPath string
}

// ========================================
// Quick Win Messages
// ========================================

// ConfirmActionMsg shows a yes/no confirmation before executing a destructive action
type ConfirmActionMsg struct {
	Title     string
	OnConfirm tea.Msg
}

// DuplicateRequestMsg duplicates a request in a collection
type DuplicateRequestMsg struct {
	CollectionName string
	RequestName    string
}

// ClearStatusMsg clears the status bar text (used for auto-dismiss)
type ClearStatusMsg struct{}

// CopyAsCurlMsg triggers copying the current request as a cURL command
type CopyAsCurlMsg struct{}
