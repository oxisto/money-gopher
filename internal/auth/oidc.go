package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/oxisto/money-gopher/internal/persistence"
)

const (
	oidcStateCookie = "mg_oidc_state"
	oidcNonceCookie = "mg_oidc_nonce"
)

// OIDCConfig holds the configuration for OIDC authentication.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	// RedirectURL is the full URL of /auth/callback (e.g. http://localhost:8080/auth/callback).
	RedirectURL string
}

// OIDCHandler implements the authorization-code OIDC flow.
type OIDCHandler struct {
	db       *persistence.DB
	verifier *gooidc.IDTokenVerifier
	oauth2   oauth2.Config
}

// NewOIDCHandler discovers the provider and returns an OIDCHandler. It
// contacts the IdP's discovery endpoint, so it requires network access.
func NewOIDCHandler(ctx context.Context, db *persistence.DB, cfg OIDCConfig) (*OIDCHandler, error) {
	provider, err := gooidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("OIDC provider discovery failed: %w", err)
	}

	return &OIDCHandler{
		db:       db,
		verifier: provider.Verifier(&gooidc.Config{ClientID: cfg.ClientID}),
		oauth2: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{gooidc.ScopeOpenID, "profile", "email"},
		},
	}, nil
}

// LoginHandler starts the OIDC authorization-code flow: generates state +
// nonce, stores them in short-lived cookies, and redirects to the IdP.
func (h *OIDCHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state := randomString(32)
	nonce := randomString(32)

	setTempCookie(w, oidcStateCookie, state)
	setTempCookie(w, oidcNonceCookie, nonce)

	http.Redirect(w, r, h.oauth2.AuthCodeURL(state, gooidc.Nonce(nonce)), http.StatusFound)
}

// CallbackHandler handles the IdP redirect: verifies state, exchanges the
// code for an ID token, provisions the user if new, creates a session, and
// redirects to the SPA root.
func (h *OIDCHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(oidcStateCookie)
	if err != nil || stateCookie.Value != r.FormValue("state") {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	clearTempCookie(w, oidcStateCookie)

	nonceCookie, err := r.Cookie(oidcNonceCookie)
	if err != nil {
		http.Error(w, "missing nonce cookie", http.StatusBadRequest)
		return
	}
	clearTempCookie(w, oidcNonceCookie)

	token, err := h.oauth2.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "code exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "missing id_token in response", http.StatusInternalServerError)
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(w, "id_token verification failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if idToken.Nonce != nonceCookie.Value {
		http.Error(w, "nonce mismatch", http.StatusUnauthorized)
		return
	}

	var claims struct {
		Name string `json:"name"`
	}
	_ = idToken.Claims(&claims)

	user, err := h.provisionUser(r.Context(), idToken.Issuer, idToken.Subject, claims.Name)
	if err != nil {
		http.Error(w, "user provisioning failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := CreateSession(r.Context(), h.db, w, user.ID); err != nil {
		http.Error(w, "session creation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// LogoutHandler clears the session and redirects to the login page.
func (h *OIDCHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(r.Context(), h.db, w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// provisionUser looks up the user by OIDC identity, creating them if new.
func (h *OIDCHandler) provisionUser(ctx context.Context, issuer, subject, displayName string) (*persistence.User, error) {
	user, err := h.db.GetUserByIdentity(ctx, persistence.GetUserByIdentityParams{
		Issuer:  issuer,
		Subject: subject,
	})
	if errors.Is(err, sql.ErrNoRows) {
		if displayName == "" {
			displayName = subject
		}
		return h.db.CreateUser(ctx, persistence.CreateUserParams{
			ID:          uuid.NewString(),
			Issuer:      issuer,
			Subject:     subject,
			DisplayName: displayName,
		})
	}
	return user, err
}

func setTempCookie(w http.ResponseWriter, name, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearTempCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
