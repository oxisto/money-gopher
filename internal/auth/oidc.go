package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/oxisto/money-gopher/internal/persistence"
)

const (
	oidcStateCookie = "mg_oidc_state"
	oidcNonceCookie = "mg_oidc_nonce"
)

// idTokenSigningAlgorithms are the JWT signing algorithms accepted for an
// id_token. Restricting to this allowlist (rather than trusting whatever
// "alg" the token header claims) prevents algorithm-confusion attacks, e.g.
// a token forged with alg=HS256 using the provider's public RSA/EC key
// reinterpreted as an HMAC secret.
var idTokenSigningAlgorithms = []string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512", "PS256", "PS384", "PS512"}

// OIDCConfig holds the configuration for OIDC authentication.
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	// RedirectURL is the full URL of /auth/callback (e.g. http://localhost:8080/auth/callback).
	RedirectURL string
}

// oidcDiscovery is the subset of an OIDC discovery document
// (/.well-known/openid-configuration) that we need.
type oidcDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// OIDCHandler implements the authorization-code OIDC flow.
type OIDCHandler struct {
	db     *persistence.DB
	issuer string
	jwks   jwkset.Storage
	oauth2 oauth2.Config
}

// NewOIDCHandler discovers the provider and returns an OIDCHandler. It
// contacts the IdP's discovery and JWKS endpoints, so it requires network access.
func NewOIDCHandler(ctx context.Context, db *persistence.DB, cfg OIDCConfig) (*OIDCHandler, error) {
	disc, err := discoverOIDC(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("OIDC provider discovery failed: %w", err)
	}
	if disc.Issuer != cfg.Issuer {
		return nil, fmt.Errorf("discovery document issuer %q does not match configured issuer %q", disc.Issuer, cfg.Issuer)
	}

	jwks, err := jwkset.NewDefaultHTTPClientCtx(ctx, []string{disc.JWKSURI})
	if err != nil {
		return nil, fmt.Errorf("fetching JWKS failed: %w", err)
	}

	return &OIDCHandler{
		db:     db,
		issuer: cfg.Issuer,
		jwks:   jwks,
		oauth2: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  disc.AuthorizationEndpoint,
				TokenURL: disc.TokenEndpoint,
			},
			Scopes: []string{"openid", "profile", "email"},
		},
	}, nil
}

// discoverOIDC fetches and parses the provider's discovery document.
func discoverOIDC(ctx context.Context, issuer string) (*oidcDiscovery, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		strings.TrimRight(issuer, "/")+"/.well-known/openid-configuration", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var disc oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&disc); err != nil {
		return nil, err
	}
	if disc.Issuer == "" || disc.AuthorizationEndpoint == "" || disc.TokenEndpoint == "" || disc.JWKSURI == "" {
		return nil, errors.New("discovery document missing required fields")
	}
	return &disc, nil
}

// keyfunc resolves the public key for a JWT by its "kid" header, looked up
// from the provider's JWKS (auto-refreshed hourly by the jwkset client).
func (h *OIDCHandler) keyfunc(ctx context.Context) jwt.Keyfunc {
	return func(t *jwt.Token) (any, error) {
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token missing kid header")
		}
		key, err := h.jwks.KeyRead(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("unknown signing key %q: %w", kid, err)
		}
		return key.Key(), nil
	}
}

// LoginHandler starts the OIDC authorization-code flow: generates state +
// nonce, stores them in short-lived cookies, and redirects to the IdP.
func (h *OIDCHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state := randomString(32)
	nonce := randomString(32)

	setTempCookie(w, oidcStateCookie, state)
	setTempCookie(w, oidcNonceCookie, nonce)

	authURL := h.oauth2.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
	http.Redirect(w, r, authURL, http.StatusFound)
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

	var claims jwt.MapClaims
	_, err = jwt.ParseWithClaims(rawIDToken, &claims, h.keyfunc(r.Context()),
		jwt.WithValidMethods(idTokenSigningAlgorithms),
		jwt.WithIssuer(h.issuer),
		jwt.WithAudience(h.oauth2.ClientID),
	)
	if err != nil {
		http.Error(w, "id_token verification failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if nonce, _ := claims["nonce"].(string); nonce != nonceCookie.Value {
		http.Error(w, "nonce mismatch", http.StatusUnauthorized)
		return
	}

	subject, _ := claims["sub"].(string)
	if subject == "" {
		http.Error(w, "id_token missing sub claim", http.StatusUnauthorized)
		return
	}
	name, _ := claims["name"].(string)

	user, err := provisionUser(r.Context(), h.db, h.issuer, subject, name)
	if err != nil {
		http.Error(w, "user provisioning failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := EnsurePersonForUser(r.Context(), h.db, user); err != nil {
		http.Error(w, "person provisioning failed: "+err.Error(), http.StatusInternalServerError)
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

// provisionUser looks up the user by (issuer, subject) identity, creating
// them if new. Shared by the external OIDC flow and the embedded builtin
// auth flow, which identify users differently but both land here.
func provisionUser(ctx context.Context, db *persistence.DB, issuer, subject, displayName string) (*persistence.User, error) {
	user, err := db.GetUserByIdentity(ctx, persistence.GetUserByIdentityParams{
		Issuer:  issuer,
		Subject: subject,
	})
	if errors.Is(err, sql.ErrNoRows) {
		if displayName == "" {
			displayName = subject
		}
		return db.CreateUser(ctx, persistence.CreateUserParams{
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
