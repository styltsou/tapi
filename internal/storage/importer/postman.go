package importer

import (
	"encoding/json"
	"fmt"

	"github.com/styltsou/tapi/internal/storage"
)

// Postman v2.1 JSON structures (only the fields we care about)

type postmanCollection struct {
	Info postmanInfo   `json:"info"`
	Item []postmanItem `json:"item"`
}

type postmanInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type postmanItem struct {
	Name    string          `json:"name"`
	Request *postmanRequest `json:"request"`
	// Folders contain nested items
	Item []postmanItem `json:"item"`
}

type postmanRequest struct {
	Method string          `json:"method"`
	URL    json.RawMessage `json:"url"`
	Header []postmanKV     `json:"header"`
	Body   *postmanBody    `json:"body"`
}

type postmanKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type postmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw"`
}

// postmanURL can be either a string or an object with a "raw" field
type postmanURL struct {
	Raw string `json:"raw"`
}

func parsePostmanURL(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try string first
	var urlStr string
	if err := json.Unmarshal(raw, &urlStr); err == nil {
		return urlStr
	}

	// Try object with "raw" field
	var urlObj postmanURL
	if err := json.Unmarshal(raw, &urlObj); err == nil {
		return urlObj.Raw
	}

	return ""
}

// importPostman parses a Postman v2.1 JSON export and returns a collection.
func importPostman(data []byte) (storage.Collection, error) {
	var pc postmanCollection
	if err := json.Unmarshal(data, &pc); err != nil {
		return storage.Collection{}, fmt.Errorf("invalid Postman JSON: %w", err)
	}

	if pc.Info.Name == "" {
		return storage.Collection{}, fmt.Errorf("Postman collection has no name")
	}

	requests := flattenPostmanItems(pc.Item)

	return storage.Collection{
		Name:     pc.Info.Name,
		Requests: requests,
	}, nil
}

// flattenPostmanItems recursively extracts requests from potentially nested folders.
func flattenPostmanItems(items []postmanItem) []storage.Request {
	var requests []storage.Request

	for _, item := range items {
		// If it's a folder (has sub-items but no request), recurse
		if item.Request == nil && len(item.Item) > 0 {
			requests = append(requests, flattenPostmanItems(item.Item)...)
			continue
		}

		if item.Request == nil {
			continue
		}

		req := storage.Request{
			Name:   item.Name,
			Method: item.Request.Method,
			URL:    parsePostmanURL(item.Request.URL),
		}

		// Headers
		if len(item.Request.Header) > 0 {
			req.Headers = make(map[string]string, len(item.Request.Header))
			for _, h := range item.Request.Header {
				req.Headers[h.Key] = h.Value
			}
		}

		// Body
		if item.Request.Body != nil && item.Request.Body.Raw != "" {
			req.Body = item.Request.Body.Raw
		}

		requests = append(requests, req)
	}

	return requests
}
