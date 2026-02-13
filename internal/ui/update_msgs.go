package ui

import (
	"fmt"
	"strings"
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

		// Stacked Layout: Request Top, Response Bottom
		// We give 50% height to each (minus header/status space)
		contentHeight := max(0, msg.Height-3) // header(1) + status(1) + tabs(1)
		
		requestHeight := contentHeight / 2
		responseHeight := max(0, contentHeight - requestHeight)
		

		// Both take full remaining width
		// We subtract 1 to account for the manual header we add in View()
		m.request.SetSize(remainingWidth, requestHeight-1)
		m.response.SetSize(remainingWidth, responseHeight-1)
		m.collections.SetSize(sidebarWidth, contentHeight)
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

	case uimsg.OpenCollectionSelectorMsg:
		m.collectionSelector.Visible = true
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
		col := msg.Collection
		m.currentCollection = &col
		m.collections.SetCollection(msg.Collection)
		m.focusedPane = PaneCollections
		newM, cmd := m.Update(tea.WindowSizeMsg{Width: m.Width, Height: m.Height}) 
		return newM.(Model), cmd, true

	case uimsg.CollectionsLoadedMsg:
		m.welcome.SetCollections(msg.Collections)
		m.loadErrors = msg.LoadErrors
		if len(m.loadErrors) > 0 {
			m.statusText = fmt.Sprintf("Warning: %d collection(s) failed to load", len(m.loadErrors))
			m.statusIsErr = true
		}
		
		// Refresh current collection if active
		if m.currentCollection != nil {
			found := false
			for i := range msg.Collections {
				if msg.Collections[i].Name == m.currentCollection.Name {
					m.currentCollection = &msg.Collections[i]
					m.collections.SetCollection(msg.Collections[i])
					found = true
					break
				}
			}
			// If current collection was deleted, go back to welcome/list
			if !found {
				m.currentCollection = nil
				m.state = uimsg.ViewWelcome
				m.focusedPane = PaneCollections
			}
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
		if strings.TrimSpace(msg.Name) == "" {
			return m, commands.ShowStatusCmd("Request name cannot be empty", true), true
		}
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
		// Close input modal and trigger creation
		m.state = uimsg.ViewCollectionList
		return m, commands.CreateRequestCmd(targetCol, newReq), true
	case uimsg.RequestCreatedMsg:
		return m, tea.Batch(
			commands.ShowStatusCmd("Request created", false),
			commands.LoadCollectionsCmd(),
		), true

	case uimsg.DeleteRequestMsg:
		return m, commands.DeleteRequestCmd(msg.CollectionName, msg.RequestName), true

	case uimsg.RequestDeletedMsg:
		// Close modal if open (implicitly handled by state change)
		if m.state == uimsg.ViewInput {
			m.state = uimsg.ViewCollectionList
		}
		
		// Reactive update: Remove request from current collection
		if m.currentCollection != nil && m.currentCollection.Name == msg.CollectionName {
			newReqs := []storage.Request{}
			for _, req := range m.currentCollection.Requests {
				if req.Name != msg.RequestName {
					newReqs = append(newReqs, req)
				}
			}
			m.currentCollection.Requests = newReqs
			m.collections.SetCollection(*m.currentCollection)
		}
		
		// Remove from tabs if open
		newTabs := []components.RequestTab{}
		activeTabClosed := false
		for i, tab := range m.tabs {
			// Check if this tab matches the deleted request
			// We might need collection name in tab to be 100% sure, but request name + baseurl is good enough for now
			// If we had collection name in tab, it would be safer. 
			// For now, let's just match by name.
			if tab.Request.Name == msg.RequestName {
				if i == m.activeTab {
					activeTabClosed = true
				}
				continue
			}
			newTabs = append(newTabs, tab)
		}
		m.tabs = newTabs
		
		// Adjust active tab
		if activeTabClosed {
			if len(m.tabs) > 0 {
				m.activeTab = max(0, min(m.activeTab, len(m.tabs)-1))
				m.loadActiveTab()
			} else {
				m.activeTab = -1
				// Clear request view
				m.request.Clear()
				m.response.Clear()
				// Show empty state or something? 
				// For now just clear it.
			}
		} else if m.activeTab >= len(m.tabs) {
			m.activeTab = len(m.tabs) - 1
		}
		
		// If the currently viewed request (not via tab) was deleted
		if m.activeTab == -1 && m.request.GetRequestName() == msg.RequestName {
			m.request.Clear()
			m.response.Clear()
		}

		return m, nil, true

	case uimsg.DeleteCollectionMsg:
		return m, commands.DeleteCollectionCmd(msg.Name), true

	case uimsg.CollectionDeletedMsg:
		if m.currentCollection != nil && m.currentCollection.Name == msg.Name {
			m.currentCollection = nil
			m.state = uimsg.ViewWelcome
		}
		return m, tea.Batch(
			commands.ShowStatusCmd("Collection deleted", false),
			commands.LoadCollectionsCmd(),
		), true

	case uimsg.CollectionRenamedMsg:
		if m.currentCollection != nil && m.currentCollection.Name == msg.OldName {
			m.currentCollection.Name = msg.NewName
		}
		return m, tea.Batch(
			commands.ShowStatusCmd("Collection renamed", false),
			commands.LoadCollectionsCmd(),
		), true

	case uimsg.CreateCollectionMsg:
		if strings.TrimSpace(msg.Name) == "" {
			return m, commands.ShowStatusCmd("Collection name cannot be empty", true), true
		}
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
		m.state = uimsg.ViewCollectionList
		return m, commands.RenameCollectionCmd(msg.OldName, msg.NewName), true

	case uimsg.EnvChangedMsg:
		m.currentEnv = &msg.NewEnv
		m.request.SetVariables(msg.NewEnv.Variables)
		return m, commands.ShowStatusCmd(fmt.Sprintf("Env: %s", msg.NewEnv.Name), false), true

	case uimsg.StatusMsg:
		m.statusText = msg.Message
		m.statusIsErr = msg.IsError

		if msg.IsError {
			m.nextNotificationID++
			notif := Notification{
				ID:      m.nextNotificationID,
				Message: msg.Message,
				IsError: true,
			}
			m.notifications = append(m.notifications, notif)

			// Clear notification after 5 seconds
			return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return uimsg.ClearNotificationMsg{ID: notif.ID}
			}), true
		}

		// Auto-dismiss status after 3 seconds
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return uimsg.ClearStatusMsg{}
		}), true

	case uimsg.ClearStatusMsg:
		m.statusText = ""
		m.statusIsErr = false
		return m, nil, true

	case uimsg.ClearNotificationMsg:
		newNotifications := []Notification{}
		for _, n := range m.notifications {
			if n.ID != msg.ID {
				newNotifications = append(newNotifications, n)
			}
		}
		m.notifications = newNotifications
		return m, nil, true

	case uimsg.ImportCollectionMsg:
		// Close input
		m.state = uimsg.ViewCollectionList

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
		// Close input
		m.state = uimsg.ViewCollectionList

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

	case uimsg.RequestDuplicatedMsg:
		return m, tea.Batch(
			commands.ShowStatusCmd("Request duplicated", false),
			commands.LoadCollectionsCmd(),
		), true

	case uimsg.EnvSavedMsg:
		return m, tea.Batch(
			commands.ShowStatusCmd("Environment saved", false),
			func() tea.Msg { return uimsg.BackMsg{} },
			commands.LoadEnvsCmd(),
		), true

	case uimsg.EnvDeletedMsg:
		return m, tea.Batch(
			commands.ShowStatusCmd("Environment deleted", false),
			commands.LoadEnvsCmd(),
		), true

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
