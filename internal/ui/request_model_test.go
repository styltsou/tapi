package ui

import (
	"strings"
	"testing"
)

func TestRequestModel_ParseURLParams(t *testing.T) {
	m := NewRequestModel()
	
	// Test case 1: Basic Path Params
	m.pathInput.SetValue("https://api.example.com/users/:id/posts/:postId")
	m.parseURLParams()
	
	if len(m.pathParamsInputs) != 2 {
		t.Errorf("Expected 2 path params, got %d", len(m.pathParamsInputs))
	}
	
	if m.pathParamsInputs[0].key.Value() != "id" {
		t.Errorf("Expected first param key 'id', got '%s'", m.pathParamsInputs[0].key.Value())
	}
	if m.pathParamsInputs[1].key.Value() != "postId" {
		t.Errorf("Expected second param key 'postId', got '%s'", m.pathParamsInputs[1].key.Value())
	}

	// Test case 2: Query Params parsing
	m.pathInput.SetValue("https://api.example.com/search?q=tapi&page=1")
	m.parseURLParams()
	
	if len(m.queryInputs) != 2 {
		t.Errorf("Expected 2 query params, got %d", len(m.queryInputs))
	}
	
	// Query params might be unordered if utilizing map internally, but here likely ordered by URL appearance or not guaranteed?
	// The implementation uses url.ParseQuery which returns a map, so order is NOT guaranteed.
	// But our implementation iterates over the map.
	
	// Let's just check existence.
	foundQ := false
	foundPage := false
	for _, qi := range m.queryInputs {
		if qi.key.Value() == "q" && qi.value.Value() == "tapi" {
			foundQ = true
		}
		if qi.key.Value() == "page" && qi.value.Value() == "1" {
			foundPage = true
		}
	}
	
	if !foundQ || !foundPage {
		t.Errorf("Query params parsing failed. FoundQ: %v, FoundPage: %v", foundQ, foundPage)
	}
}

func TestRequestModel_HighlightVars(t *testing.T) {
	m := NewRequestModel()
	m.SetVariables(map[string]string{
		"base_url": "https://api.test",
		"token":    "secret",
	})
	
	// We need to define colors to match expected output, but verifying exact ANSI codes is brittle.
	// We can check if it contains the replaced value or the original based on existence.
	
	// Test 1: Defined variable
	input := "Bearer {{token}}"
	output := m.highlightVars(input)
	
	if strings.Contains(output, "{{token}}") {
		t.Error("Highlighter should have replaced {{token}} with value (visually)")
	}
	if !strings.Contains(output, "secret") {
		t.Error("Highlighter should show value 'secret' for defined variable")
	}
	
	// Test 2: Undefined variable
	input2 := "{{missing}}"
	output2 := m.highlightVars(input2)
	
	if !strings.Contains(output2, "{{missing}}") {
		t.Error("Highlighter should keep {{missing}} literal for undefined variable")
	}
}

func TestRequestModel_SyncURLFromParams(t *testing.T) {
	m := NewRequestModel()
	m.pathInput.SetValue("https://api.com/endpoint")
	m.parseURLParams() // clear state
	
	// Manually add query params
	m.addRow(&m.queryInputs, "sort", "asc")
	
	// Call sync
	m.syncURLFromParams()
	
	// Check if URL updated
	expected := "https://api.com/endpoint?sort=asc"
	if m.pathInput.Value() != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, m.pathInput.Value())
	}
}
