package importer

import (
	"fmt"
	"strings"

	"github.com/styltsou/tapi/internal/storage"
)

// importCurl parses a cURL command string and returns a single request.
func importCurl(input string) (storage.Request, error) {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "curl ") && !strings.HasPrefix(input, "curl\t") {
		return storage.Request{}, fmt.Errorf("not a valid cURL command")
	}

	tokens := tokenizeCurl(input)

	req := storage.Request{
		Name:    "Imported Request",
		Method:  "GET",
		Headers: make(map[string]string),
	}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		switch {
		case token == "-X" || token == "--request":
			if i+1 < len(tokens) {
				i++
				req.Method = strings.ToUpper(tokens[i])
			}

		case token == "-H" || token == "--header":
			if i+1 < len(tokens) {
				i++
				parts := strings.SplitN(tokens[i], ":", 2)
				if len(parts) == 2 {
					req.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
				}
			}

		case token == "-d" || token == "--data" || token == "--data-raw" || token == "--data-binary":
			if i+1 < len(tokens) {
				i++
				req.Body = tokens[i]
				// If method is still GET and we have a body, switch to POST
				if req.Method == "GET" {
					req.Method = "POST"
				}
			}

		case !strings.HasPrefix(token, "-") && token != "curl":
			// This is likely the URL
			url := strings.Trim(token, "'\"")
			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
				req.URL = url
			}
		}
	}

	if req.URL == "" {
		return storage.Request{}, fmt.Errorf("no URL found in cURL command")
	}

	// Derive a name from the URL path
	if idx := strings.LastIndex(req.URL, "/"); idx > 0 {
		path := req.URL[idx+1:]
		if qIdx := strings.Index(path, "?"); qIdx > 0 {
			path = path[:qIdx]
		}
		if path != "" {
			req.Name = path
		}
	}

	return req, nil
}

// tokenizeCurl splits a cURL command into tokens, respecting quotes.
// Handles line continuations with backslash.
func tokenizeCurl(input string) []string {
	// Remove line continuations
	input = strings.ReplaceAll(input, "\\\n", " ")
	input = strings.ReplaceAll(input, "\\\r\n", " ")

	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		switch {
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
		case ch == '"' && !inSingle:
			inDouble = !inDouble
		case ch == '\\' && inDouble && i+1 < len(input):
			// Escape character inside double quotes
			i++
			current.WriteByte(input[i])
		case (ch == ' ' || ch == '\t') && !inSingle && !inDouble:
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}
