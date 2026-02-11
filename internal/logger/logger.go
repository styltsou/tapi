// Package logger/logger.go
package logger

import (
	"os"

	"github.com/charmbracelet/log"
)

var Logger *log.Logger

func Init() error {
	// Create/open log file
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return err
	}

	// Create logger that writes to file
	Logger = log.NewWithOptions(logFile, log.Options{
		ReportTimestamp: true,
		TimeFormat:      "15:04:05",
		Prefix:          "API Client",
	})

	Logger.Info("Logger initialized")
	return nil
}
