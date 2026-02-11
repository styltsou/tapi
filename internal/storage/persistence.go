package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/styltsou/tapi/internal/logger"
	"gopkg.in/yaml.v3"
)

func GetStoragePath(subDir string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	storagePath := filepath.Join(homeDir, ".tapi", subDir)

	return storagePath, nil
}

func EnsureDir(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		// Log more details about the error
		logger.Logger.Error("MkdirAll failed",
			"path", path,
			"error", err,
			"errorType", fmt.Sprintf("%T", err))
	}
	return err
}

// LoadCollections loads all .yaml collection files from the collections directory
func LoadCollections() ([]Collection, error) {
	// Get path to collections directory
	collectionsPath, err := GetStoragePath("collections")
	if err != nil {
		return nil, err
	}

	// Ensure directory exists
	//	if err := EnsureDir(collectionsPath); err != nil {
	//
	//return nil, err
	//}

	// Find all .yaml files
	pattern := filepath.Join(collectionsPath, "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var collections []Collection

	// Iterate through files
	for _, file := range files {
		// Read file
		data, err := os.ReadFile(file)
		if err != nil {
			logger.Logger.Error("Failed to read collection file", "file", file, "error", err)
			continue // Skip this file and continue
		}

		// Unmarshal YAML
		var collection Collection
		if err := yaml.Unmarshal(data, &collection); err != nil {
			logger.Logger.Error("Failed to parse collection file", "file", file, "error", err)
			continue // Skip corrupt file
		}

		// Derive collection name from filename
		// my-api.yaml -> "My Api"
		filename := filepath.Base(file)
		nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))
		collection.Name = formatName(nameWithoutExt)

		collections = append(collections, collection)
		logger.Logger.Info("Loaded collection", "name", collection.Name, "file", file)
	}

	return collections, nil
}

// formatName converts a filename to a human-readable name
// my-api -> "My Api"
// my_api_v2 -> "My Api V2"
func formatName(filename string) string {
	// Replace dashes and underscores with spaces
	name := strings.NewReplacer("-", " ", "_", " ").Replace(filename)

	// Title case each word
	words := strings.Fields(name)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return strings.Join(words, " ")
}

// SaveCollection saves a collection to disk atomically
func SaveCollection(c Collection) error {
	// Get collections directory path
	collectionsPath, err := GetStoragePath("collections")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := EnsureDir(collectionsPath); err != nil {
		return err
	}

	// Determine filename from collection name
	filename := slug.Make(c.Name) + ".yaml"
	finalPath := filepath.Join(collectionsPath, filename)
	tempPath := finalPath + ".tmp"

	// Marshal collection to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		logger.Logger.Error("Failed to marshal collection", "name", c.Name, "error", err)
		return err
	}

	// Write to temporary file first
	if err := os.WriteFile(tempPath, data, 0o644); err != nil {
		logger.Logger.Error("Failed to write temp file", "path", tempPath, "error", err)
		return err
	}

	// Atomically rename temp file to final file
	if err := os.Rename(tempPath, finalPath); err != nil {
		logger.Logger.Error("Failed to rename temp file", "temp", tempPath, "final", finalPath, "error", err)
		// Clean up temp file on failure
		os.Remove(tempPath)
		return err
	}

	logger.Logger.Info("Saved collection", "name", c.Name, "file", finalPath)
	return nil
}

// DeleteCollection deletes a collection from disk
func DeleteCollection(name string) error {
	collectionsPath, err := GetStoragePath("collections")
	if err != nil {
		return err
	}

	filename := slug.Make(name) + ".yaml"
	finalPath := filepath.Join(collectionsPath, filename)

	if err := os.Remove(finalPath); err != nil {
		return err
	}

	logger.Logger.Info("Deleted collection", "name", name, "file", finalPath)
	return nil
}

// ExportCollection copies a collection YAML file to the given destination path.
func ExportCollection(name, destPath string) error {
	collectionsPath, err := GetStoragePath("collections")
	if err != nil {
		return err
	}

	srcFile := filepath.Join(collectionsPath, slug.Make(name)+".yaml")
	data, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("collection %q not found: %w", name, err)
	}

	if err := os.WriteFile(destPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write export: %w", err)
	}

	logger.Logger.Info("Exported collection", "name", name, "dest", destPath)
	return nil
}
