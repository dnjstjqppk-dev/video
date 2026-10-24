package security

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Claims struct {
	Subject       string          `json:"sub"`
	Issuer        string          `json:"iss"`
	Audience      json.RawMessage `json:"aud"`
	Exp           int64           `json:"exp"`
	Nbf           int64           `json:"nbf"`
	Iat           int64           `json:"iat"`
	Email         string          `json:"email"`
	EmailVerified bool            `json:"email_verified"`
}
type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}
type jwkSet struct {
	Keys []jwk `json:"keys"`
}

var ErrJWKSUnavailable = errors.New("jwks unavailable")

type Validator struct {
	Issuer, Audience, JWKSURL string
	Client                    *http.Client
	mu                        sync.RWMutex
	fetchMu                   sync.Mutex
	lastFetch                 time.Time
	keys                      map[string]*rsa.PublicKey
	cachedUntil               time.Time
}

func NewValidator(issuer, audience, jwksURL string) *Validator {
	return &Validator{Issuer: issuer, Audience: audience, JWKSURL: jwksURL, Client: &http.Client{Timeout: 5 * time.Second}, keys: map[string]*rsa.PublicKey{}}
}
func (v *Validator) refresh(ctx context.Context, kid string) error {
	v.fetchMu.Lock()
	defer v.fetchMu.Unlock()
	v.mu.RLock()
	_, present := v.keys[kid]
	count := len(v.keys)
	expired := time.Now().After(v.cachedUntil)
	last := v.lastFetch
	v.mu.RUnlock()
	if present && !expired {
		return nil
	}
	if count > 0 && !present && time.Since(last) < 30*time.Second {
		return nil
	}
	if count == 0 && !last.IsZero() && time.Since(last) < 5*time.Second {
		return ErrJWKSUnavailable
	}
	v.mu.Lock()
	v.lastFetch = time.Now()
	v.mu.Unlock()
	req, err := http.NewRequestWithContext(ctx, "GET", v.JWKSURL, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWKSUnavailable, err)
	}
	resp, err := v.Client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrJWKSUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: http %d", ErrJWKSUnavailable, resp.StatusCode)
	}
	var set jwkSet
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&set); err != nil {
		return fmt.Errorf("%w: %v", ErrJWKSUnavailable, err)
	}
	keys := map[string]*rsa.PublicKey{}
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Kid == "" || (k.Alg != "" && k.Alg != "RS256") || (k.Use != "" && k.Use != "sig") {
			continue
		}
		n, e1 := base64.RawURLEncoding.DecodeString(k.N)
		eb, e2 := base64.RawURLEncoding.DecodeString(k.E)
		if e1 != nil || e2 != nil || len(eb) == 0 || len(eb) > 4 {
			continue
		}
		exponent := 0
		for _, x := range eb {
			exponent = exponent<<8 | int(x)
		}
		if exponent < 3 || exponent%2 == 0 || len(n) < 256 {
			continue
		}
		keys[k.Kid] = &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: exponent}
	}
	if len(keys) == 0 {
		return fmt.Errorf("%w: empty or unusable keys", ErrJWKSUnavailable)
	}
	v.mu.Lock()
	v.keys = keys
	v.cachedUntil = time.Now().Add(5 * time.Minute)
	v.mu.Unlock()
	return nil
}
func (v *Validator) key(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key := v.keys[kid]
	expired := time.Now().After(v.cachedUntil)
	v.mu.RUnlock()
	if key != nil && !expired {
		return key, nil
	}
	if err := v.refresh(ctx, kid); err != nil {
		return nil, err
	}
	v.mu.RLock()
	defer v.mu.RUnlock()
	key = v.keys[kid]
	if key == nil {
		return nil, errors.New("unknown token key")
	}
	return key, nil
}
func (v *Validator) Validate(ctx context.Context, token string) (Claims, error) {
	var c Claims
	if len(token) > 8192 {
		return c, errors.New("token too large")
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c, errors.New("malformed JWT")
	}
	hdrRaw, e := base64.RawURLEncoding.DecodeString(parts[0])
	if e != nil {
		return c, e
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}
	if len(hdrRaw) > 1024 || json.Unmarshal(hdrRaw, &header) != nil || header.Alg != "RS256" || header.Kid == "" {
		return c, errors.New("unsupported jwt header")
	}
	key, e := v.key(ctx, header.Kid)
	if e != nil {
		return c, e
	}
	sig, e := base64.RawURLEncoding.DecodeString(parts[2])
	if e != nil {
		return c, e
	}
	hash := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if e = rsa.VerifyPKCS1v15(key, crypto.SHA256, hash[:], sig); e != nil {
		return c, errors.New("invalid signature")
	}
	payload, e := base64.RawURLEncoding.DecodeString(parts[1])
	if e != nil || len(payload) > 8192 || json.Unmarshal(payload, &c) != nil {
		return c, errors.New("invalid payload")
	}
	now := time.Now().Unix()
	if c.Issuer != v.Issuer || c.Subject == "" || c.Exp <= now || c.Exp > now+86400*30 || c.Nbf > now+60 || c.Iat > now+60 || !audienceOK(c.Audience, v.Audience) {
		return c, errors.New("invalid JWT claims")
	}
	return c, nil
}
func audienceOK(raw json.RawMessage, want string) bool {
	var one string
	if json.Unmarshal(raw, &one) == nil {
		return subtle.ConstantTimeCompare([]byte(one), []byte(want)) == 1
	}
	var many []string
	if json.Unmarshal(raw, &many) != nil {
		return false
	}
	for _, s := range many {
		if subtle.ConstantTimeCompare([]byte(s), []byte(want)) == 1 {
			return true
		}
	}
	return false
}
