package exporter

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/styltsou/tapi/internal/storage"
)

// ExportCurl generates a cURL command string from a storage.Request.
// If the request URL is relative, it is resolved against baseURL.
func ExportCurl(req storage.Request, baseURL string) string {
	var parts []string
	parts = append(parts, "curl")

	// Method (skip -X for GET since it's the default)
	if req.Method != "" && req.Method != "GET" {
		parts = append(parts, "-X", req.Method)
	}

	// Resolve URL
	resolvedURL := resolveURL(req.URL, baseURL)
	parts = append(parts, shellQuote(resolvedURL))

	// Headers (sorted for deterministic output)
	if len(req.Headers) > 0 {
		keys := make([]string, 0, len(req.Headers))
		for k := range req.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			parts = append(parts, "-H", shellQuote(fmt.Sprintf("%s: %s", k, req.Headers[k])))
		}
	}

	// Body
	if req.Body != "" {
		parts = append(parts, "-d", shellQuote(req.Body))
	}

	return strings.Join(parts, " ")
}

// resolveURL resolves a potentially relative URL against a base URL.
func resolveURL(rawURL, baseURL string) string {
	if baseURL == "" {
		return rawURL
	}
	if strings.HasPrefix(rawURL, "http://") || strings.HasPrefix(rawURL, "https://") {
		return rawURL
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return rawURL
	}
	rel, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return base.ResolveReference(rel).String()
}

// shellQuote wraps a string in single quotes, escaping any single quotes within.
func shellQuote(s string) string {
	// Replace ' with '\'' (end quote, escaped quote, start quote)
	escaped := strings.ReplaceAll(s, "'", `'\''`)
	return "'" + escaped + "'"
}
