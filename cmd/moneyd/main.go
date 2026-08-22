// moneyd is the Money Gopher server. It serves the GraphQL API under
// /graphql and an interactive GraphiQL explorer under /graphiql. In a later
// milestone it also serves the embedded SvelteKit SPA.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	oauth2go "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"

	"github.com/oxisto/money-gopher/internal/api"
	"github.com/oxisto/money-gopher/internal/auth"
	"github.com/oxisto/money-gopher/internal/importer"
	"github.com/oxisto/money-gopher/internal/persistence"
	"github.com/oxisto/money-gopher/internal/quotes"

	_ "modernc.org/sqlite"
)

var (
	addr             = flag.String("addr", ":8080", "the address to listen on")
	publicURL        = flag.String("public-url", "http://localhost:8080", "public base URL of moneyd (used to build OIDC redirect URLs)")
	dbPath           = flag.String("db", "moneyd-data", "path to the lightsql data directory (created if missing)")
	importSQLite     = flag.String("import-sqlite", "", "path to a pre-lightsql SQLite database to copy into -db at startup; a one-time cutover, since lightsql now persists to disk (re-running it against an already-migrated -db fails on the first duplicate primary key)")
	quoteInterval    = flag.Duration("quote-interval", time.Hour, "how often to refresh quotes; 0 disables the background refresh")
	authMode         = flag.String("auth", "dev", `authentication mode: "dev" (fixed dev user), "builtin" (embedded auth server), or "oidc"`)
	builtinAuthAddr  = flag.String("builtin-auth-addr", ":8081", "address for the embedded auth server (--auth=builtin)")
	builtinUser      = flag.String("auth-user", "admin", "username for the embedded auth server (--auth=builtin)")
	builtinPassword  = flag.String("auth-password", "", "password for the embedded auth server (--auth=builtin, required)")
	oidcIssuer       = flag.String("oidc-issuer", "", "OIDC issuer URL (required when --auth=oidc)")
	oidcClientID     = flag.String("oidc-client-id", "", "OIDC client ID")
	oidcClientSecret = flag.String("oidc-client-secret", "", "OIDC client secret")
)

func main() {
	flag.Parse()

	if err := run(*addr, *dbPath, *quoteInterval); err != nil {
		slog.Error("moneyd failed", "err", err)
		os.Exit(1)
	}
}

func run(addr string, dbPath string, quoteInterval time.Duration) error {
	db, err := persistence.OpenDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if *importSQLite != "" {
		n, err := importSQLiteData(context.Background(), *importSQLite, dbPath)
		if err != nil {
			return fmt.Errorf("import from %q: %w", *importSQLite, err)
		}
		slog.Info("imported data from SQLite", "source", *importSQLite, "rows", n)
	}

	updater := &quotes.Updater{DB: db, Registry: quotes.DefaultRegistry()}
	if quoteInterval > 0 {
		go updater.Run(context.Background(), quoteInterval)
	}

	svc := importer.NewService(db)
	mux := http.NewServeMux()

	var protect func(http.Handler) http.Handler

	switch *authMode {
	case "builtin":
		if *builtinPassword == "" {
			return fmt.Errorf("--auth-password is required when --auth=builtin")
		}
		clientSecret := oauth2go.GenerateSecret()
		redirectURL := *publicURL + "/auth/callback"
		builtinIssuer := "http://localhost" + *builtinAuthAddr

		authServer := oauth2go.NewServer(*builtinAuthAddr,
			oauth2go.WithPublicURL(builtinIssuer),
			oauth2go.WithClient("moneyd", clientSecret, redirectURL),
			login.WithLoginPage(login.WithUser(*builtinUser, *builtinPassword)),
		)
		go func() {
			slog.Info("builtin auth server listening", "addr", *builtinAuthAddr)
			if err := authServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("builtin auth server failed", "err", err)
			}
		}()

		builtinHandler := auth.NewBuiltinHandler(db, auth.BuiltinConfig{
			Issuer:       builtinIssuer,
			ClientID:     "moneyd",
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			PublicKeys:   authServer.PublicKeys,
		})

		mux.HandleFunc("/auth/login", builtinHandler.LoginHandler)
		mux.HandleFunc("/auth/callback", builtinHandler.CallbackHandler)
		mux.HandleFunc("/auth/logout", builtinHandler.LogoutHandler)

		protect = func(h http.Handler) http.Handler {
			return auth.SessionMiddleware(db, h)
		}
		slog.Info("builtin auth enabled", "user", *builtinUser, "auth-addr", *builtinAuthAddr)

	case "oidc":
		if *oidcIssuer == "" || *oidcClientID == "" {
			return fmt.Errorf("--oidc-issuer and --oidc-client-id are required when --auth=oidc")
		}
		redirectURL := *publicURL + "/auth/callback"
		oidcHandler, err := auth.NewOIDCHandler(context.Background(), db, auth.OIDCConfig{
			Issuer:       *oidcIssuer,
			ClientID:     *oidcClientID,
			ClientSecret: *oidcClientSecret,
			RedirectURL:  redirectURL,
		})
		if err != nil {
			return err
		}

		mux.HandleFunc("/auth/login", oidcHandler.LoginHandler)
		mux.HandleFunc("/auth/callback", oidcHandler.CallbackHandler)
		mux.HandleFunc("/auth/logout", oidcHandler.LogoutHandler)

		protect = func(h http.Handler) http.Handler {
			return auth.SessionMiddleware(db, h)
		}
		slog.Info("OIDC auth enabled", "issuer", *oidcIssuer)

	default: // "dev"
		user, person, err := auth.EnsureDevUser(context.Background(), db)
		if err != nil {
			return err
		}
		protect = func(h http.Handler) http.Handler {
			return auth.DevMiddleware(user, person, h)
		}
		// Stub auth routes so the UI's "Sign out" link doesn't 404 in dev mode.
		mux.HandleFunc("/auth/logout", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/", http.StatusFound)
		})
		slog.Info("dev auth enabled; all requests pinned to dev user")
	}

	mux.Handle("/graphql", protect(api.NewHandler(api.NewSchema(db, updater))))
	mux.Handle("/upload", protect(uploadHandler(svc)))
	mux.Handle("/documents/", protect(documentHandler(db)))
	mux.HandleFunc("/graphiql", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(graphiql)
	})

	slog.Info("moneyd listening", "addr", addr, "db", dbPath)

	return http.ListenAndServe(addr, mux)
}

