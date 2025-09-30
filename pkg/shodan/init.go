package shodan

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// InitFromEnv initializes a Shodan client from environment variables
// It attempts to load from .env file first, then falls back to system environment
func InitFromEnv() (*Client, error) {
	// Try to load .env file (ignore errors as it may not exist)
	_ = godotenv.Load()

	apiKey := os.Getenv("SHODAN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SHODAN_API_KEY environment variable is required")
	}

	client := NewClient(apiKey)

	// Validate the API key
	if err := client.IsValidAPIKey(); err != nil {
		return nil, fmt.Errorf("failed to validate Shodan API key: %w", err)
	}

	return client, nil
}
