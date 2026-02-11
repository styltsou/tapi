package storage

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadEnvironments(t *testing.T) {
	// Create a temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "tapi-env-test-*")
	if err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Mock storage path
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	// Create environment file
	envDir := filepath.Join(tmpDir, ".tapi", "environments")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	devEnv := Environment{
		Name: "dev",
		Variables: map[string]string{
			"host": "localhost",
		},
	}
	data, _ := yaml.Marshal(devEnv)
	if err := os.WriteFile(filepath.Join(envDir, "dev.yaml"), data, 0644); err != nil {
		t.Fatalf("Failed to write env file: %v", err)
	}

	// Test Load
	envs, err := LoadEnvironments()
	if err != nil {
		t.Fatalf("LoadEnvironments failed: %v", err)
	}

	if len(envs) != 1 {
		t.Errorf("Expected 1 environment, got %d", len(envs))
	} else if envs[0].Name != "dev" || envs[0].Variables["host"] != "localhost" {
		t.Errorf("Environment data mismatch: %+v", envs[0])
	}
}

func TestSaveEnvironment(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tapi-env-save-test-*")
	if err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	envDir := filepath.Join(tmpDir, ".tapi", "environments")
	if err := os.MkdirAll(envDir, 0755); err != nil {
		t.Fatalf("Failed to create env dir: %v", err)
	}

	testEnv := Environment{
		Name: "Production (Stable)",
		Variables: map[string]string{
			"site": "example.com",
		},
	}

	if err := SaveEnvironment(testEnv); err != nil {
		t.Fatalf("SaveEnvironment failed: %v", err)
	}

	// Verify filename is sanitized (slugified)
	expectedFile := filepath.Join(envDir, "production-stable.yaml")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected sanitized file %s to exist", expectedFile)
	}
}
