package ui

import (
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/storage"
)

// RequestTab represents an open request tab
type RequestTab struct {
	Request  storage.Request
	BaseURL  string
	Response *http.ProcessedResponse
	Label    string // e.g. "GET /users"
}

// --- Tab Helper Methods ---

func (m *Model) saveCurrentTab() {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		req, _ := m.request.buildRequest()
		m.tabs[m.activeTab].Request = req
		// Response is already saved on ResponseReadyMsg, but we can ensure it here if needed
		// m.tabs[m.activeTab].Response = m.response.GetResponse() // (if GetResponse existed)
	}
}

func (m *Model) loadActiveTab() {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		tab := m.tabs[m.activeTab]
		m.request.LoadRequest(tab.Request, tab.BaseURL)
		
		if tab.Response != nil {
			m.response.SetResponse(tab.Response, tab.Request)
			m.response.SetLoading(false)
		} else {
			// Clear response pane for new/empty tab
			m.response = NewResponseModel()
			m.response.SetSize(m.response.width, m.response.height)
		}
	}
}

func (m *Model) closeTab(index int) {
	if index < 0 || index >= len(m.tabs) {
		return
	}

	// Remove tab at index
	m.tabs = append(m.tabs[:index], m.tabs[index+1:]...)

	// Adjust active tab
	if m.activeTab >= len(m.tabs) {
		m.activeTab = len(m.tabs) - 1
	}
	if m.activeTab < 0 {
		m.activeTab = 0
	}

	// If no tabs left, clear request/response or load welcome?
	if len(m.tabs) == 0 {
		m.activeTab = -1
		// For now, let's keep the dashboard but maybe clear it
		m.request = NewRequestModel()
		m.response = NewResponseModel()
		m.request.SetSize(m.request.width, m.request.height)
		m.response.SetSize(m.response.width, m.response.height)
	} else {
		m.loadActiveTab()
	}
}
