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
	cfg, err := config.Load()
	if err != nil {
		logger.Logger.Error("Failed to load config, using defaults", "error", err)
	}





	// First-run experience: create demo collection if none exist
	collections, _, err := storage.LoadCollections()
	if err == nil && len(collections) == 0 {
		logger.Logger.Info("First run detected, creating demo collection")
		if err := storage.CreateDemoCollection(); err != nil {
			logger.Logger.Error("Failed to create demo collection", "error", err)
		} else {
			// Reload collections to include the demo
			collections, _, _ = storage.LoadCollections()
		}
	}

	// Create and run program
	p := tea.NewProgram(ui.NewModel(cfg), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
