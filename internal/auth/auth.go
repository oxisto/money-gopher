package auth

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/oxisto/money-gopher/internal/persistence"
)

type userKey struct{}
type personKey struct{}
type sessionIDKey struct{}

// ErrNotAuthenticated is returned when a request carries no user.
var ErrNotAuthenticated = errors.New("not authenticated")

// WithUser returns a context carrying the given user.
func WithUser(ctx context.Context, user *persistence.User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

// UserFromContext returns the authenticated user of the request.
func UserFromContext(ctx context.Context) (*persistence.User, error) {
	user, ok := ctx.Value(userKey{}).(*persistence.User)
	if !ok {
		return nil, ErrNotAuthenticated
	}
	return user, nil
}

// WithPerson returns a context carrying the active person (the financial
// entity whose data the current request operates on).
func WithPerson(ctx context.Context, person *persistence.Person) context.Context {
	return context.WithValue(ctx, personKey{}, person)
}

// PersonFromContext returns the active person for the request.
func PersonFromContext(ctx context.Context) (*persistence.Person, error) {
	person, ok := ctx.Value(personKey{}).(*persistence.Person)
	if !ok {
		return nil, ErrNotAuthenticated
	}
	return person, nil
}

// WithSessionID stores the session ID in the context so mutations can update it.
func WithSessionID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, id)
}

// SessionIDFromContext returns the session ID for the current request.
func SessionIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(sessionIDKey{}).(string)
	return id, ok
}

// Middleware attaches the user resolved by lookup to each request.
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
	devID      = "dev"
)

// EnsureDevUser returns the development user and its associated person,
// creating both on first use.
func EnsureDevUser(ctx context.Context, db *persistence.DB) (*persistence.User, *persistence.Person, error) {
	user, err := db.GetUserByIdentity(ctx, persistence.GetUserByIdentityParams{
		Issuer:  devIssuer,
		Subject: devSubject,
	})
	if errors.Is(err, sql.ErrNoRows) {
		user, err = db.CreateUser(ctx, persistence.CreateUserParams{
			ID:          devID,
			Issuer:      devIssuer,
			Subject:     devSubject,
			DisplayName: "Development User",
		})
	}
	if err != nil {
		return nil, nil, err
	}

	person, err := db.GetPerson(ctx, devID)
	if errors.Is(err, sql.ErrNoRows) {
		person, err = db.CreatePerson(ctx, persistence.CreatePersonParams{
			ID:          devID,
			DisplayName: "Development User",
		})
		if err != nil {
			return nil, nil, err
		}
		err = db.GrantPersonAccess(ctx, persistence.GrantPersonAccessParams{
			UserID:   devID,
			PersonID: devID,
		})
	}
	if err != nil {
		return nil, nil, err
	}

	return user, person, nil
}

// EnsurePersonForUser returns a person the user can access, creating a
// default one (named after the user) on their first authenticated request.
func EnsurePersonForUser(ctx context.Context, db *persistence.DB, user *persistence.User) (*persistence.Person, error) {
	persons, err := db.ListPersonsForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if len(persons) > 0 {
		return persons[0], nil
	}

	person, err := db.CreatePerson(ctx, persistence.CreatePersonParams{
		ID:          uuid.NewString(),
		DisplayName: user.DisplayName,
	})
	if err != nil {
		return nil, err
	}
	if err := db.GrantPersonAccess(ctx, persistence.GrantPersonAccessParams{
		UserID:   user.ID,
		PersonID: person.ID,
	}); err != nil {
		return nil, err
	}
	return person, nil
}

// DevMiddleware pins every request to the development user and their default person.
func DevMiddleware(user *persistence.User, person *persistence.Person, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := WithUser(r.Context(), user)
		ctx = WithPerson(ctx, person)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
