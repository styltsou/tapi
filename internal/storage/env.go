package storage

import (
	"os"
	"path/filepath"

	"github.com/gosimple/slug"
	"github.com/styltsou/tapi/internal/logger"
	"gopkg.in/yaml.v3"
)

// LoadEnvironments loads all .yaml environment files from the environments directory
func LoadEnvironments() ([]Environment, error) {
	envsPath, err := GetStoragePath("environments")
	if err != nil {
		return nil, err
	}

	// Use Glob to find all .yaml files
	pattern := filepath.Join(envsPath, "*.yaml")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var envs []Environment
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			logger.Logger.Error("Failed to read environment file", "file", file, "error", err)
			continue
		}

		var env Environment
		if err := yaml.Unmarshal(data, &env); err != nil {
			logger.Logger.Error("Failed to parse environment file", "file", file, "error", err)
			continue
		}

		envs = append(envs, env)
		logger.Logger.Info("Loaded environment", "name", env.Name, "file", file)
	}

	return envs, nil
}

// SaveEnvironment saves an environment to a .yaml file in the environments directory
func SaveEnvironment(env Environment) error {
	envsPath, err := GetStoragePath("environments")
	if err != nil {
		return err
	}

	// Sanitize name for filename
	filename := filepath.Join(envsPath, slug.Make(env.Name)+".yaml")

	data, err := yaml.Marshal(env)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return err
	}

	logger.Logger.Info("Saved environment", "name", env.Name, "file", filename)
	return nil
}

// DeleteEnvironment deletes an environment file
func DeleteEnvironment(name string) error {
	envsPath, err := GetStoragePath("environments")
	if err != nil {
		return err
	}

	filename := filepath.Join(envsPath, slug.Make(name)+".yaml")
	
	if err := os.Remove(filename); err != nil {
		return err
	}
	
	logger.Logger.Info("Deleted environment", "name", name, "file", filename)
	return nil
}
