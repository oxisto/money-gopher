package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/golang-jwt/jwt/v5"
)

// testProvider spins up a fake OIDC provider serving a discovery document
// and a JWKS, and can mint id_tokens signed with its key.
type testProvider struct {
	server *httptest.Server
	key    *ecdsa.PrivateKey
	kid    string
}

func newTestProvider(t *testing.T) *testProvider {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	tp := &testProvider{key: key, kid: "test-kid"}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		disc := oidcDiscovery{
			Issuer:                tp.server.URL,
			AuthorizationEndpoint: tp.server.URL + "/authorize",
			TokenEndpoint:         tp.server.URL + "/token",
			JWKSURI:               tp.server.URL + "/jwks",
		}
		_ = json.NewEncoder(w).Encode(disc)
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		jwk, err := jwkset.NewJWKFromKey(&key.PublicKey, jwkset.JWKOptions{
			Metadata: jwkset.JWKMetadataOptions{KID: tp.kid},
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		store := jwkset.NewMemoryStorage()
		if err := store.KeyWrite(r.Context(), jwk); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		raw, err := store.JSONPublic(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(raw)
	})

	tp.server = httptest.NewServer(mux)
	t.Cleanup(tp.server.Close)
	return tp
}

func (tp *testProvider) signIDToken(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = tp.kid
	signed, err := token.SignedString(tp.key)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

func TestNewOIDCHandler_Discovery(t *testing.T) {
	tp := newTestProvider(t)
	ctx := context.Background()

	h, err := NewOIDCHandler(ctx, nil, OIDCConfig{
		Issuer:       tp.server.URL,
		ClientID:     "moneyd",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
	})
	if err != nil {
		t.Fatalf("NewOIDCHandler() error = %v", err)
	}
	if h.oauth2.Endpoint.AuthURL != tp.server.URL+"/authorize" {
		t.Errorf("AuthURL = %q, want %q", h.oauth2.Endpoint.AuthURL, tp.server.URL+"/authorize")
	}
	if h.oauth2.Endpoint.TokenURL != tp.server.URL+"/token" {
		t.Errorf("TokenURL = %q, want %q", h.oauth2.Endpoint.TokenURL, tp.server.URL+"/token")
	}
}

func TestOIDCHandler_VerifyIDToken(t *testing.T) {
	tp := newTestProvider(t)
	ctx := context.Background()

	h, err := NewOIDCHandler(ctx, nil, OIDCConfig{
		Issuer:       tp.server.URL,
		ClientID:     "moneyd",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
	})
	if err != nil {
		t.Fatalf("NewOIDCHandler() error = %v", err)
	}

	baseClaims := func() jwt.MapClaims {
		return jwt.MapClaims{
			"iss":   tp.server.URL,
			"aud":   "moneyd",
			"sub":   "user-123",
			"nonce": "the-nonce",
			"name":  "Test User",
			"exp":   time.Now().Add(time.Hour).Unix(),
			"iat":   time.Now().Unix(),
		}
	}

	t.Run("valid token verifies", func(t *testing.T) {
		raw := tp.signIDToken(t, baseClaims())
		var claims jwt.MapClaims
		_, err := jwt.ParseWithClaims(raw, &claims, h.keyfunc(ctx),
			jwt.WithValidMethods(idTokenSigningAlgorithms),
			jwt.WithIssuer(h.issuer),
			jwt.WithAudience(h.oauth2.ClientID),
		)
		if err != nil {
			t.Fatalf("ParseWithClaims() error = %v", err)
		}
		if claims["sub"] != "user-123" {
			t.Errorf("sub = %v, want user-123", claims["sub"])
		}
	})

	t.Run("wrong audience rejected", func(t *testing.T) {
		claims := baseClaims()
		claims["aud"] = "someone-else"
		raw := tp.signIDToken(t, claims)
		_, err := jwt.ParseWithClaims(raw, &jwt.MapClaims{}, h.keyfunc(ctx),
			jwt.WithValidMethods(idTokenSigningAlgorithms),
			jwt.WithIssuer(h.issuer),
			jwt.WithAudience(h.oauth2.ClientID),
		)
		if err == nil {
			t.Fatal("expected error for wrong audience, got nil")
		}
	})

	t.Run("wrong issuer rejected", func(t *testing.T) {
		claims := baseClaims()
		claims["iss"] = "https://evil.example.com"
		raw := tp.signIDToken(t, claims)
		_, err := jwt.ParseWithClaims(raw, &jwt.MapClaims{}, h.keyfunc(ctx),
			jwt.WithValidMethods(idTokenSigningAlgorithms),
			jwt.WithIssuer(h.issuer),
			jwt.WithAudience(h.oauth2.ClientID),
		)
		if err == nil {
			t.Fatal("expected error for wrong issuer, got nil")
		}
	})

	t.Run("expired token rejected", func(t *testing.T) {
		claims := baseClaims()
		claims["exp"] = time.Now().Add(-time.Hour).Unix()
		raw := tp.signIDToken(t, claims)
		_, err := jwt.ParseWithClaims(raw, &jwt.MapClaims{}, h.keyfunc(ctx),
			jwt.WithValidMethods(idTokenSigningAlgorithms),
			jwt.WithIssuer(h.issuer),
			jwt.WithAudience(h.oauth2.ClientID),
		)
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})

	t.Run("HS256-signed token rejected despite valid-looking claims", func(t *testing.T) {
		// Simulates an algorithm-confusion attack: sign with HMAC using the
		// EC public key's bytes as the "secret", hoping a naive verifier
		// accepts it because it doesn't restrict signing methods.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, baseClaims())
		token.Header["kid"] = tp.kid
		raw, err := token.SignedString([]byte("attacker-controlled-or-guessed-key"))
		if err != nil {
			t.Fatalf("SignedString() error = %v", err)
		}
		_, err = jwt.ParseWithClaims(raw, &jwt.MapClaims{}, h.keyfunc(ctx),
			jwt.WithValidMethods(idTokenSigningAlgorithms),
			jwt.WithIssuer(h.issuer),
			jwt.WithAudience(h.oauth2.ClientID),
		)
		if err == nil {
			t.Fatal("expected error for HS256-signed token, got nil")
		}
	})

	t.Run("unknown kid rejected", func(t *testing.T) {
		claims := baseClaims()
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token.Header["kid"] = "no-such-key"
		raw, err := token.SignedString(tp.key)
		if err != nil {
			t.Fatalf("SignedString() error = %v", err)
		}
		_, err = jwt.ParseWithClaims(raw, &jwt.MapClaims{}, h.keyfunc(ctx),
			jwt.WithValidMethods(idTokenSigningAlgorithms),
			jwt.WithIssuer(h.issuer),
			jwt.WithAudience(h.oauth2.ClientID),
		)
		if err == nil {
			t.Fatal("expected error for unknown kid, got nil")
		}
	})
}

func TestNewOIDCHandler_IssuerMismatch(t *testing.T) {
	tp := newTestProvider(t)
	ctx := context.Background()

	_, err := NewOIDCHandler(ctx, nil, OIDCConfig{
		// Trailing slash makes this differ textually from the discovery
		// document's "issuer" field, which must be rejected per spec to
		// prevent issuer mix-up attacks.
		Issuer:       tp.server.URL + "/",
		ClientID:     "moneyd",
		ClientSecret: "secret",
		RedirectURL:  "http://localhost:8080/auth/callback",
	})
	if err == nil {
		t.Fatal("expected error for issuer mismatch, got nil")
	}
}
