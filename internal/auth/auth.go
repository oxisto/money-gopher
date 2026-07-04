// Package auth provides authentication for moneyd. For now it only contains
// the dev mode, which pins every request to a local development user; real
// OIDC support arrives in a later milestone.
package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/oxisto/money-gopher/internal/persistence"
)

type userKey struct{}

// ErrNotAuthenticated is returned when a request carries no user.
var ErrNotAuthenticated = errors.New("not authenticated")

// WithUser returns a context carrying the given user.
func WithUser(ctx context.Context, user *persistence.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// UserFromContext returns the authenticated user of the request, or
// ErrNotAuthenticated if there is none.
func UserFromContext(ctx context.Context) (*persistence.User, error) {
	user, ok := ctx.Value(userKey{}).(*persistence.User)
	if !ok {
		return nil, ErrNotAuthenticated
	}

	return user, nil
}

// Middleware attaches the user resolved by lookup to each request. lookup is
// called per request so that user data is always current.
func Middleware(lookup func(r *http.Request) (*persistence.User, error), next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := lookup(r)
		if err != nil {
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	})
}

const (
	devIssuer  = "urn:money-gopher:dev"
	devSubject = "dev"
)

// EnsureDevUser returns the development user, creating it on first use. Dev
// mode pins all requests to this user so moneyd runs without an IdP.
func EnsureDevUser(ctx context.Context, db *persistence.DB) (*persistence.User, error) {
	user, err := db.GetUserByIdentity(ctx, persistence.GetUserByIdentityParams{
		Issuer:  devIssuer,
		Subject: devSubject,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return db.CreateUser(ctx, persistence.CreateUserParams{
			ID:          "dev",
			Issuer:      devIssuer,
			Subject:     devSubject,
			DisplayName: "Development User",
		})
	}

	return user, err
}

// DevMiddleware pins every request to the development user.
func DevMiddleware(user *persistence.User, next http.Handler) http.Handler {
	return Middleware(func(*http.Request) (*persistence.User, error) {
		return user, nil
	}, next)
}
