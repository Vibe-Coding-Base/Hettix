// Package adminui embeds the built admin single-page app so every frontend
// (the headless HTTP server and the desktop shell) serves identical assets.
//
// The admin directory is produced by the frontend build (`make build-admin`)
// and is not committed; a build fails until it has been generated.
package adminui

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// The `all:` prefix embeds every file, including any whose name starts with
// "." or "_" that a plain embed would skip.
//
//go:embed all:admin
var content embed.FS

// FS returns the admin SPA file system rooted at index.html.
func FS() (fs.FS, error) {
	return fs.Sub(content, "admin")
}

// Handler serves the admin SPA, falling back to index.html for any path that
// doesn't resolve to a file. The admin UI uses client-side routing, so deep
// links like /proxy/logs have no file of their own and must be handed the SPA
// entrypoint.
func Handler() (http.Handler, error) {
	fsys, err := FS()
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}

		if _, err := fs.Stat(fsys, name); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	}), nil
}
