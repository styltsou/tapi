package ui

import (
	"os"
	"testing"

	"github.com/styltsou/tapi/internal/logger"
)

func TestMain(m *testing.M) {
	_ = logger.Init()
	os.Exit(m.Run())
}

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.state != ViewWelcome {
		t.Errorf("Expected initial state ViewWelcome, got %v", m.state)
	}

	if m.env.visible {
		t.Error("Expected environment modal to be hidden initially")
	}
}

func TestModel_Update_Navigation(t *testing.T) {
	m := NewModel()

	// Test FocusMsg
	m2, _ := m.Update(FocusMsg{Target: ViewRequestBuilder})
	m = m2.(Model)
	if m.state != ViewRequestBuilder {
		t.Errorf("Expected state ViewRequestBuilder after FocusMsg, got %v", m.state)
	}

	// Test BackMsg
	m3, _ := m.Update(BackMsg{})
	m = m3.(Model)
	if m.state != ViewWelcome {
		t.Errorf("Expected state ViewWelcome after BackMsg (no collection), got %v", m.state)
	}
}

func TestModel_Update_EnvToggle(t *testing.T) {
	// Note: We'd need to mock KeyMsg for better testing, but we can test the handler directly if exposed
	// or test via EnvChangedMsg which is simpler.
	
	m := NewModel()
	if m.env.visible {
		t.Fatal("Env should be hidden")
	}
	
	// Since keys are private, we can't easily trigger KeyMsg without mimicking the bubbletea flow
	// but we can verify the state after Msg handlers.
}
