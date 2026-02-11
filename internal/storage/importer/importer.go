// Package importer handles importing collections from external formats
// (Postman v2.1, Insomnia v4, cURL).
package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/styltsou/tapi/internal/storage"
)

// ImportFromFile reads a file and auto-detects its format (Postman, Insomnia, or cURL).
// Returns one or more collections parsed from the file.
func ImportFromFile(path string) ([]storage.Collection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ImportFromBytes(data)
}

// ImportFromBytes auto-detects the format and imports from raw bytes.
func ImportFromBytes(data []byte) ([]storage.Collection, error) {
	content := strings.TrimSpace(string(data))

	// Check if it's a cURL command
	if strings.HasPrefix(content, "curl ") || strings.HasPrefix(content, "curl\t") {
		req, err := importCurl(content)
		if err != nil {
			return nil, fmt.Errorf("cURL import failed: %w", err)
		}
		return []storage.Collection{
			{
				Name:     "Imported from cURL",
				Requests: []storage.Request{req},
			},
		}, nil
	}

	// Try JSON-based formats
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unsupported format: file is not valid JSON or cURL")
	}

	// Detect Postman: has "info" and "item" top-level keys
	if _, hasInfo := raw["info"]; hasInfo {
		if _, hasItem := raw["item"]; hasItem {
			col, err := importPostman(data)
			if err != nil {
				return nil, err
			}
			return []storage.Collection{col}, nil
		}
	}

	// Detect Insomnia: has "_type" and "resources" top-level keys
	if _, hasType := raw["_type"]; hasType {
		if _, hasResources := raw["resources"]; hasResources {
			col, err := importInsomnia(data)
			if err != nil {
				return nil, err
			}
			return []storage.Collection{col}, nil
		}
	}

	return nil, fmt.Errorf("unsupported format: could not detect Postman or Insomnia structure")
}
