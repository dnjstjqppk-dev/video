package security

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func encode(v any) string { b, _ := json.Marshal(v); return base64.RawURLEncoding.EncodeToString(b) }
func TestJWTSignatureAndAudience(t *testing.T) {
	key, e := rsa.GenerateKey(rand.Reader, 2048)
	if e != nil {
		t.Fatal(e)
	}
	n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	exp := big.NewInt(int64(key.PublicKey.E)).Bytes()
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]any{"kty": "RSA", "kid": "local", "alg": "RS256", "use": "sig", "n": n, "e": base64.RawURLEncoding.EncodeToString(exp)}}})
	}))
	defer jwks.Close()
	issuer := "https://oidc.example.test"
	v := NewValidator(issuer, "test-api", jwks.URL)
	signed := func(aud any) string {
		header := encode(map[string]any{"alg": "RS256", "kid": "local"})
		payload := encode(map[string]any{"iss": issuer, "aud": aud, "sub": "test-subject", "exp": time.Now().Add(3 * time.Minute).Unix()})
		data := header + "." + payload
		digest := sha256.Sum256([]byte(data))
		sig, _ := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
		return data + "." + base64.RawURLEncoding.EncodeToString(sig)
	}
	claims, err := v.Validate(context.Background(), signed([]string{"other", "test-api"}))
	if err != nil || claims.Subject != "test-subject" {
		t.Fatalf("expected valid token: %v", err)
	}
	if _, err := v.Validate(context.Background(), signed("different-audience")); err == nil {
		t.Fatal("aud mismatch accepted")
	}
	valid := signed("test-api")
	tampered := valid[:len(valid)-2] + "zz"
	if _, err := v.Validate(context.Background(), tampered); err == nil {
		t.Fatal("signature tamper accepted")
	}
}
