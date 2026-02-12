package http

import (
	"testing"
)

func TestBaseURLResolution(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		reqURL   string
		expected string
	}{
		{
			name:     "BaseURL with slash, req relative",
			baseURL:  "http://api.com/v1/",
			reqURL:   "users",
			expected: "http://api.com/v1/users",
		},
		{
			name:     "BaseURL without slash, req relative",
			baseURL:  "http://api.com/v1", // Implicitly adds slash
			reqURL:   "users",
			expected: "http://api.com/v1/users",
		},
		{
			name:     "BaseURL root, req relative",
			baseURL:  "http://api.com",
			reqURL:   "users",
			expected: "http://api.com/users",
		},
		{
			name:     "Request is absolute URL",
			baseURL:  "http://api.com/v1",
			reqURL:   "http://other.com/foo",
			expected: "http://other.com/foo",
		},
	}

	// client := NewClient()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveURL(tt.baseURL, tt.reqURL)
			if err != nil {
				t.Fatalf("ResolveURL error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ResolveURL(%q, %q) = %q, want %q", tt.baseURL, tt.reqURL, got, tt.expected)
			}
		})
	}
}
