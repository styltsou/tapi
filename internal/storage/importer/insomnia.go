package importer

import (
	"encoding/json"
	"fmt"

	"github.com/styltsou/tapi/internal/storage"
)

// Insomnia v4 JSON structures

type insomniaExport struct {
	Type         string             `json:"_type"`
	ExportFormat int                `json:"__export_format"`
	Resources    []json.RawMessage  `json:"resources"`
}

type insomniaResource struct {
	Type     string `json:"_type"`
	ID       string `json:"_id"`
	ParentID string `json:"parentId"`
	Name     string `json:"name"`
}

type insomniaRequest struct {
	insomniaResource
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers []insomniaHeader  `json:"headers"`
	Body    *insomniaBody     `json:"body"`
}

type insomniaHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type insomniaBody struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

// importInsomnia parses an Insomnia v4 JSON export and returns a collection.
func importInsomnia(data []byte) (storage.Collection, error) {
	var export insomniaExport
	if err := json.Unmarshal(data, &export); err != nil {
		return storage.Collection{}, fmt.Errorf("invalid Insomnia JSON: %w", err)
	}

	// Find workspace name
	collectionName := "Imported Collection"
	var requests []storage.Request

	for _, raw := range export.Resources {
		// Peek at _type to decide how to unmarshal
		var base insomniaResource
		if err := json.Unmarshal(raw, &base); err != nil {
			continue
		}

		switch base.Type {
		case "workspace":
			if base.Name != "" {
				collectionName = base.Name
			}

		case "request":
			var req insomniaRequest
			if err := json.Unmarshal(raw, &req); err != nil {
				continue
			}

			r := storage.Request{
				Name:   req.Name,
				Method: req.Method,
				URL:    req.URL,
			}

			if len(req.Headers) > 0 {
				r.Headers = make(map[string]string, len(req.Headers))
				for _, h := range req.Headers {
					r.Headers[h.Name] = h.Value
				}
			}

			if req.Body != nil && req.Body.Text != "" {
				r.Body = req.Body.Text
			}

			requests = append(requests, r)
		}
	}

	if len(requests) == 0 {
		return storage.Collection{}, fmt.Errorf("no requests found in Insomnia export")
	}

	return storage.Collection{
		Name:     collectionName,
		Requests: requests,
	}, nil
}
