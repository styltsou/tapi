package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds application-wide configuration
type Config struct {
	Timeout        int               `yaml:"timeout"`         // request timeout in seconds
	DefaultHeaders map[string]string `yaml:"default_headers"` // headers added to every request
	Theme          ThemeConfig       `yaml:"theme"`
}

// ThemeConfig holds color customization
type ThemeConfig struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Accent    string `yaml:"accent"`
	Error     string `yaml:"error"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Timeout:        30,
		DefaultHeaders: nil,
		Theme: ThemeConfig{
			Primary:   "#7D56F4",
			Secondary: "#04B575",
			Accent:    "#EE6FF8",
			Error:     "#FF4C4C",
		},
	}
}

// Load reads config from ~/.tapi/config.yaml, returning defaults if the file doesn't exist.
func Load() Config {
	cfg := DefaultConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	path := filepath.Join(homeDir, ".tapi", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg // File doesn't exist or unreadable, use defaults
	}

	// Parse YAML over defaults (partial overrides work)
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig() // Corrupt file, use defaults
	}

	return cfg
}

// LoadFrom reads config from a specific path (useful for testing)
func LoadFrom(path string) Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}

// TimeoutDuration returns the timeout as a time.Duration
func (c Config) TimeoutDuration() time.Duration {
	if c.Timeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.Timeout) * time.Second
}
