package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
)

func TestMain(m *testing.M) {
	_ = logger.Init()
	os.Exit(m.Run())
}

func TestClient_Execute_URLResolution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	client := NewClient(10 * time.Second)
	client.HTTPClient.Timeout = 1 * time.Second

	tests := []struct {
		name    string
		reqURL  string
		baseURL string
		wantURL string
	}{
		{
			name:    "Absolute URL",
			reqURL:  server.URL + "/direct",
			baseURL: "http://other.com",
			wantURL: "/direct",
		},
		{
			name:    "Relative URL",
			reqURL:  "/relative",
			baseURL: server.URL,
			wantURL: "/relative",
		},
		{
			name:    "Relative URL with trailing/leading slashes",
			reqURL:  "api/test",
			baseURL: server.URL + "/",
			wantURL: "/api/test",
		},
		{
			name:    "Relative URL with leading slash on base with trailing",
			reqURL:  "/api/test",
			baseURL: server.URL + "/",
			wantURL: "/api/test",
		},
		{
			name:    "Base URL with path and relative path",
			reqURL:  "users",
			baseURL: server.URL + "/api/v1/",
			wantURL: "/api/v1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := storage.Request{
				Method: "GET",
				URL:    tt.reqURL,
			}
			_, err := client.Execute(req, tt.baseURL)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}
		})
	}
}

func TestClient_Execute_RequestData(t *testing.T) {
	var capturedHeader string
	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Test")
		bodyNodes, _ := io.ReadAll(r.Body)
		capturedBody = string(bodyNodes)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(10 * time.Second)
	req := storage.Request{
		Method: "POST",
		URL:    server.URL,
		Headers: map[string]string{
			"X-Test": "Value",
		},
		Body: "test-body",
	}

	_, err := client.Execute(req, "")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if capturedHeader != "Value" {
		t.Errorf("Expected header Value, got %s", capturedHeader)
	}

	if capturedBody != "test-body" {
		t.Errorf("Expected body test-body, got %s", capturedBody)
	}
}

func TestClient_Execute_BasicAuth(t *testing.T) {
	var capturedAuth string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(10 * time.Second)

	// Test with Basic Auth
	req := storage.Request{
		Method: "GET",
		URL:    server.URL,
		Auth: &storage.BasicAuth{
			Username: "admin",
			Password: "secret123",
		},
	}

	_, err := client.Execute(req, "")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if capturedAuth == "" {
		t.Fatal("Expected Authorization header to be set")
	}
	if capturedAuth != "Basic YWRtaW46c2VjcmV0MTIz" {
		t.Errorf("Unexpected Authorization header: %s", capturedAuth)
	}

	// Test without Basic Auth
	capturedAuth = ""
	reqNoAuth := storage.Request{
		Method: "GET",
		URL:    server.URL,
	}
	_, err = client.Execute(reqNoAuth, "")
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	if capturedAuth != "" {
		t.Errorf("Expected no Authorization header, got: %s", capturedAuth)
	}
}
