package exporter

import (
	"strings"
	"testing"

	"github.com/styltsou/tapi/internal/storage"
)

func TestExportCurl(t *testing.T) {
	tests := []struct {
		name     string
		req      storage.Request
		baseURL  string
		expected []string // substrings to check for
	}{
		{
			name: "Basic GET",
			req: storage.Request{
				Method: "GET",
				URL:    "https://api.example.com/users",
			},
			expected: []string{"curl", "'https://api.example.com/users'"},
		},
		{
			name: "POST with body and headers",
			req: storage.Request{
				Method: "POST",
				URL:    "/users",
				Body:   `{"name":"Alice"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
					"Authorization": "Bearer token",
				},
			},
			baseURL: "https://api.example.com",
			expected: []string{
				"curl",
				"-X POST",
				"'https://api.example.com/users'",
				"-H 'Content-Type: application/json'",
				"-H 'Authorization: Bearer token'",
				"-d '{\"name\":\"Alice\"}'",
			},
		},
		{
			name: "Single quotes escaping",
			req: storage.Request{
				Method: "POST",
				URL:    "https://example.com/query",
				Body:   `It's a "test"`,
			},
			expected: []string{
				`'It'\''s a "test"'`, // It's -> 'It'\''s'
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExportCurl(tt.req, tt.baseURL)
			for _, exp := range tt.expected {
				if !strings.Contains(got, exp) {
					t.Errorf("expected command to contain %q, got: %s", exp, got)
				}
			}
		})
	}
}
