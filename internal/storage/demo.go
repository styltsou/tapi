package storage

import "github.com/styltsou/tapi/internal/logger"

// CreateDemoCollection creates a default collection for first-time users
func CreateDemoCollection() error {
	demo := Collection{
		Name:    "Demo Collection",
		BaseURL: "https://httpbin.org",
		Requests: []Request{
			{
				Name:   "Get IP",
				Method: "GET",
				URL:    "/ip",
			},
			{
				Name:   "Post JSON",
				Method: "POST",
				URL:    "/post",
				Body:   `{"message": "Hello TAPI!"}`,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
			},
		},
	}

	if err := SaveCollection(demo); err != nil {
		logger.Logger.Error("Failed to create demo collection", "error", err)
		return err
	}
	
	logger.Logger.Info("Created demo collection")
	return nil
}
