package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/oxisto/money-gopher/internal/persistence"
)

const (
	sessionCookieName = "mg_session"
	sessionDuration   = 30 * 24 * time.Hour // 30 days
)

// CreateSession creates a new session for user and sets the session cookie.
func CreateSession(ctx context.Context, db *persistence.DB, w http.ResponseWriter, userID string) error {
	session, err := db.CreateSession(ctx, persistence.CreateSessionParams{
		ID:        uuid.NewString(),
		UserID:    userID,
		ExpiresAt: time.Now().UTC().Add(sessionDuration),
	})
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		// Secure is omitted here; set it in production behind TLS.
	})

	return nil
}

// ClearSession deletes the session from DB and clears the cookie.
func ClearSession(ctx context.Context, db *persistence.DB, w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		_ = db.DeleteSession(ctx, cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SessionMiddleware resolves the session cookie to a user and active person
// on each request. Requests without a valid session receive 401.
func SessionMiddleware(db *persistence.DB, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}

		session, err := db.GetSession(r.Context(), cookie.Value)
		if err != nil {
			http.Error(w, "session invalid or expired", http.StatusUnauthorized)
			return
		}

		user, err := db.GetUser(r.Context(), session.UserID)
		if err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}

		// Resolve the active person: use the session's stored person if set,
		// otherwise default to the first person the user can access.
		var person *persistence.Person
		if session.PersonID.Valid {
			person, err = db.GetPerson(r.Context(), session.PersonID.String)
		}
		if person == nil {
			persons, lerr := db.ListPersonsForUser(r.Context(), user.ID)
			if lerr == nil && len(persons) > 0 {
				person = persons[0]
				err = nil
			}
		}
		if err != nil || person == nil {
			http.Error(w, "no accessible person found", http.StatusUnauthorized)
			return
		}

		ctx := WithUser(r.Context(), user)
		ctx = WithPerson(ctx, person)
		ctx = WithSessionID(ctx, session.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
