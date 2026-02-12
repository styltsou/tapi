package http

import (
	"fmt"
	"net/url"
	"strings"
)

// ResolveURL combines a base URL and a relative request URL.
// It ensures that paths are joined correctly, adding a slash if needed.
func ResolveURL(baseURL, reqURL string) (string, error) {
	if strings.HasPrefix(reqURL, "http://") || strings.HasPrefix(reqURL, "https://") {
		return reqURL, nil
	}

	if baseURL == "" {
		return reqURL, nil
	}

	// Ensure baseURL ends with a slash to treat it as a directory
	// This prevents "http://api.com/v1" + "users" becoming "http://api.com/users"
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	rel, err := url.Parse(reqURL)
	if err != nil {
		return "", fmt.Errorf("invalid relative URL: %w", err)
	}

	return base.ResolveReference(rel).String(), nil
}
