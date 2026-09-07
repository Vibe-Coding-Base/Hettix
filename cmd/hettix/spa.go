package main

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// spaFileServer serves static files from fsys, falling back to index.html for
// any path that doesn't resolve to a file. The admin UI is a single-page app
// with client-side routing, so deep links like /proxy/logs have no file of
// their own and must be handed the SPA entrypoint.
func spaFileServer(fsys fs.FS) http.Handler {
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
	})
}
