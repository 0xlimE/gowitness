package registry

import (
	"sync"
	"time"
)

// DatabaseInstance represents a single database instance with its metadata
type DatabaseInstance struct {
	UUID          string    `json:"uuid"`
	Name          string    `json:"name"`
	FolderPath    string    `json:"folder_path"`    // Path to databases/{uuid}/
	DatabasePath  string    `json:"database_path"`  // Path to databases/{uuid}/database.db
	ScreenshotDir string    `json:"screenshot_dir"` // Path to databases/{uuid}/screenshots/
	CreatedAt     time.Time `json:"created_at"`
	IsActive      bool      `json:"is_active"`
}

// DatabaseRegistry manages multiple database instances in a thread-safe manner
type DatabaseRegistry struct {
	instances  map[string]*DatabaseInstance
	mutex      sync.RWMutex
	configPath string
}

// RegistryConfig represents the configuration file structure
type RegistryConfig struct {
	Databases []*DatabaseInstance `json:"databases"`
	Version   string              `json:"version"`
	UpdatedAt time.Time           `json:"updated_at"`
}
