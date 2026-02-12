package ui

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"

	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/storage"
)

// Update handles all messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// 1. Handle Key Messages
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		var cmd tea.Cmd
		var handled bool
		m, cmd, handled = m.handleKeyMsg(keyMsg)
		if cmd != nil || handled {
			return m, cmd
		}
	}

	// 2. Handle Application Messages
	var cmd tea.Cmd
	var handled bool
	m, cmd, handled = m.handleAppMsg(msg)
	if handled {
		return m, cmd
	}

	// Route updates based on focus and state
	if m.menu.Visible {
		newMenu, menuCmd := m.menu.Update(msg)
		m.menu = newMenu
		cmds = append(cmds, menuCmd)
	} else if m.env.Visible {
		newEnv, envCmd := m.env.Update(msg)
		m.env = newEnv
		cmds = append(cmds, envCmd)
	} else if m.state == uimsg.ViewEnvEditor {
		newEditor, editCmd := m.envEditor.Update(msg)
		m.envEditor = newEditor
		cmds = append(cmds, editCmd)
	} else if m.collectionSelector.Visible {
		newSelector, selCmd := m.collectionSelector.Update(msg)
		m.collectionSelector = newSelector
		cmds = append(cmds, selCmd)
	} else if m.state == uimsg.ViewInput {
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
