package storage

import (
	"testing"
)

func TestSubstitute(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		env      map[string]string
		expected string
	}{
		{
			name:     "Simple substitution",
			input:    "http://{{host}}",
			env:      map[string]string{"host": "localhost"},
			expected: "http://localhost",
		},
		{
			name:     "Multiple substitution",
			input:    "http://{{host}}:{{port}}",
			env:      map[string]string{"host": "localhost", "port": "8080"},
			expected: "http://localhost:8080",
		},
		{
			name:     "Recursive substitution",
			input:    "{{url}}/api",
			env:      map[string]string{"url": "http://{{host}}", "host": "localhost"},
			expected: "http://localhost/api",
		},
		{
			name:     "Missing variable",
			input:    "http://{{unknown}}",
			env:      map[string]string{"host": "localhost"},
			expected: "http://{{unknown}}",
		},
		{
			name:     "Nested with spaces",
			input:    "{{ url }}",
			env:      map[string]string{"url": "http://localhost"},
			expected: "http://localhost",
		},
		{
			name:     "No variables",
			input:    "http://localhost",
			env:      map[string]string{"host": "localhost"},
			expected: "http://localhost",
		},
		{
			name:     "Nil env",
			input:    "http://{{host}}",
			env:      nil,
			expected: "http://{{host}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := Substitute(tt.input, tt.env)
			if actual != tt.expected {
				t.Errorf("Substitute() = %v, want %v", actual, tt.expected)
			}
		})
	}
}
