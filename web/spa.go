package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

//go:embed ui/dist/*
var ui embed.FS

// SpaHandler handles request to the SPA
func SpaHandler() http.HandlerFunc {
	spaFS, err := fs.Sub(ui, "ui/dist")
	if err != nil {
		panic(fmt.Errorf("failed getting the sub tree for the site files: %w", err))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(path.Clean(r.URL.Path), "/")

		f, err := spaFS.Open(path)
		if err == nil {
			defer f.Close()
		}

		if os.IsNotExist(err) {
			// For any path that doesn't exist, serve the index.html file
			// This handles SPA routing where /screenshot/1, /overview, etc. should all serve index.html
			r.URL.Path = "/"
		}

		http.FileServer(http.FS(spaFS)).ServeHTTP(w, r)
	}
}