// uploadHandler handles POST /upload: stores a document and asynchronously
// runs the import pipeline. The request must be multipart/form-data with a
// "file" field. It responds 201 with {"id":"<docID>"} on success.
func uploadHandler(svc *importer.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		person, err := auth.PersonFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		f, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "missing file field: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer f.Close()

		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, "read error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		doc, err := svc.Upload(r.Context(), person.ID, header.Filename, contentType, data)
		if err != nil {
			http.Error(w, "upload failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Process asynchronously so the response returns immediately.
		go func() {
			if err := svc.Process(context.Background(), doc.ID, person.ID); err != nil {
				slog.Error("import pipeline failed", "doc", doc.ID, "err", err)
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = fmt.Fprintf(w, `{"id":%q}`, doc.ID)
	})
}

// documentHandler serves GET /documents/{id} — the raw bytes of a stored
// document (typically a PDF) so the browser can display it inline.
func documentHandler(db *persistence.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		person, err := auth.PersonFromContext(r.Context())
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// URL pattern: /documents/{id}
		id := r.URL.Path[len("/documents/"):]
		if id == "" {
			http.Error(w, "missing document id", http.StatusBadRequest)
			return
		}

		doc, err := db.GetDocument(r.Context(), persistence.GetDocumentParams{
			ID:       id,
			PersonID: person.ID,
		})
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		ct := doc.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		w.Header().Set("Content-Type", ct)
		w.Header().Set("Content-Disposition", "inline")
		_, _ = w.Write(doc.Data)
	})
}

// graphiql is a minimal GraphiQL page loading the explorer from a CDN and
// pointing it at /graphql.
var graphiql = []byte(`<!doctype html>
<html lang="en">
	<head>
		<title>GraphiQL — Money Gopher</title>
		<style>body { margin: 0; } #graphiql { height: 100vh; }</style>
		<script crossorigin src="https://unpkg.com/react@18/umd/react.production.min.js"></script>
		<script crossorigin src="https://unpkg.com/react-dom@18/umd/react-dom.production.min.js"></script>
		<link rel="stylesheet" href="https://unpkg.com/graphiql/graphiql.min.css" />
	</head>
	<body>
		<div id="graphiql">Loading…</div>
		<script crossorigin src="https://unpkg.com/graphiql/graphiql.min.js"></script>
		<script>
			ReactDOM.createRoot(document.getElementById('graphiql')).render(
				React.createElement(GraphiQL, {
					fetcher: GraphiQL.createFetcher({ url: '/graphql' }),
				})
			);
		</script>
	</body>
</html>
`)

// tablesInFKOrder lists tables in an order safe to INSERT into: every table
// appears after every table its foreign keys reference. sessions is
// intentionally excluded — session tokens are short-lived and tied to a
// specific server process, so users just log in again after importing.
var tablesInFKOrder = []string{
	"users",
	"persons",
	"user_person_access",
	"securities",
	"security_identifiers",
	"listings",
	"cash_accounts",
	"portfolios",
	"quotes",
	"transactions",
	"documents",
	"staged_transactions",
}

// importSQLiteData copies every row from a pre-lightsql SQLite database at
// sqlitePath into the lightsql database at dbPath, which must already exist
// (migrations applied) in this same process — lightsql instances are
// process-local, so this cannot be done from a separate tool. Meant as a
// one-time cutover: lightsql persists to disk now, so re-running this
// against an already-migrated directory fails loudly on the first duplicate
// primary key rather than silently re-seeding.
func importSQLiteData(ctx context.Context, sqlitePath, dbPath string) (int, error) {
	src, err := sql.Open("sqlite", sqlitePath)
	if err != nil {
		return 0, fmt.Errorf("open sqlite database: %w", err)
	}
	defer src.Close()

	// Reaches the same engine persistence.OpenDB already opened: lightsql's
	// instance registry is keyed by resolved directory path, and pooled
	// connections to one instance are the expected way to use it.
	dst, err := sql.Open("lightsql", "file:"+dbPath)
	if err != nil {
		return 0, fmt.Errorf("open lightsql database: %w", err)
	}
	defer dst.Close()

	total := 0
	for _, table := range tablesInFKOrder {
		n, err := copySQLiteTable(ctx, src, dst, table)
		if err != nil {
			return total, fmt.Errorf("copy table %q: %w", table, err)
		}
		total += n
	}

	return total, nil
}

// copySQLiteTable reads every row of table from src and inserts it into dst,
// column-for-column by name — the source and destination schemas are
// identical, so no per-table mapping is needed.
func copySQLiteTable(ctx context.Context, src, dst *sql.DB, table string) (int, error) {
	rows, err := src.QueryContext(ctx, "SELECT * FROM "+table)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	insert := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	n := 0
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return n, fmt.Errorf("scan row %d: %w", n, err)
		}

		if _, err := dst.ExecContext(ctx, insert, values...); err != nil {
			return n, fmt.Errorf("insert row %d: %w", n, err)
		}
		n++
	}

	return n, rows.Err()
}
