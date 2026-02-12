package ui

import (
	"github.com/styltsou/tapi/internal/ui/components"
)

// --- Tab Helper Methods ---

func (m *Model) saveCurrentTab() {
	if m.activeTab >= 0 && m.activeTab < len(m.tabs) {
		req, _ := m.request.BuildRequest()
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
			w, h := m.response.Width, m.response.Height
			m.response = components.NewResponseModel()
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
		rw, rh := m.request.Width, m.request.Height
		respW, respH := m.response.Width, m.response.Height
		m.request = components.NewRequestModel()
		m.response = components.NewResponseModel()
		m.request.SetSize(rw, rh)
		m.response.SetSize(respW, respH)
	} else {
		m.loadActiveTab()
	}
}
