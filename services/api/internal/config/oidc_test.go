package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestValidateProduction(t *testing.T) {
	t.Setenv("GVP_ENV", "production")
	t.Setenv("OIDC_ISSUER", "https://tenant-12345.zitadel.cloud")
	t.Setenv("OIDC_AUDIENCE", "api-project-id")
	t.Setenv("OIDC_JWKS_URL", "https://tenant-12345.zitadel.cloud/oauth/v2/keys")
	cfg := &Config{JWTSecret: "a-real-32-char-minimum-secret-value!"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid production config, got %v", err)
	}
}

func TestValidateProductionMissing(t *testing.T) {
	t.Setenv("GVP_ENV", "production")
	t.Setenv("OIDC_ISSUER", "")
	t.Setenv("OIDC_AUDIENCE", "")
	t.Setenv("OIDC_JWKS_URL", "")
	cfg := &Config{JWTSecret: defaultDevJWTSecret}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for incomplete production config")
	}
}

func TestValidateProductionRejectsPlaceholderIssuer(t *testing.T) {
	t.Setenv("GVP_ENV", "production")
	t.Setenv("OIDC_ISSUER", "https://auth.example.com")
	t.Setenv("OIDC_AUDIENCE", "api")
	t.Setenv("OIDC_JWKS_URL", "https://auth.example.com/keys")
	cfg := &Config{JWTSecret: "a-real-32-char-minimum-secret-value!"}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected placeholder issuer rejection")
	}
}

func TestValidateNonProductionIsLenient(t *testing.T) {
	t.Setenv("GVP_ENV", "development")
	t.Setenv("OIDC_ISSUER", "")
	cfg := &Config{JWTSecret: defaultDevJWTSecret}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected lenient dev validation, got %v", err)
	}
}

func TestProbeOIDCDiscovery(t *testing.T) {
	newServer := func(status int, body string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/.well-known/openid-configuration" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(status)
			_, _ = w.Write([]byte(body))
		}))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("accepts matching issuer", func(t *testing.T) {
		dyn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"issuer":"` + "http://" + r.Host + `","jwks_uri":"http://` + r.Host + `/keys"}`))
		}))
		defer dyn.Close()
		if err := ProbeOIDCDiscovery(ctx, dyn.URL); err != nil {
			t.Fatalf("expected probe success, got %v", err)
		}
	})

	t.Run("rejects issuer mismatch", func(t *testing.T) {
		srv := newServer(http.StatusOK, `{"issuer":"https://other.example.com","jwks_uri":"https://other.example.com/keys"}`)
		defer srv.Close()
		if err := ProbeOIDCDiscovery(ctx, srv.URL); err == nil {
			t.Fatal("expected issuer mismatch error")
		}
	})

	t.Run("rejects non-200", func(t *testing.T) {
		srv := newServer(http.StatusInternalServerError, `{}`)
		defer srv.Close()
		if err := ProbeOIDCDiscovery(ctx, srv.URL); err == nil {
			t.Fatal("expected status error")
		}
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		srv := newServer(http.StatusOK, `not-json`)
		defer srv.Close()
		if err := ProbeOIDCDiscovery(ctx, srv.URL); err == nil {
			t.Fatal("expected decode error")
		}
	})

	t.Run("rejects empty issuer", func(t *testing.T) {
		if err := ProbeOIDCDiscovery(ctx, "  "); err == nil {
			t.Fatal("expected empty issuer error")
		}
	})
}
