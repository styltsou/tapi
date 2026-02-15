package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/ui/msg"
)

func TestExecuteCommand(t *testing.T) {
	// Setup
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	tests := []struct {
		name     string
		cmdInput string
		wantQuit bool
		wantMsg  string
	}{
		{
			name:     "Quit Command",
			cmdInput: ":q",
			wantQuit: true,
		},
		{
			name:     "Quit Short",
			cmdInput: ":quit",
			wantQuit: true,
		},
		{
			name:     "Help Command",
			cmdInput: ":h",
			wantQuit: false,
		},
		{
			name:     "Unknown Command",
			cmdInput: ":unknown",
			wantQuit: false,
			wantMsg:  "Unknown command: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert input to command string (remove leading :)
			cmdStr := tt.cmdInput
			if len(cmdStr) > 0 && cmdStr[0] == ':' {
				cmdStr = cmdStr[1:]
			}

			// We are testing executeCommand directly since we are in package ui
			newM, cmd, handled := m.executeCommand(cmdStr)

			if !handled {
				t.Error("executeCommand should return handled=true")
			}

			if tt.wantQuit {
				if cmd == nil {
					t.Error("Expected quit command, got nil")
				} else {
					// Verify it's a quit command (this is tricky with tea.Cmd, 
					// but usually tea.Quit returns a specific message or is a specialized function.
					// tea.Quit() returns a tea.QuitMsg. 
					// We can check if the cmd produces tea.QuitMsg.
					message := cmd()
					if _, ok := message.(tea.QuitMsg); !ok {
						t.Errorf("Expected tea.QuitMsg, got %T", message)
					}
				}
			}

			// For unknown command, we expect a StatusMsg
			if tt.wantMsg != "" {
				if cmd == nil {
					t.Error("Expected status command, got nil")
				} else {
					message := cmd()
					statusMsg, ok := message.(msg.StatusMsg)
					if !ok {
						t.Errorf("Expected StatusMsg, got %T", message)
					} else if statusMsg.Message != tt.wantMsg {
						t.Errorf("Expected message %q, got %q", tt.wantMsg, statusMsg.Message)
					}
				}
			}
			
			// For Help, verify help overlay visibility is toggled
			if tt.cmdInput == ":h" {
				// newM is a Model (value), so check its field
				if !newM.helpOverlay.Visible {
					t.Error("Help overlay should be visible after :h")
				}
			}
		})
	}
}

func TestKeyHandling(t *testing.T) {
	cfg := config.DefaultConfig()
	m := NewModel(cfg)

	// Ensure we start in Normal mode, Request pane, and NOT in Welcome screen
	m.state = msg.ViewCollectionList
	m.mode = ModeNormal
	m.focusedPane = PaneRequest

	// Test 1: 'i' to enter insert mode
	msg1 := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	newM, _ := m.Update(msg1)
	m1 := newM.(Model)
	if m1.mode != ModeInsert {
		t.Errorf("Expected ModeInsert after 'i', got %v", m1.mode)
	}

	// Test 2: 'Esc' to exit insert mode
	msg2 := tea.KeyMsg{Type: tea.KeyEsc}
	newM, _ = m1.Update(msg2)
	m2 := newM.(Model)
	if m2.mode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m2.mode)
	}

	// Test 3: Tab to cycle panes
	// Start at PaneRequest
	m2.focusedPane = PaneRequest
	msg3 := tea.KeyMsg{Type: tea.KeyTab}
	
	// Tab -> PaneResponse
	newM, _ = m2.Update(msg3)
	m3 := newM.(Model)
	if m3.focusedPane != PaneResponse {
		t.Errorf("Expected PaneResponse after Tab, got %v", m3.focusedPane)
	}

	// Tab -> PaneCollections
	newM, _ = m3.Update(msg3)
	m4 := newM.(Model)
	if m4.focusedPane != PaneCollections {
		t.Errorf("Expected PaneCollections after 2nd Tab, got %v", m4.focusedPane)
	}

	// Tab -> Back to PaneRequest
	newM, _ = m4.Update(msg3)
	m5 := newM.(Model)
	if m5.focusedPane != PaneRequest {
		t.Errorf("Expected PaneRequest after 3rd Tab, got %v", m5.focusedPane)
	}

	// Test 4: Leader Key (Space)
	msgSpace := tea.KeyMsg{Type: tea.KeySpace}
	newM, _ = m5.Update(msgSpace)
	m6 := newM.(Model)
	if !m6.leaderActive {
		t.Error("Expected leaderActive to be true after Space")
	}
}
