package auth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"github.com/oxisto/money-gopher/internal/persistence"
)

const builtinStateCookie = "mg_builtin_state"

// BuiltinConfig configures authentication against the embedded oauth2go
// authorization server (see cmd/moneyd's -auth=builtin mode).
//
// Unlike a real OIDC provider, oauth2go is a plain OAuth2 authorization
// server: it never issues an id_token, even though it exposes an OIDC
// discovery document. Its access token is, however, a signed JWT carrying
// the authenticated username in its "sub" claim, so identity is verified
// directly from that token's ES256 signature instead of going through
// OIDC id_token validation.
type BuiltinConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	// RedirectURL is the full URL of /auth/callback.
	RedirectURL string
	// PublicKeys returns the authorization server's current signing keys,
	// indexed by kid. Typically bound directly to (*oauth2go.AuthorizationServer).PublicKeys.
	PublicKeys func() map[int]*ecdsa.PublicKey
}

// BuiltinHandler implements the authorization-code flow against the
// embedded oauth2go authorization server.
type BuiltinHandler struct {
	db     *persistence.DB
	cfg    BuiltinConfig
	oauth2 oauth2.Config
}

// NewBuiltinHandler returns a BuiltinHandler for the given configuration.
func NewBuiltinHandler(db *persistence.DB, cfg BuiltinConfig) *BuiltinHandler {
	return &BuiltinHandler{
		db:  db,
		cfg: cfg,
		oauth2: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.Issuer + "/authorize",
				TokenURL: cfg.Issuer + "/token",
			},
		},
	}
}

// LoginHandler starts the authorization-code flow: stores a CSRF state in a
// short-lived cookie and redirects to the embedded login page.
func (h *BuiltinHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state := randomString(32)
	setTempCookie(w, builtinStateCookie, state)
	http.Redirect(w, r, h.oauth2.AuthCodeURL(state), http.StatusFound)
}

// CallbackHandler handles the redirect back from the embedded authorization
// server: verifies state, exchanges the code, verifies the access token's
// signature, provisions the user and their default person, and creates a
// session.
func (h *BuiltinHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie(builtinStateCookie)
	if err != nil || stateCookie.Value != r.FormValue("state") {
		http.Error(w, "invalid state parameter", http.StatusBadRequest)
		return
	}
	clearTempCookie(w, builtinStateCookie)

	token, err := h.oauth2.Exchange(r.Context(), r.FormValue("code"))
	if err != nil {
		http.Error(w, "code exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	subject, err := h.verifyAccessToken(token.AccessToken)
	if err != nil {
		http.Error(w, "access token verification failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	user, err := provisionUser(r.Context(), h.db, h.cfg.Issuer, subject, subject)
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
func (h *BuiltinHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ClearSession(r.Context(), h.db, w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// verifyAccessToken checks the JWT's ES256 signature against the
// authorization server's current public keys and returns its subject claim.
func (h *BuiltinHandler) verifyAccessToken(raw string) (string, error) {
	var claims jwt.MapClaims

	_, err := jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		kidStr, _ := t.Header["kid"].(string)
		var kid int
		if _, err := fmt.Sscanf(kidStr, "%d", &kid); err != nil {
			return nil, fmt.Errorf("invalid kid header %q: %w", kidStr, err)
		}

		key, ok := h.cfg.PublicKeys()[kid]
		if !ok {
			return nil, fmt.Errorf("unknown signing key %d", kid)
		}
		return key, nil
	})
	if err != nil {
		return "", err
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("access token missing sub claim")
	}
	return sub, nil
}
