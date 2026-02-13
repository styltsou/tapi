package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/storage"
	"github.com/styltsou/tapi/internal/ui/components"
	uimsg "github.com/styltsou/tapi/internal/ui/msg"
)

func TestModel_Update_CommandMode(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.state = uimsg.ViewCollectionList // Ensure we are not in Welcome screen

	// Simulate pressing ":"
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	m = m2.(Model)

	if m.mode != ModeCommand {
		t.Errorf("Expected ModeCommand, got %v", m.mode)
	}

	// Simulate pressing "esc"
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = m2.(Model)

	if m.mode != ModeNormal {
		t.Errorf("Expected ModeNormal after Esc, got %v", m.mode)
	}
}

func TestModel_Update_LeaderKey(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.state = uimsg.ViewCollectionList

	// Simulate pressing space " "
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = m2.(Model)

	if !m.leaderActive {
		t.Error("Expected leaderActive to be true after pressing space")
	}
}

func TestModel_Update_ReactiveCollectionRefresh(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	
	// Setup initial state with a selected collection
	col := storage.Collection{Name: "Test Col", Requests: []storage.Request{{Name: "R1"}}}
	m.currentCollection = &col
	m.state = uimsg.ViewCollectionList
	
	// Simulate reloading collections with updated data (e.g. new request added)
	updatedCol := storage.Collection{Name: "Test Col", Requests: []storage.Request{{Name: "R1"}, {Name: "R2"}}}
	loadedMsg := uimsg.CollectionsLoadedMsg{
		Collections: []storage.Collection{updatedCol},
	}

	m2, _ := m.Update(loadedMsg)
	m = m2.(Model)

	if m.currentCollection == nil {
		t.Fatal("currentCollection should not be nil")
	}
	
	if len(m.currentCollection.Requests) != 2 {
		t.Errorf("Expected 2 requests in updated collection, got %d", len(m.currentCollection.Requests))
	}
}

func TestModel_Update_CreateRequestClosesModal(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.state = uimsg.ViewInput // Simulate being in input mode
	
	// Simulate CreateRequestMsg
	createMsg := uimsg.CreateRequestMsg{Name: "New Req"}
	
	m2, cmd := m.Update(createMsg)
	m = m2.(Model)

	if m.state == uimsg.ViewInput {
		t.Error("Expected state to change from ViewInput after CreateRequestMsg")
	}
	
	if cmd == nil {
		t.Error("Expected a command to be returned (commands.CreateRequestCmd)")
	}
}

func TestModel_Update_RequestDeletedMsg(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.state = uimsg.ViewInput // Simulate confirmation modal open
	
	// Setup a tab for the request to be deleted
	reqName := "Delete Me"
	m.tabs = []components.RequestTab{
		{Request: storage.Request{Name: reqName}, BaseURL: "http://test.com"},
	}
	m.activeTab = 0
	
	// Send RequestDeletedMsg
	msg := uimsg.RequestDeletedMsg{
		CollectionName: "Col",
		RequestName:    reqName,
	}
	
	m2, _ := m.Update(msg)
	m = m2.(Model)
	
	// Verify state transition (modal closed)
	if m.state != uimsg.ViewCollectionList {
		t.Errorf("Expected state ViewCollectionList, got %v", m.state)
	}
	
	// Verify tab removed
	if len(m.tabs) != 0 {
		t.Errorf("Expected tabs to be empty, got %d", len(m.tabs))
	}
	
	// Verify active tab reset
	if m.activeTab != -1 {
		t.Errorf("Expected activeTab to be -1, got %d", m.activeTab)
	}
}

func TestModel_CommandExecution(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.state = uimsg.ViewCollectionList
	
	// Enter command mode
	m.mode = ModeCommand
	m.commandInput.Prompt = ":"
	m.commandInput.SetValue("q")
	
	// Simulate Enter key
	m2, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(Model)
	
	if m.mode != ModeNormal {
		t.Errorf("Expected ModeNormal after execution, got %v", m.mode)
	}
	
	if cmd == nil {
		t.Error("Expected a command (tea.Quit) to be returned")
	}
	
	// Verify View doesn't panic in Command Mode
	m.mode = ModeCommand
	m.Width = 100
	m.Height = 30
	view := m.View()
	if len(view) == 0 {
		t.Error("View returned empty string")
	}
}
