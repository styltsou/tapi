// Package http contains all the request execution engine
package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
)

// Custom errors
var (
	ErrInvalidURL = errors.New("invalid URL")
	ErrTimeout    = errors.New("request timeout")
	ErrNetwork    = errors.New("network error")
)

// Client wraps http.Client with custom configuration
type Client struct {
	HTTPClient *http.Client
}

// NewClient creates a new HTTP client with safe defaults
func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			// Align with the 30s context timeout
			Timeout: 35 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 10 redirects
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// Execute performs an HTTP request with timeout and error handling
func (c *Client) Execute(req storage.Request, baseURL string) (*ProcessedResponse, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Resolve URL robustly
	fullURL := req.URL
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") && baseURL != "" {
		base, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base URL: %w", err)
		}
		rel, err := url.Parse(req.URL)
		if err != nil {
			return nil, fmt.Errorf("invalid relative URL: %w", err)
		}
		fullURL = base.ResolveReference(rel).String()
	}

	// Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, fullURL, strings.NewReader(req.Body))
	if err != nil {
		logger.Logger.Error("Failed to create request", "url", fullURL, "error", err)
		return nil, ErrInvalidURL
	}

	// Inject headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Execute request and measure time
	start := time.Now()
	resp, err := c.HTTPClient.Do(httpReq)

	// Error handling
	if err != nil {
		duration := time.Since(start)
		// Check if context was cancelled (timeout)
		if ctx.Err() == context.DeadlineExceeded {
			logger.Logger.Info("Request timeout", "url", fullURL, "duration", duration)
			return nil, ErrTimeout
		}

		// Network-related errors (DNS, connection refused, etc.)
		logger.Logger.Info("Network error", "url", fullURL, "error", err)
		return nil, errors.Join(ErrNetwork, err)
	}

	// Process response using internal/http/response.go
	processed, err := ProcessResponse(resp, start)
	if err != nil {
		return nil, err
	}

	logger.Logger.Info("Request completed",
		"method", req.Method,
		"url", fullURL,
		"status", processed.StatusCode,
		"duration", processed.Duration,
	)

	return processed, nil
}
