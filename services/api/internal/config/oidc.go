package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultDevJWTSecret = "super-secret-jwt-key-for-development-32chars!"

// OIDC holds the external OpenID Connect issuer settings. Any
// standards-compliant provider works (ZITADEL, Auth0, Keycloak); the API
// validates RS256 JWTs against the provider JWKS endpoint.
type OIDC struct {
	Issuer   string
	Audience string
	JWKSURL  string
}

// LoadOIDC reads OIDC settings from the environment.
func LoadOIDC() OIDC {
	return OIDC{
		Issuer:   strings.TrimSpace(os.Getenv("OIDC_ISSUER")),
		Audience: strings.TrimSpace(os.Getenv("OIDC_AUDIENCE")),
		JWKSURL:  strings.TrimSpace(os.Getenv("OIDC_JWKS_URL")),
	}
}

// IsProduction reports whether the API runs in production mode.
func IsProduction() bool {
	return strings.ToLower(os.Getenv("GVP_ENV")) == "production"
}

// Validate fails fast on configuration that cannot serve production traffic.
// Development and test environments only warn, preserving the existing
// fail-open local behavior (dev tokens stay gated behind GVP_ALLOW_DEV_TOKENS).
func (c *Config) Validate() error {
	if !IsProduction() {
		return nil
	}
	oidc := LoadOIDC()
	var missing []string
	if oidc.Issuer == "" {
		missing = append(missing, "OIDC_ISSUER (e.g. https://<tenant>.zitadel.cloud)")
	}
	if oidc.Audience == "" {
		missing = append(missing, "OIDC_AUDIENCE (API project resource ID)")
	}
	if oidc.JWKSURL == "" {
		missing = append(missing, "OIDC_JWKS_URL (issuer jwks.json URL)")
	}
	if c.JWTSecret == "" || c.JWTSecret == defaultDevJWTSecret {
		missing = append(missing, "JWT_SECRET (unique 32+ char secret, dev default forbidden)")
	}
	if len(missing) > 0 {
		return fmt.Errorf("production config incomplete, refusing to boot:\n  - %s\nprovision the values per docs/operations/OIDC_SETUP.md", strings.Join(missing, "\n  - "))
	}
	if strings.Contains(oidc.Issuer, "example.com") {
		return fmt.Errorf("OIDC_ISSUER still points at a placeholder (%s); provision a real issuer per docs/operations/OIDC_SETUP.md", oidc.Issuer)
	}
	return nil
}

// discoveryDocument is the minimal subset of RFC 8414 / OIDC Discovery used
// to prove an issuer is reachable and matches the configured value.
type discoveryDocument struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

// ProbeOIDCDiscovery fetches the issuer discovery document and verifies the
// issuer claim matches. It is a provisioning check, not a per-request cost:
// call it once at startup.
func ProbeOIDCDiscovery(ctx context.Context, issuer string) error {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	if issuer == "" {
		return fmt.Errorf("empty issuer")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issuer+"/.well-known/openid-configuration", nil)
	if err != nil {
		return fmt.Errorf("build discovery request: %w", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("issuer %s unreachable: %w", issuer, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("issuer %s returned discovery status %d", issuer, resp.StatusCode)
	}
	var doc discoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return fmt.Errorf("decode discovery document: %w", err)
	}
	if !strings.EqualFold(strings.TrimRight(doc.Issuer, "/"), issuer) {
		return fmt.Errorf("discovery issuer mismatch: got %q, want %q", doc.Issuer, issuer)
	}
	return nil
}
