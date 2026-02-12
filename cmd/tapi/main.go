package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/styltsou/tapi/internal/config"
	"github.com/styltsou/tapi/internal/logger"
	"github.com/styltsou/tapi/internal/storage"
	"github.com/styltsou/tapi/internal/ui"
)

const Version = "v0.1.0"

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}

	logger.Logger.Info("Starting TAPI " + Version)

	// Load configuration
	cfg := config.Load()

	// Ensure collections directory exists
	collectionsPath, err := storage.GetStoragePath("collections")
	if err != nil {
		fmt.Printf("Error locating storage path: %v\n", err)
		os.Exit(1)
	}

	if err := storage.EnsureDir(collectionsPath); err != nil {
		fmt.Printf("Error creating storage directory: %v\n", err)
		os.Exit(1)
	}

	// First-run experience: create demo collection if none exist
	collections, err := storage.LoadCollections()
	if err == nil && len(collections) == 0 {
		logger.Logger.Info("First run detected, creating demo collection")
		demo := storage.Collection{
			Name:    "Demo Collection",
			BaseURL: "https://httpbin.org",
			Requests: []storage.Request{
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
		if err := storage.SaveCollection(demo); err != nil {
			logger.Logger.Error("Failed to create demo collection", "error", err)
		}
	}

	// Create and run program
	p := tea.NewProgram(ui.NewModel(cfg), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
