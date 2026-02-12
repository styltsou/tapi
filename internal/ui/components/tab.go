package components

import (
	"github.com/styltsou/tapi/internal/http"
	"github.com/styltsou/tapi/internal/storage"
)

// RequestTab represents an open request tab
type RequestTab struct {
	Request  storage.Request
	BaseURL  string
	Response *http.ProcessedResponse
	Label    string // e.g. "GET /users"
}
