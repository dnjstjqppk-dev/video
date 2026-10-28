package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
)

type nonceContextKey struct{}

var NonceKey = nonceContextKey{}

// SecurityHeadersMiddleware injects standard production hardening headers (TASK-024 - TASK-028)
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate 128-bit cryptographic nonce for CSP Level 3
		nonceBytes := make([]byte, 16)
		if _, err := rand.Read(nonceBytes); err != nil {
			http.Error(w, "internal security error", http.StatusInternalServerError)
			return
		}
		nonce := base64.RawStdEncoding.EncodeToString(nonceBytes)

		// TASK-026: Strict-Transport-Security (HSTS)
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")

		// TASK-027: X-Content-Type-Options & X-Frame-Options
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// TASK-028: Cross-Origin-Opener-Policy & Cross-Origin-Resource-Policy
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		// TASK-025: Content-Security-Policy (CSP Level 3 Nonce)
		csp := fmt.Sprintf("default-src 'self'; script-src 'self' 'nonce-%s' https:; style-src 'self' 'unsafe-inline' https:; img-src 'self' data: https: blob:; media-src 'self' https: blob:; connect-src 'self' https: wss:; object-src 'none'; base-uri 'self'; frame-ancestors 'self';", nonce)
		w.Header().Set("Content-Security-Policy", csp)

		// Additional modern browser hardening
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "accelerometer=(), camera=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Expose nonce via request context
		ctx := context.WithValue(r.Context(), NonceKey, nonce)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCSPNonce returns the CSP nonce from request context
func GetCSPNonce(ctx context.Context) string {
	if n, ok := ctx.Value(NonceKey).(string); ok {
		return n
	}
	return ""
}
