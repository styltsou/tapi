package ui

import (
	"strings"
	"testing"

	"github.com/styltsou/tapi/internal/storage"
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

func TestRequestModel_LoadRequest_WithAuth(t *testing.T) {
	m := NewRequestModel()

	req := storage.Request{
		Name:   "Auth Request",
		Method: "POST",
		URL:    "/api/login",
		Auth: &storage.BasicAuth{
			Username: "testuser",
			Password: "testpass",
		},
	}

	m.LoadRequest(req, "http://example.com")

	if !m.authEnabled {
		t.Error("Expected authEnabled to be true after loading request with auth")
	}
	if m.authUsername.Value() != "testuser" {
		t.Errorf("authUsername = %q, want %q", m.authUsername.Value(), "testuser")
	}
	if m.authPassword.Value() != "testpass" {
		t.Errorf("authPassword = %q, want %q", m.authPassword.Value(), "testpass")
	}
}

func TestRequestModel_LoadRequest_WithoutAuth(t *testing.T) {
	m := NewRequestModel()

	// First load a request with auth to set state
	m.LoadRequest(storage.Request{
		Name: "First", Method: "GET", URL: "/first",
		Auth: &storage.BasicAuth{Username: "user1", Password: "pass1"},
	}, "")

	if !m.authEnabled {
		t.Fatal("Expected auth to be enabled after first load")
	}

	// Now load a request without auth — should clear everything
	m.LoadRequest(storage.Request{
		Name: "Second", Method: "GET", URL: "/second",
	}, "")

	if m.authEnabled {
		t.Error("Expected authEnabled to be false after loading request without auth")
	}
	if m.authUsername.Value() != "" {
		t.Errorf("authUsername should be empty, got %q", m.authUsername.Value())
	}
	if m.authPassword.Value() != "" {
		t.Errorf("authPassword should be empty, got %q", m.authPassword.Value())
	}
}

func TestRequestModel_BuildRequest_WithAuth(t *testing.T) {
	m := NewRequestModel()

	m.LoadRequest(storage.Request{
		Name: "Auth Req", Method: "GET", URL: "/protected",
		Auth: &storage.BasicAuth{Username: "admin", Password: "secret"},
	}, "")

	req, _ := m.buildRequest()

	if req.Auth == nil {
		t.Fatal("Expected Auth in built request")
	}
	if req.Auth.Username != "admin" {
		t.Errorf("Auth.Username = %q, want %q", req.Auth.Username, "admin")
	}
	if req.Auth.Password != "secret" {
		t.Errorf("Auth.Password = %q, want %q", req.Auth.Password, "secret")
	}
}

func TestRequestModel_BuildRequest_AuthDisabled(t *testing.T) {
	m := NewRequestModel()

	m.LoadRequest(storage.Request{
		Name: "Auth Req", Method: "GET", URL: "/protected",
		Auth: &storage.BasicAuth{Username: "admin", Password: "secret"},
	}, "")

	// Now disable auth
	m.authEnabled = false

	req, _ := m.buildRequest()

	if req.Auth != nil {
		t.Errorf("Expected nil Auth when disabled, got %+v", req.Auth)
	}
}

func TestRequestModel_BuildRequest_NoAuth(t *testing.T) {
	m := NewRequestModel()

	m.LoadRequest(storage.Request{
		Name: "Plain Req", Method: "GET", URL: "/public",
	}, "")

	req, _ := m.buildRequest()

	if req.Auth != nil {
		t.Errorf("Expected nil Auth for request without auth, got %+v", req.Auth)
	}
}

func TestRequestModel_AuthToggle(t *testing.T) {
	m := NewRequestModel()

	// Initially disabled
	if m.authEnabled {
		t.Error("Auth should be disabled by default")
	}

	// Toggle on
	m.authEnabled = true
	m.authUsername.SetValue("user")
	m.authPassword.SetValue("pass")

	req, _ := m.buildRequest()
	if req.Auth == nil {
		t.Fatal("Expected Auth after enabling")
	}
	if req.Auth.Username != "user" || req.Auth.Password != "pass" {
		t.Errorf("Auth credentials mismatch: got %+v", req.Auth)
	}

	// Toggle off — credentials remain in fields but buildRequest excludes them
	m.authEnabled = false
	req2, _ := m.buildRequest()
	if req2.Auth != nil {
		t.Error("Expected nil Auth after disabling")
	}

	// Fields should still have values (not cleared on disable)
	if m.authUsername.Value() != "user" {
		t.Error("Username field should retain value when auth is disabled")
	}
}
