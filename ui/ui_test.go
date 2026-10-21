package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestNewHandler(t *testing.T) {
	fsys := fstest.MapFS{
		"index.html":                  {Data: []byte("<!doctype html><title>Money Gopher</title>")},
		"robots.txt":                  {Data: []byte("User-agent: *")},
		"_app/immutable/entry/app.js": {Data: []byte("console.log('app')")},
		"_app/version.json":           {Data: []byte(`{"version":"1"}`)},
		"holdings/.keep":              {Data: []byte("")}, // makes "holdings" a directory
	}
	h := NewHandler(fsys)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{"root serves index", "/", 200, "Money Gopher", "no-cache"},
		{"real file is served as is", "/robots.txt", 200, "User-agent", ""},
		{"hashed asset is cached for a year", "/_app/immutable/entry/app.js", 200, "console.log", "public, max-age=31536000, immutable"},
		{"non-hashed app file is not cached long", "/_app/version.json", 200, "version", ""},
		{"client route falls back to index", "/holdings/abc", 200, "Money Gopher", "no-cache"},
		{"directory falls back to index, not a listing", "/holdings", 200, "Money Gopher", "no-cache"},
		{"index.html itself is not redirected", "/index.html", 200, "Money Gopher", "no-cache"},
		{"unknown deep path falls back to index", "/no/such/thing", 200, "Money Gopher", "no-cache"},
		{"missing file with an extension is a 404", "/favicon.ico", 404, "not found", ""},
		{"stale hashed asset is a 404, not index", "/_app/immutable/entry/gone.js", 404, "not found", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Errorf("Cache-Control = %q, want %q", got, tt.wantCache)
			}
		})
	}

	t.Run("POST is rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", rec.Code)
		}
	})
}

// TestNewHandler_NotBuilt covers a binary compiled without an npm build:
// build/ holds only .gitkeep, so there is no index.html to fall back to.
func TestNewHandler_NotBuilt(t *testing.T) {
	h := NewHandler(fstest.MapFS{".gitkeep": {Data: []byte("")}})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "npm run build") {
		t.Errorf("body = %q, want a hint to build the UI", rec.Body.String())
	}
}
