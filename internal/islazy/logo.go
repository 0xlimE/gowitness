package islazy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FetchClearbitLogo fetches a company logo from Clearbit and saves it to the target directory
// Returns the path to the saved logo file, or an error if the fetch fails
func FetchClearbitLogo(domain, targetDir string) (string, error) {
	// Construct Clearbit logo URL
	clearbitURL := fmt.Sprintf("https://logo.clearbit.com/%s", domain)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request to Clearbit
	resp, err := client.Get(clearbitURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch logo from Clearbit: %w", err)
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Clearbit returned status %d for domain %s", resp.StatusCode, domain)
	}

	// Determine file extension from Content-Type
	contentType := resp.Header.Get("Content-Type")
	var extension string
	switch {
	case strings.Contains(contentType, "image/png"):
		extension = ".png"
	case strings.Contains(contentType, "image/jpeg"), strings.Contains(contentType, "image/jpg"):
		extension = ".jpg"
	case strings.Contains(contentType, "image/svg+xml"):
		extension = ".svg"
	default:
		// Default to png if we can't determine
		extension = ".png"
	}

	// Create logo file path
	logoPath := filepath.Join(targetDir, "logo"+extension)

	// Create the file
	out, err := os.Create(logoPath)
	if err != nil {
		return "", fmt.Errorf("failed to create logo file: %w", err)
	}
	defer out.Close()

	// Write the response body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save logo to file: %w", err)
	}

	return logoPath, nil
}
