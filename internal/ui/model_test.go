package ui

import (
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
	"github.com/styltsou/tapi/internal/ui/styles"

	"fmt"
	"os"
	"testing"

	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/logger"
)

func TestMain(m *testing.M) {
	_ = logger.Init()
	os.Exit(m.Run())
}

func TestNewModel(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	if m.state != uimsg.ViewWelcome {
		t.Errorf("Expected initial state uimsg.ViewWelcome, got %v", m.state)
	}

	if m.env.Visible {
		t.Error("Expected environment modal to be hidden initially")
	}
}

func TestModel_Update_Navigation(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	// Test uimsg.FocusMsg
	m2, _ := m.Update(uimsg.FocusMsg{Target: uimsg.ViewRequestBuilder})
	m = m2.(Model)
	if m.state != uimsg.ViewRequestBuilder {
		t.Errorf("Expected state uimsg.ViewRequestBuilder after uimsg.FocusMsg, got %v", m.state)
	}

	// Test uimsg.BackMsg
	m3, _ := m.Update(uimsg.BackMsg{})
	m = m3.(Model)
	if m.state != uimsg.ViewWelcome {
		t.Errorf("Expected state uimsg.ViewWelcome after uimsg.BackMsg (no collection), got %v", m.state)
	}
}

func TestModel_Update_EnvToggle(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	if m.env.Visible {
		t.Fatal("Env should be hidden")
	}
}

func TestExpandTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot determine home directory")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"~/foo/bar.json", home + "/foo/bar.json"},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"", ""},
		{"~nope", "~nope"}, // only ~/... should expand, not ~user
	}

	for _, tt := range tests {
		result := expandTilde(tt.input)
		if result != tt.expected {
			t.Errorf("expandTilde(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestErrMsg_SurfacesToStatusBar(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.Width = 200
	m.Height = 40
	
	errMsg := uimsg.ErrMsg{Err: fmt.Errorf("test error")}
	m2, cmd := m.Update(errMsg)
	m = m2.(Model)
	
	if cmd == nil {
		t.Fatal("Expected a command to show status, got nil")
	}
	
	// Execute the command to get the uimsg.StatusMsg
	msg := cmd()
	statusMsg, ok := msg.(uimsg.StatusMsg)
	if !ok {
		t.Fatalf("Expected uimsg.StatusMsg, got %T", msg)
	}
	if !statusMsg.IsError {
		t.Error("Expected status to be an error")
	}
	if statusMsg.Message != "Error: test error" {
		t.Errorf("status = %q, want %q", statusMsg.Message, "Error: test error")
	}
	
	_ = m // avoid unused variable
}

func TestMethodBadge_HEAD(t *testing.T) {
	// Just verify it doesn't panic and produces non-empty output
	badge := styles.MethodBadge("HEAD")
	if badge == "" {
		t.Error("Expected non-empty badge for HEAD")
	}
}
