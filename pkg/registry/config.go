package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// DefaultConfigFileName is the default name for the registry config file
	DefaultConfigFileName = "databases.json"
	// ConfigVersion is the current version of the config format
	ConfigVersion = "1.0"
)

// LoadConfig loads the registry configuration from the specified file
func LoadConfig(configPath string) (*RegistryConfig, error) {
	// If file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &RegistryConfig{
			Databases: make([]*DatabaseInstance, 0),
			Version:   ConfigVersion,
			UpdatedAt: time.Now(),
		}, nil
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var config RegistryConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the registry configuration to the specified file
func SaveConfig(configPath string, config *RegistryConfig) error {
	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update metadata
	config.Version = ConfigVersion
	config.UpdatedAt = time.Now()

	// Create temporary file for atomic write
	tempPath := configPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		file.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to encode config: %w", err)
	}

	file.Close()

	// Atomically replace the config file
	if err := os.Rename(tempPath, configPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

// GetDefaultConfigPath returns the default path for the config file
func GetDefaultConfigPath() string {
	return DefaultConfigFileName
}
