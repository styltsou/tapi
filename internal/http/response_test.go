package http

import (
	"bytes"
	"io"
	"net/http"
	"testing"
	"time"
)


func TestProcessResponse(t *testing.T) {
	// Create a mock http.Response
	body := "Hello, TAPI!"
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	start := time.Now().Add(-1 * time.Second)
	processed, err := ProcessResponse(resp, start)

	if err != nil {
		t.Fatalf("ProcessResponse failed: %v", err)
	}

	if processed.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", processed.StatusCode)
	}

	if string(processed.Body) != body {
		t.Errorf("Expected body %q, got %q", body, string(processed.Body))
	}

	if processed.Duration < 1*time.Second {
		t.Errorf("Expected duration >= 1s, got %v", processed.Duration)
	}

	if processed.GetHeader("Content-Type") != "text/plain" {
		t.Errorf("Expected Content-Type text/plain, got %s", processed.GetHeader("Content-Type"))
	}
}

func TestProcessedResponse_Helpers(t *testing.T) {
	resp := &ProcessedResponse{StatusCode: 201}
	if !resp.IsSuccess() {
		t.Error("Expected IsSuccess() to be true for 201")
	}

	resp.StatusCode = 404
	if !resp.IsError() {
		t.Error("Expected IsError() to be true for 404")
	}
	if resp.IsSuccess() {
		t.Error("Expected IsSuccess() to be false for 404")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tt := range tests {
		resp := &ProcessedResponse{Size: tt.size}
		if actual := resp.FormatSize(); actual != tt.expected {
			t.Errorf("FormatSize(%d) = %s, want %s", tt.size, actual, tt.expected)
		}
	}
}

func TestMaxBodySize(t *testing.T) {
	// Create a body larger than MaxResponseBodySize
	largeBody := bytes.Repeat([]byte("a"), MaxResponseBodySize+100)
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(largeBody)),
	}

	processed, err := ProcessResponse(resp, time.Now())
	if err != nil {
		t.Fatalf("ProcessResponse failed: %v", err)
	}

	if !processed.Truncated {
		t.Error("Expected response to be truncated")
	}

	if int64(len(processed.Body)) != MaxResponseBodySize {
		t.Errorf("Expected body size %d, got %d", MaxResponseBodySize, len(processed.Body))
	}
}
