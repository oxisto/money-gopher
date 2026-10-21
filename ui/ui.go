// Package ui serves the SvelteKit SPA from the moneyd binary. The SPA is
// built with `npm run build` into build/site, which go:embed picks up at
// compile time — so a `go build` after an npm build yields one self-contained
// binary, while a `go build` without one still compiles and serves a hint.
package ui

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// build/.gitkeep keeps this pattern valid in a fresh checkout; build/site is
// gitignored and recreated by every npm build.
//
//go:embed all:build
var embedded embed.FS

// Handler serves the embedded SPA.
func Handler() http.Handler {
	// fs.Sub doesn't check that the directory exists; without an npm build
	// every Open below fails with ErrNotExist, which NewHandler turns into
	// the hint.
	sub, err := fs.Sub(embedded, "build/site")
	if err != nil {
		panic(err) // only for an invalid path, and this one is fixed
	}
	return NewHandler(sub)
}

// NewHandler serves a SvelteKit static build from fsys: real files as they
// are, and index.html for every other path, so client-side routes survive a
// reload or a deep link. Hashed assets under _app/immutable/ are served with
// a long cache lifetime; index.html is never cached, so a new build is picked
// up on the next load.
func NewHandler(fsys fs.FS) http.Handler {
	files := http.FS(fsys)
	fileServer := http.FileServer(files)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name != "" && name != "index.html" && exists(fsys, name) {
			if strings.HasPrefix(name, "_app/immutable/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		// Client-side routes never have an extension; a missing file that
		// has one (/favicon.ico, a stale hashed asset) is a real 404, not a
		// page to render index.html for.
		if path.Ext(name) != "" && name != "index.html" {
			http.NotFound(w, r)
			return
		}

		index, err := files.Open("/index.html")
		if errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "The UI is not part of this build. Run `npm run build` in ui/ and rebuild moneyd, or use the Vite dev server (`npm run dev`).", http.StatusServiceUnavailable)
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer index.Close()

		stat, err := index.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Cache-Control", "no-cache")
		// ServeContent, not the file server: the latter would redirect
		// /index.html to / and we want the same document on every path.
		http.ServeContent(w, r, "index.html", stat.ModTime(), index)
	})
}

// exists reports whether name is a regular file in fsys. Directories don't
// count: /holdings is a client-side route, not a request for a listing.
func exists(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}
