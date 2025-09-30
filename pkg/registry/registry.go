package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// NewDatabaseRegistry creates a new database registry instance
func NewDatabaseRegistry(configPath string) (*DatabaseRegistry, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load registry config: %w", err)
	}

	registry := &DatabaseRegistry{
		instances:  make(map[string]*DatabaseInstance),
		configPath: configPath,
	}

	// Load existing database instances
	for _, instance := range config.Databases {
		registry.instances[instance.UUID] = instance
	}

	return registry, nil
}

// Add creates a new database instance with the given name
func (r *DatabaseRegistry) Add(name string) (*DatabaseInstance, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Generate new UUID
	newUUID := uuid.New().String()

	// Create folder structure
	basePath := "databases"
	folderPath := filepath.Join(basePath, newUUID)
	databasePath := filepath.Join(folderPath, "database.db")
	screenshotDir := filepath.Join(folderPath, "screenshots")

	// Create directories
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database folder: %w", err)
	}

	if err := os.MkdirAll(screenshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create screenshots folder: %w", err)
	}

	// Create database instance
	instance := &DatabaseInstance{
		UUID:          newUUID,
		Name:          name,
		FolderPath:    folderPath,
		DatabasePath:  databasePath,
		ScreenshotDir: screenshotDir,
		CreatedAt:     time.Now(),
		IsActive:      true,
	}

	// Add to registry
	r.instances[newUUID] = instance

	// Save to config
	if err := r.saveConfig(); err != nil {
		// Rollback: remove from memory and filesystem
		delete(r.instances, newUUID)
		os.RemoveAll(folderPath)
		return nil, fmt.Errorf("failed to save config: %w", err)
	}

	return instance, nil
}

// Get retrieves a database instance by UUID
func (r *DatabaseRegistry) Get(uuid string) (*DatabaseInstance, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instance, exists := r.instances[uuid]
	return instance, exists
}

// List returns all database instances
func (r *DatabaseRegistry) List() []*DatabaseInstance {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	instances := make([]*DatabaseInstance, 0, len(r.instances))
	for _, instance := range r.instances {
		instances = append(instances, instance)
	}

	return instances
}

// Remove removes a database instance and its folder
func (r *DatabaseRegistry) Remove(uuid string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.instances[uuid]
	if !exists {
		return fmt.Errorf("database with UUID %s not found", uuid)
	}

	// Remove from filesystem
	if err := os.RemoveAll(instance.FolderPath); err != nil {
		return fmt.Errorf("failed to remove database folder: %w", err)
	}

	// Remove from memory
	delete(r.instances, uuid)

	// Save to config
	if err := r.saveConfig(); err != nil {
		return fmt.Errorf("failed to save config after removal: %w", err)
	}

	return nil
}

// SetActive sets the active status of a database instance
func (r *DatabaseRegistry) SetActive(uuid string, active bool) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	instance, exists := r.instances[uuid]
	if !exists {
		return fmt.Errorf("database with UUID %s not found", uuid)
	}

	instance.IsActive = active

	return r.saveConfig()
}

// saveConfig saves the current state to the config file
func (r *DatabaseRegistry) saveConfig() error {
	instances := make([]*DatabaseInstance, 0, len(r.instances))
	for _, instance := range r.instances {
		instances = append(instances, instance)
	}

	config := &RegistryConfig{
		Databases: instances,
	}

	return SaveConfig(r.configPath, config)
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// GenerateUUID generates a new UUID string
func GenerateUUID() string {
	return uuid.New().String()
}
