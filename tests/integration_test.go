package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/charmbracelet/log"
	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
	"github.com/styltsou/tapi/internal/ui"
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
)

func TestIntegration_RequestExecution(t *testing.T) {
	// 0. Init Logger (prevent panic)
	logger.Logger = log.New(os.Stdout)
	// 1. Setup Mock Server
	mockResponse := `{"status": "ok", "message": "integration test"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("Expected path /test, got %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, mockResponse)
	}))
	defer server.Close()

	// 2. Initialize Model
	cfg := config.DefaultConfig()
	cfg.Timeout = 5 // 5 seconds for tests
	model := ui.NewModel(cfg)

	// 3. Prepare Test Data
	req := storage.Request{
		Name:   "Integration Request",
		Method: "GET",
		URL:    "/test",
	}

	// 4. Simulate sending ExecuteRequestMsg
	// We are testing how the model updates when it receives an ExecuteRequestMsg
	// Note: We can't easily simulate the *entire* tea program loop without a tea.TestProgram (which is experimental/complex).
	// Instead, we will directly call Update with the messages that would be generated.

	// Step 4a: Send ExecuteRequestMsg
	execMsg := uimsg.ExecuteRequestMsg{
		Request: req,
		BaseURL: server.URL,
	}
	
	// Update model with ExecuteRequestMsg
	updatedModel, cmd := model.Update(execMsg)
	m := updatedModel.(ui.Model)

	// Verify state changed to waiting (response pane focused/loading)
	// We can't easily check private fields, but we can check the command returned.
	// The command from ExecuteRequestMsg should be an HTTP request command.
	
	if cmd == nil {
		t.Fatal("Expected command from ExecuteRequestMsg, got nil")
	}

	// Step 4b: Execute the command to get the ResponseReadyMsg
	// This simulates the async I/O
	// Step 4b: Execute the command to get the ResponseReadyMsg
	msg := cmd()
	
	var responseReadyMsg uimsg.ResponseReadyMsg
	found := false

	if batchMsg, ok := msg.(tea.BatchMsg); ok {
		for _, cmd := range batchMsg {
			m := cmd()
			if rrm, ok := m.(uimsg.ResponseReadyMsg); ok {
				responseReadyMsg = rrm
				found = true
				break
			}
			if errMsg, ok := m.(uimsg.ErrMsg); ok {
				t.Fatalf("Command returned ErrMsg: %v", errMsg.Err)
			}
		}
	} else if rrm, ok := msg.(uimsg.ResponseReadyMsg); ok {
		responseReadyMsg = rrm
		found = true
	} else if errMsg, ok := msg.(uimsg.ErrMsg); ok {
		t.Fatalf("Command returned ErrMsg: %v", errMsg.Err)
	}

	if !found {
		t.Fatalf("Expected ResponseReadyMsg, got %T: %v", msg, msg)
	}

	// Step 4c: Verify Response content
	resp := responseReadyMsg.Response
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if string(resp.Body) != mockResponse {
		t.Errorf("Expected body %q, got %q", mockResponse, string(resp.Body))
	}

	// Step 5: Update model with ResponseReadyMsg
	finalModel, _ := m.Update(responseReadyMsg)
	fm := finalModel.(ui.Model)

	// Here we would ideally verify that the response is visible in the UI, 
	// but accessing private fields like 'response' model state is hard from outside package.
	// For integration tests, verifying the message flow (Input -> Cmd -> Output Msg) is often sufficient 
	// to prove the core logic works wiring works.
	_ = fm
}
