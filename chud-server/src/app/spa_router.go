package app

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/sebnow/chud/platform/constants"
)

func newSPAHandler(staticFiles embed.FS) http.Handler {
	staticRoot, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic("Failed to read embedded static files: " + err.Error())
	}

	fileServer := http.FileServer(http.FS(staticRoot))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		requestPath := path.Clean("/" + r.URL.Path)
		relativePath := strings.TrimPrefix(requestPath, "/")

		if strings.HasPrefix(relativePath, "assets/") {
			w.Header().Set(constants.HTTPHeaderCacheControl, "public, max-age=31536000, immutable")
		} else if relativePath != "" && relativePath != "." {
			w.Header().Set(constants.HTTPHeaderCacheControl, "no-cache, no-store, must-revalidate")
		}

		if relativePath == "" || relativePath == "." || relativePath == "index.html" {
			w.Header().Set(constants.HTTPHeaderContentType, "text/html; charset=utf-8")
			fileServer.ServeHTTP(w, r)
			return
		}

		if file, err := staticRoot.Open(relativePath); err == nil {
			file.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(relativePath, ".") {
			w.Header().Set(constants.HTTPHeaderContentType, "text/html; charset=utf-8")
			w.Header().Set(constants.HTTPHeaderCacheControl, "no-cache, no-store, must-revalidate")
			http.ServeFileFS(w, r, staticRoot, "index.html")
			return
		}

		http.NotFound(w, r)
	})
}
