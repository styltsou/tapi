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
			w, h := m.response.width, m.response.height
			m.response = NewResponseModel()
			m.response.SetSize(w, h)
		}
	}
}

func (m *Model) closeTab(index int) {
	if index < 0 || index >= len(m.tabs) {
		return
	}

	m.tabs = append(m.tabs[:index], m.tabs[index+1:]...)

	if m.activeTab >= len(m.tabs) {
		m.activeTab = len(m.tabs) - 1
	}
	if m.activeTab < 0 {
		m.activeTab = 0
	}

	if len(m.tabs) == 0 {
		m.activeTab = -1
		rw, rh := m.request.width, m.request.height
		respW, respH := m.response.width, m.response.height
		m.request = NewRequestModel()
		m.response = NewResponseModel()
		m.request.SetSize(rw, rh)
		m.response.SetSize(respW, respH)
	} else {
		m.loadActiveTab()
	}
}
