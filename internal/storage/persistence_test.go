package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/styltsou/tapi/internal/logger"
)

func TestMain(m *testing.M) {
	_ = logger.Init()
	os.Exit(m.Run())
}

func TestFormatName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-api", "My Api"},
		{"user_service", "User Service"},
		{"API_v1-test", "Api V1 Test"},
		{"multiple---dashes", "Multiple Dashes"},
	}

	for _, tt := range tests {
		if actual := formatName(tt.input); actual != tt.expected {
			t.Errorf("formatName(%q) = %q, want %q", tt.input, actual, tt.expected)
		}
	}
}

func TestPersistence(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "tapi-test-*")
	if err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock storage path
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Ensure collections dir exists
	colDir := filepath.Join(tmpDir, ".tapi", "collections")
	if err := os.MkdirAll(colDir, 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}

	col := Collection{
		Name:    "Test Collection",
		BaseURL: "http://test.com",
		Requests: []Request{
			{Name: "Req 1", Method: "GET", URL: "/test"},
		},
	}

	// Test Save
	if err := SaveCollection(col); err != nil {
		t.Fatalf("SaveCollection failed: %v", err)
	}

	// Test Load
	cols, _, err := LoadCollections()
	if err != nil {
		t.Fatalf("LoadCollections failed: %v", err)
	}

	found := false
	for _, c := range cols {
		if c.Name == col.Name {
			found = true
			if len(c.Requests) != 1 || c.Requests[0].Name != "Req 1" {
				t.Errorf("Loaded collection data mismatch: %+v", c)
			}
			break
		}
	}

	if !found {
		t.Error("Saved collection not found in LoadCollections")
	}
}

func TestPersistence_BasicAuth(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tapi-auth-test-*")
	if err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	colDir := filepath.Join(tmpDir, ".tapi", "collections")
	if err := os.MkdirAll(colDir, 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}

	// Save a collection with auth
	col := Collection{
		Name:    "Auth Test",
		BaseURL: "http://auth.example.com",
		Requests: []Request{
			{
				Name:   "With Auth",
				Method: "GET",
				URL:    "/protected",
				Auth: &BasicAuth{
					Username: "admin",
					Password: "s3cret!",
				},
			},
			{
				Name:   "No Auth",
				Method: "GET",
				URL:    "/public",
			},
		},
	}

	if err := SaveCollection(col); err != nil {
		t.Fatalf("SaveCollection failed: %v", err)
	}

	// Load it back
	cols, _, err := LoadCollections()
	if err != nil {
		t.Fatalf("LoadCollections failed: %v", err)
	}

	var loaded *Collection
	for i, c := range cols {
		if c.Name == "Auth Test" {
			loaded = &cols[i]
			break
		}
	}
	if loaded == nil {
		t.Fatal("Saved collection not found")
	}

	if len(loaded.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(loaded.Requests))
	}

	// Verify request with auth
	withAuth := loaded.Requests[0]
	if withAuth.Auth == nil {
		t.Fatal("Expected Auth to be set on first request")
	}
	if withAuth.Auth.Username != "admin" {
		t.Errorf("Username = %q, want %q", withAuth.Auth.Username, "admin")
	}
	if withAuth.Auth.Password != "s3cret!" {
		t.Errorf("Password = %q, want %q", withAuth.Auth.Password, "s3cret!")
	}

	// Verify request without auth
	noAuth := loaded.Requests[1]
	if noAuth.Auth != nil {
		t.Errorf("Expected Auth to be nil on second request, got %+v", noAuth.Auth)
	}
}
