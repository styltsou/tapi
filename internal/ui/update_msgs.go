package ui

import (
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
	"github.com/styltsou/tapi/internal/storage/exporter"
	"github.com/styltsou/tapi/internal/storage/importer"
	"github.com/styltsou/tapi/internal/ui/commands"
	"github.com/styltsou/tapi/internal/ui/components"
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
)

func (m Model) handleAppMsg(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		// Check minimum dimensions
		if m.Width < minWidth || m.Height < minHeight {
			m.tooSmall = true
			return m, nil, true
		}
		m.tooSmall = false

		// Dashboard layout calculation
		sidebarWidth := 0
		if m.sidebarVisible {
			sidebarWidth = 30
			if m.Width < 120 {
				sidebarWidth = 25
			}
		}
		
		remainingWidth := max(0, m.Width-sidebarWidth)
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
		return m, nil, true

	case uimsg.ErrMsg:
		logger.Logger.Error("Application error", "error", msg.Err)
		return m, commands.ShowStatusCmd("Error: "+msg.Err.Error(), true), true

	case uimsg.CopyAsCurlMsg:
		// Delegate to leader key handler logic
		req, targetedURL := m.request.BuildRequest()
		if targetedURL != "" {
			req.URL = targetedURL
		}
		curlCmd := exporter.ExportCurl(req, m.request.BaseURL)
		if err := clipboard.WriteAll(curlCmd); err != nil {
			return m, commands.ShowStatusCmd("Failed to copy cURL", true), true
		}
		return m, commands.ShowStatusCmd("cURL copied to clipboard", false), true

	case uimsg.ToggleSidebarMsg:
		m.sidebarVisible = !m.sidebarVisible
		if m.sidebarVisible {
			m.focusedPane = PaneCollections
		} else {
			m.focusedPane = PaneRequest
		}
		// Recalculate layout immediately
		newM, cmd := m.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height})
		return newM.(Model), cmd, true

	case uimsg.FocusRequestMsg:
		m.focusedPane = PaneRequest
		return m, nil, true

	case uimsg.FocusMsg:
		m.state = msg.Target
		if msg.Target == uimsg.ViewEnvEditor {
			if env, ok := msg.Data.(storage.Environment); ok {
				m.envEditor.SetEnvironment(env)
			}
		}
		return m, nil, true

	case uimsg.BackMsg:
		if m.currentCollection == nil {
			// No collection selected, go back to welcome
			m.state = uimsg.ViewWelcome
		} else {
			m.state = uimsg.ViewCollectionList
		}
		m.focusedPane = PaneCollections
		return m, nil, true

	case uimsg.ExecuteRequestMsg:
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
		return m, commands.ExecuteRequestCmd(m.httpClient, finalReq, msg.BaseURL), true

	case uimsg.ResponseReadyMsg:
		m.response.SetResponse(msg.Response, msg.Request)
		m.response.SetLoading(false)
		// Save response to current tab
		if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
			m.tabs[m.activeTab].Response = msg.Response
		}
		return m, nil, true

	case uimsg.CollectionSelectedMsg:
		m.state = uimsg.ViewCollectionList
		m.currentCollection = &msg.Collection
		m.collections.SetCollection(msg.Collection)
		m.focusedPane = PaneRequest
		newM, cmd := m.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height}) 
		return newM.(Model), cmd, true

	case uimsg.CollectionsLoadedMsg:
		m.welcome.SetCollections(msg.Collections)
		m.loadErrors = msg.LoadErrors
		if len(m.loadErrors) > 0 {
			m.statusText = fmt.Sprintf("Warning: %d collection(s) failed to load", len(m.loadErrors))
			m.statusIsErr = true
		}
		// Also forward to collectionSelector
		newSelector, selCmd := m.collectionSelector.Update(msg)
		m.collectionSelector = newSelector
		return m, selCmd, true

	case uimsg.RequestSelectedMsg:
		// Check if this request is already open in a tab
		for i, tab := range m.tabs {
			if tab.Request.Name == msg.Request.Name && tab.BaseURL == msg.BaseURL {
				// Switch to existing tab
				m.saveCurrentTab()
				m.activeTab = i
				m.loadActiveTab()
				m.focusedPane = PaneRequest
				return m, nil, true
			}
		}
		// Save current tab state before opening new one
		m.saveCurrentTab()
		// Open new tab
		label := msg.Request.Method + " " + msg.Request.Name
		if len(label) > 20 {
			label = label[:20] + "…"
		}
		m.tabs = append(m.tabs, components.RequestTab{
			Request:  msg.Request,
			BaseURL:  msg.BaseURL,
			Response: nil,
			Label:    label,
		})
		m.activeTab = len(m.tabs) - 1
		m.request.LoadRequest(msg.Request, msg.BaseURL)
		// Clear response for new tab
		w, h := m.response.Width, m.response.Height
		m.response = components.NewResponseModel()
		m.response.SetSize(w, h)
		m.focusedPane = PaneRequest
		return m, nil, true

	case uimsg.CreateEnvMsg:
		newEnv := storage.Environment{Name: msg.Name, Variables: make(map[string]string)}
		m.state = uimsg.ViewEnvEditor
		m.envEditor.SetEnvironment(newEnv)
		return m, nil, true

	case uimsg.DeleteEnvMsg:
		return m, func() tea.Msg {
			return uimsg.ConfirmActionMsg{
				Title:     "Delete environment: " + msg.Name + "?",
				OnConfirm: confirmedDeleteEnvMsg{Name: msg.Name},
			}
		}, true

	case confirmedDeleteEnvMsg:
		return m, commands.DeleteEnvCmd(msg.Name), true

	case uimsg.PromptForInputMsg:
		m.state = uimsg.ViewInput
		m.input.Title = msg.Title
		m.input.TextInput.Placeholder = msg.Placeholder
		m.input.TextInput.SetValue("")
		m.input.OnCommitMsg = msg.OnCommit
		m.input.OnCancelMsg = func() tea.Msg { return uimsg.BackMsg{} }
		// m.input.Init() returns a Cmd, usually Blink
		return m, m.input.Init(), true

	case uimsg.CreateRequestMsg:
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
		return m, commands.CreateRequestCmd(targetCol, newReq), true

	case uimsg.DeleteRequestMsg:
		return m, commands.DeleteRequestCmd(msg.CollectionName, msg.RequestName), true

	case uimsg.DeleteCollectionMsg:
		return m, commands.DeleteCollectionCmd(msg.Name), true

	case uimsg.CreateCollectionMsg:
		col := storage.Collection{Name: msg.Name, Requests: []storage.Request{}}
		if err := storage.SaveCollection(col); err != nil {
			m.statusText = "Error: " + err.Error()
			m.statusIsErr = true
			return m, nil, true
		}
		// Select the newly created collection and go to workspace
		return m, func() tea.Msg {
			return uimsg.CollectionSelectedMsg{Collection: col}
		}, true

	case uimsg.RenameCollectionMsg:
		return m, commands.RenameCollectionCmd(msg.OldName, msg.NewName), true

	case uimsg.EnvChangedMsg:
		m.currentEnv = &msg.NewEnv
		m.request.SetVariables(msg.NewEnv.Variables)
		return m, commands.ShowStatusCmd(fmt.Sprintf("Env: %s", msg.NewEnv.Name), false), true

	case uimsg.StatusMsg:
		m.statusText = msg.Message
		m.statusIsErr = msg.IsError
		// Auto-dismiss status after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return uimsg.ClearStatusMsg{}
		}), true

	case uimsg.ClearStatusMsg:
		m.statusText = ""
		m.statusIsErr = false
		return m, nil, true

	case uimsg.ImportCollectionMsg:
		importPath := expandTilde(msg.Path)
		collections, err := importer.ImportFromFile(importPath)
		if err != nil {
			return m, commands.ShowStatusCmd("Import failed: "+err.Error(), true), true
		}
		for _, col := range collections {
			if saveErr := storage.SaveCollection(col); saveErr != nil {
				return m, commands.ShowStatusCmd("Import save failed: "+saveErr.Error(), true), true
			}
		}
		// Reload and select the first imported collection
		if len(collections) > 0 {
			return m, tea.Batch(
				commands.ShowStatusCmd(fmt.Sprintf("Imported %d collection(s)", len(collections)), false),
				commands.LoadCollectionsCmd(),
				func() tea.Msg {
					return uimsg.CollectionSelectedMsg{Collection: collections[0]}
				},
			), true
		}
		return m, commands.ShowStatusCmd("Import successful", false), true

	case uimsg.ExportCollectionMsg:
		if m.currentCollection == nil {
			return m, commands.ShowStatusCmd("No collection selected to export", true), true
		}
		exportPath := expandTilde(msg.DestPath)
		if err := storage.ExportCollection(m.currentCollection.Name, exportPath); err != nil {
			return m, commands.ShowStatusCmd("Export failed: "+err.Error(), true), true
		}
		return m, commands.ShowStatusCmd("Exported to "+exportPath, false), true

	case uimsg.ConfirmActionMsg:
		m.state = uimsg.ViewInput
		confirmAction := msg.OnConfirm
		m.input = components.NewInputModel(
			msg.Title,
			"Type 'yes' to confirm",
			func(val string) tea.Msg {
				if val == "yes" {
					return confirmAction
				}
				return uimsg.BackMsg{}
			},
			func() tea.Msg { return uimsg.BackMsg{} },
		)
		m.input.SetSize(m.Width, m.Height)
		return m, m.input.Init(), true

	case uimsg.DuplicateRequestMsg:
		return m, commands.DuplicateRequestCmd(msg.CollectionName, msg.RequestName), true

	case uimsg.RequestSaveResponseMsg:
		m.state = uimsg.ViewInput
		m.input = components.NewInputModel(
			"Save Response Body",
			"filename.json",
			func(val string) tea.Msg {
				return uimsg.SaveResponseBodyMsg{Filename: val, Body: msg.Body}
			},
			func() tea.Msg { return uimsg.BackMsg{} },
		)
		m.input.SetSize(m.Width, m.Height)
		return m, m.input.Init(), true

	case uimsg.SaveResponseBodyMsg:
		m.state = uimsg.ViewCollectionList
		m.focusedPane = PaneResponse
		return m, commands.SaveResponseBodyCmd(expandTilde(msg.Filename), msg.Body), true
	}

	return m, nil, false
}
