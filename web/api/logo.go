package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/sensepost/gowitness/pkg/log"
)

// LogoHandler returns the company logo if available
//
//	@Summary		Get company logo
//	@Description	Get the company logo for the current scan session
//	@Tags			Results
//	@Produce		png
//	@Produce		jpeg
//	@Success		200	{file}		binary
//	@Failure		404	{string}	string	"Logo not found"
//	@Router			/logo [get]
func (h *ApiHandler) LogoHandler(w http.ResponseWriter, r *http.Request) {
	// The screenshot path is typically targets/<target>/screenshots/
	// We need to go up one level to find the logo in targets/<target>/
	targetDir := filepath.Dir(h.ScreenshotPath)

	// List of possible logo filenames to check
	possibleLogos := []string{
		filepath.Join(targetDir, "logo.png"),
		filepath.Join(targetDir, "logo.jpg"),
		filepath.Join(targetDir, "logo.jpeg"),
		filepath.Join(targetDir, "logo.svg"),
	}

	var logoPath string
	var found bool

	// Check each possible logo file
	for _, path := range possibleLogos {
		if _, err := os.Stat(path); err == nil {
			logoPath = path
			found = true
			break
		}
	}

	if !found {
		log.Debug("no logo file found in target directory", "target_dir", targetDir)
		http.Error(w, "Logo file not found", http.StatusNotFound)
		return
	}

	// Determine content type based on file extension
	ext := filepath.Ext(logoPath)
	var contentType string
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".svg":
		contentType = "image/svg+xml"
	default:
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, logoPath)
}
