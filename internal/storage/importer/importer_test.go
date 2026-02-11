package importer

import (
	"testing"
)

// ========================================
// Postman Tests
// ========================================

func TestImportPostman(t *testing.T) {
	t.Run("basic collection", func(t *testing.T) {
		data := []byte(`{
			"info": {
				"name": "My API",
				"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
			},
			"item": [
				{
					"name": "Get Users",
					"request": {
						"method": "GET",
						"url": "https://api.example.com/users",
						"header": [
							{"key": "Accept", "value": "application/json"}
						]
					}
				},
				{
					"name": "Create User",
					"request": {
						"method": "POST",
						"url": {"raw": "https://api.example.com/users"},
						"header": [
							{"key": "Content-Type", "value": "application/json"}
						],
						"body": {
							"mode": "raw",
							"raw": "{\"name\": \"John\"}"
						}
					}
				}
			]
		}`)

		col, err := importPostman(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if col.Name != "My API" {
			t.Errorf("name = %q, want %q", col.Name, "My API")
		}
		if len(col.Requests) != 2 {
			t.Fatalf("got %d requests, want 2", len(col.Requests))
		}

		// First request
		r := col.Requests[0]
		if r.Name != "Get Users" || r.Method != "GET" || r.URL != "https://api.example.com/users" {
			t.Errorf("request 0 mismatch: %+v", r)
		}
		if r.Headers["Accept"] != "application/json" {
			t.Errorf("missing Accept header")
		}

		// Second request
		r = col.Requests[1]
		if r.Method != "POST" || r.Body != `{"name": "John"}` {
			t.Errorf("request 1 mismatch: method=%s body=%s", r.Method, r.Body)
		}
	})

	t.Run("nested folders", func(t *testing.T) {
		data := []byte(`{
			"info": {"name": "Nested"},
			"item": [
				{
					"name": "Auth Folder",
					"item": [
						{
							"name": "Login",
							"request": {"method": "POST", "url": "https://api.example.com/login"}
						}
					]
				},
				{
					"name": "Health",
					"request": {"method": "GET", "url": "https://api.example.com/health"}
				}
			]
		}`)

		col, err := importPostman(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(col.Requests) != 2 {
			t.Fatalf("got %d requests, want 2 (flattened)", len(col.Requests))
		}
		if col.Requests[0].Name != "Login" {
			t.Errorf("expected flattened request 'Login', got %q", col.Requests[0].Name)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := importPostman([]byte(`not json`))
		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := importPostman([]byte(`{"info": {}, "item": []}`))
		if err == nil {
			t.Error("expected error for missing name")
		}
	})
}

// ========================================
// Insomnia Tests
// ========================================

func TestImportInsomnia(t *testing.T) {
	t.Run("basic export", func(t *testing.T) {
		data := []byte(`{
			"_type": "export",
			"__export_format": 4,
			"resources": [
				{
					"_type": "workspace",
					"_id": "wrk_1",
					"name": "My Workspace"
				},
				{
					"_type": "request",
					"_id": "req_1",
					"parentId": "wrk_1",
					"name": "Get Items",
					"method": "GET",
					"url": "https://api.example.com/items",
					"headers": [
						{"name": "Accept", "value": "application/json"}
					]
				},
				{
					"_type": "request",
					"_id": "req_2",
					"parentId": "wrk_1",
					"name": "Create Item",
					"method": "POST",
					"url": "https://api.example.com/items",
					"headers": [],
					"body": {"mimeType": "application/json", "text": "{\"title\": \"test\"}"}
				}
			]
		}`)

		col, err := importInsomnia(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if col.Name != "My Workspace" {
			t.Errorf("name = %q, want %q", col.Name, "My Workspace")
		}
		if len(col.Requests) != 2 {
			t.Fatalf("got %d requests, want 2", len(col.Requests))
		}
		if col.Requests[0].Method != "GET" || col.Requests[0].Headers["Accept"] != "application/json" {
			t.Errorf("request 0 mismatch: %+v", col.Requests[0])
		}
		if col.Requests[1].Body != `{"title": "test"}` {
			t.Errorf("request 1 body = %q", col.Requests[1].Body)
		}
	})

	t.Run("no requests", func(t *testing.T) {
		data := []byte(`{"_type": "export", "__export_format": 4, "resources": [{"_type": "workspace", "_id": "w1", "name": "Empty"}]}`)
		_, err := importInsomnia(data)
		if err == nil {
			t.Error("expected error for empty export")
		}
	})
}

// ========================================
// cURL Tests
// ========================================

func TestImportCurl(t *testing.T) {
	t.Run("simple GET", func(t *testing.T) {
		req, err := importCurl(`curl https://api.example.com/users`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Method != "GET" {
			t.Errorf("method = %q, want GET", req.Method)
		}
		if req.URL != "https://api.example.com/users" {
			t.Errorf("url = %q", req.URL)
		}
	})

	t.Run("POST with body and headers", func(t *testing.T) {
		cmd := `curl -X POST https://api.example.com/users -H 'Content-Type: application/json' -d '{"name": "John"}'`
		req, err := importCurl(cmd)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Method != "POST" {
			t.Errorf("method = %q, want POST", req.Method)
		}
		if req.Headers["Content-Type"] != "application/json" {
			t.Errorf("Content-Type header = %q", req.Headers["Content-Type"])
		}
		if req.Body != `{"name": "John"}` {
			t.Errorf("body = %q", req.Body)
		}
	})

	t.Run("implicit POST with data", func(t *testing.T) {
		req, err := importCurl(`curl https://api.example.com/data -d 'hello=world'`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Method != "POST" {
			t.Errorf("method = %q, want POST (implicit from -d)", req.Method)
		}
	})

	t.Run("quoted URL", func(t *testing.T) {
		req, err := importCurl(`curl "https://api.example.com/search?q=test"`)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.URL != "https://api.example.com/search?q=test" {
			t.Errorf("url = %q", req.URL)
		}
	})

	t.Run("no URL", func(t *testing.T) {
		_, err := importCurl(`curl -H "Accept: text/html"`)
		if err == nil {
			t.Error("expected error for missing URL")
		}
	})

	t.Run("not curl", func(t *testing.T) {
		_, err := importCurl(`wget https://example.com`)
		if err == nil {
			t.Error("expected error for non-curl command")
		}
	})
}

// ========================================
// Auto-Detection Tests
// ========================================

func TestImportFromBytes(t *testing.T) {
	t.Run("detects postman", func(t *testing.T) {
		data := []byte(`{"info": {"name": "Test"}, "item": [{"name": "R1", "request": {"method": "GET", "url": "https://example.com"}}]}`)
		cols, err := ImportFromBytes(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cols) != 1 || cols[0].Name != "Test" {
			t.Errorf("unexpected result: %+v", cols)
		}
	})

	t.Run("detects insomnia", func(t *testing.T) {
		data := []byte(`{"_type": "export", "__export_format": 4, "resources": [{"_type": "workspace", "_id": "w1", "name": "Ins"}, {"_type": "request", "_id": "r1", "parentId": "w1", "name": "R1", "method": "GET", "url": "https://example.com"}]}`)
		cols, err := ImportFromBytes(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cols) != 1 || cols[0].Name != "Ins" {
			t.Errorf("unexpected result: %+v", cols)
		}
	})

	t.Run("detects curl", func(t *testing.T) {
		data := []byte(`curl https://example.com/api`)
		cols, err := ImportFromBytes(data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cols) != 1 || len(cols[0].Requests) != 1 {
			t.Errorf("unexpected result: %+v", cols)
		}
	})

	t.Run("unsupported format", func(t *testing.T) {
		_, err := ImportFromBytes([]byte(`<xml>not supported</xml>`))
		if err == nil {
			t.Error("expected error for unsupported format")
		}
	})
}
