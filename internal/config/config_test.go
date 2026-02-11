package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", cfg.Timeout)
	}
	if cfg.Theme.Primary != "#7D56F4" {
		t.Errorf("Expected default primary color #7D56F4, got %s", cfg.Theme.Primary)
	}
}

func TestLoadFrom_Valid(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	content := []byte(`
timeout: 60
default_headers:
  User-Agent: TAPI-Test
  Accept: application/json
theme:
  primary: "#FF0000"
`)
	if err := os.WriteFile(configPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	cfg := LoadFrom(configPath)

	if cfg.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", cfg.Timeout)
	}
	if len(cfg.DefaultHeaders) != 2 {
		t.Errorf("Expected 2 default headers, got %d", len(cfg.DefaultHeaders))
	}
	if cfg.DefaultHeaders["User-Agent"] != "TAPI-Test" {
		t.Errorf("Expected User-Agent TAPI-Test, got %s", cfg.DefaultHeaders["User-Agent"])
	}
	if cfg.Theme.Primary != "#FF0000" {
		t.Errorf("Expected primary color #FF0000, got %s", cfg.Theme.Primary)
	}
	// Check that defaults are preserved for missing fields (secondary color)
	if cfg.Theme.Secondary != "#04B575" {
		t.Errorf("Expected default secondary color #04B575, got %s", cfg.Theme.Secondary)
	}
}

func TestLoadFrom_MissingFile(t *testing.T) {
	cfg := LoadFrom("/path/to/non/existent/file.yaml")
	// Should return defaults
	if cfg.Timeout != 30 {
		t.Errorf("Expected default timeout 30, got %d", cfg.Timeout)
	}
}

func TestTimeoutDuration(t *testing.T) {
	cfg := Config{Timeout: 10}
	if cfg.TimeoutDuration() != 10*time.Second {
		t.Errorf("Expected 10s duration, got %v", cfg.TimeoutDuration().String())
	}
}
