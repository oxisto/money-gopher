package main

import (
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/persistence"
)

var hiddenInput = regexp.MustCompile(`name="(csrf_token|return_url)" value="([^"]*)"`)

// TestBuiltinAuthFlow drives a browser through the embedded login: every
// hop stays on the one server, and only a correct password ends in a
// session that the protected routes accept.
func TestBuiltinAuthFlow(t *testing.T) {
	db, err := persistence.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("OpenDB() error = %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	mux.Handle("/graphql", auth.SessionMiddleware(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("signed in"))
	})))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("spa"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	// The server's URL is only known once it listens, and both the browser
	// and the code exchange go through it.
	mountBuiltinAuth(mux, db, srv.URL, srv.URL, "admin", "secret")

	jar, _ := cookiejar.New(nil)
	browser := &http.Client{Jar: jar}

	get := func(path string) *http.Response {
		t.Helper()
		resp, err := browser.Get(srv.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		return resp
	}
	body := func(resp *http.Response) string {
		t.Helper()
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	if resp := get("/graphql"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("before login: /graphql status = %d, want 401", resp.StatusCode)
	}

	// /auth/login → /auth/builtin/authorize → /auth/builtin/login, all on
	// srv; a hop to another port would fail the request outright.
	resp := get("/auth/login")
	if got := resp.Request.URL.Path; got != builtinAuthPrefix+"/login" {
		t.Fatalf("login redirect landed on %s, want %s/login", got, builtinAuthPrefix)
	}
	page := body(resp)
	form := url.Values{}
	for _, m := range hiddenInput.FindAllStringSubmatch(page, -1) {
		form.Set(m[1], html.UnescapeString(m[2]))
	}
	if form.Get("csrf_token") == "" || form.Get("return_url") == "" {
		t.Fatalf("login page lacks the hidden fields:\n%s", page)
	}

	// A wrong password comes back to the login page and leaves no session.
	form.Set("username", "admin")
	form.Set("password", "wrong")
	resp, err = browser.PostForm(srv.URL+builtinAuthPrefix+"/login", form)
	if err != nil {
		t.Fatal(err)
	}
	if page := body(resp); !strings.Contains(page, "Invalid credentials") {
		t.Errorf("wrong password: page does not say so:\n%s", page)
	}
	if resp := get("/graphql"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("after wrong password: /graphql status = %d, want 401", resp.StatusCode)
	}

	// The right one goes login → authorize → /auth/callback → /.
	form.Set("password", "secret")
	resp, err = browser.PostForm(srv.URL+builtinAuthPrefix+"/login", form)
	if err != nil {
		t.Fatal(err)
	}
	if got := body(resp); got != "spa" || resp.Request.URL.Path != "/" {
		t.Fatalf("after login: landed on %s with %q, want / with \"spa\"", resp.Request.URL.Path, got)
	}

	resp = get("/graphql")
	if got := body(resp); resp.StatusCode != http.StatusOK || got != "signed in" {
		t.Fatalf("after login: /graphql = %d %q, want 200 \"signed in\"", resp.StatusCode, got)
	}

	user, err := db.GetUserByIdentity(t.Context(), persistence.GetUserByIdentityParams{
		Issuer: auth.BuiltinIssuer, Subject: "admin",
	})
	if err != nil {
		t.Fatalf("user was not provisioned under %s: %v", auth.BuiltinIssuer, err)
	}
	if user.DisplayName != "admin" {
		t.Errorf("DisplayName = %q, want admin", user.DisplayName)
	}

	// Logging out drops the session again.
	if resp := get("/auth/logout"); resp.Request.URL.Path != "/login" {
		t.Errorf("logout landed on %s, want /login", resp.Request.URL.Path)
	}
	if resp := get("/graphql"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("after logout: /graphql status = %d, want 401", resp.StatusCode)
	}
}

func TestLocalURL(t *testing.T) {
	tests := map[string]string{
		":8080":            "http://127.0.0.1:8080",
		"0.0.0.0:8080":     "http://127.0.0.1:8080",
		"[::]:8080":        "http://127.0.0.1:8080",
		"localhost:9090":   "http://localhost:9090",
		"192.168.1.5:8080": "http://192.168.1.5:8080",
	}
	for addr, want := range tests {
		if got := localURL(addr); got != want {
			t.Errorf("localURL(%q) = %q, want %q", addr, got, want)
		}
	}
}
