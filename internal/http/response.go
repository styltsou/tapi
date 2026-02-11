package http

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// MaxResponseBodySize limits response body to 10MB to prevent memory exhaustion
	MaxResponseBodySize = 10 * 1024 * 1024 // 10MB
)

// ProcessedResponse contains the HTTP response and metadata
type ProcessedResponse struct {
	StatusCode int
	Status     string
	Headers    http.Header
	Body       []byte
	Duration   time.Duration
	Size       int64
	Truncated  bool // Indicates if body was truncated due to size limit
}

// ProcessResponse transforms raw http.Response into a TUI-friendly format
func ProcessResponse(resp *http.Response, start time.Time) (*ProcessedResponse, error) {
	// Ensure body is closed
	defer resp.Body.Close()

	// Read body with size limit to prevent memory exhaustion
	limitedReader := io.LimitReader(resp.Body, MaxResponseBodySize)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if body was truncated
	truncated := false
	if int64(len(bodyBytes)) >= MaxResponseBodySize {
		// Check if there's more data
		oneByte := make([]byte, 1)
		n, _ := resp.Body.Read(oneByte)
		if n > 0 {
			truncated = true
		}
	}

	// Calculate duration
	duration := time.Since(start)

	// Copy headers (make a copy to avoid referencing the original)
	headers := make(http.Header)
	for key, values := range resp.Header {
		headers[key] = append([]string{}, values...)
	}

	response := &ProcessedResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Headers:    headers,
		Body:       bodyBytes,
		Duration:   duration,
		Size:       int64(len(bodyBytes)),
		Truncated:  truncated,
	}

	return response, nil
}

// BodyString returns the body as a string
func (r *ProcessedResponse) BodyString() string {
	return string(r.Body)
}

// IsSuccess returns true if status code is 2xx
func (r *ProcessedResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError returns true if status code is 4xx or 5xx
func (r *ProcessedResponse) IsError() bool {
	return r.StatusCode >= 400
}

// GetHeader returns the first value for a given header key
func (r *ProcessedResponse) GetHeader(key string) string {
	return r.Headers.Get(key)
}

// FormatSize returns a human-readable size (e.g., "1.5 KB", "2.3 MB")
func (r *ProcessedResponse) FormatSize() string {
	size := float64(r.Size)
	units := []string{"B", "KB", "MB", "GB"}

	unitIndex := 0
	for size >= 1024 && unitIndex < len(units)-1 {
		size /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%d %s", int(size), units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}
