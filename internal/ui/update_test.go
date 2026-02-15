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
					msg := cmd()
					if _, ok := msg.(tea.QuitMsg); !ok {
						t.Errorf("Expected tea.QuitMsg, got %T", msg)
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
