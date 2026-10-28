package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	CreatedAt  time.Time
}

type IdempotencyManager struct {
	mu      sync.RWMutex
	records map[string]*CachedResponse
	ttl     time.Duration
}

func NewIdempotencyManager(ttl time.Duration) *IdempotencyManager {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	im := &IdempotencyManager{
		records: make(map[string]*CachedResponse),
		ttl:     ttl,
	}

	// Periodic cleanup of expired idempotency keys
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		for range ticker.C {
			im.cleanup()
		}
	}()

	return im
}

func (im *IdempotencyManager) cleanup() {
	im.mu.Lock()
	defer im.mu.Unlock()

	cutoff := time.Now().Add(-im.ttl)
	for key, rec := range im.records {
		if rec.CreatedAt.Before(cutoff) {
			delete(im.records, key)
		}
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Middleware enforces idempotency on mutating operations (POST, PUT, PATCH, DELETE) (TASK-085)
func (im *IdempotencyManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass idempotency cache for webhook endpoints (they have their own idempotency logic)
		if strings.HasPrefix(r.URL.Path, "/v1/payments/webhooks") {
			next.ServeHTTP(w, r)
			return
		}

		// Only check mutating requests with an Idempotency-Key header
		idempotencyKey := r.Header.Get("X-Idempotency-Key")
		if idempotencyKey == "" || (r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodDelete) {
			next.ServeHTTP(w, r)
			return
		}

		im.mu.RLock()
		cached, exists := im.records[idempotencyKey]
		im.mu.RUnlock()

		if exists && time.Since(cached.CreatedAt) < im.ttl {
			// Replay cached response exactly
			for k, values := range cached.Header {
				for _, v := range values {
					w.Header().Add(k, v)
				}
			}
			w.Header().Set("X-Cache-Lookup", "HIT-IDEMPOTENT")
			w.WriteHeader(cached.StatusCode)
			w.Write(cached.Body)
			return
		}

		// Record new response
		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		// Cache successful or business logic response
		im.mu.Lock()
		hdrCopy := make(http.Header)
		for k, v := range rec.Header() {
			hdrCopy[k] = v
		}
		im.records[idempotencyKey] = &CachedResponse{
			StatusCode: rec.statusCode,
			Header:     hdrCopy,
			Body:       rec.body.Bytes(),
			CreatedAt:  time.Now(),
		}
		im.mu.Unlock()
	})
}

// VerifyWebhookHMAC validates an incoming webhook's HMAC-SHA256 signature (TASK-084)
func VerifyWebhookHMAC(payload []byte, secret, signatureHeader string) bool {
	if secret == "" || signatureHeader == "" {
		return false
	}
	parts := strings.Split(signatureHeader, "=")
	var hexSig string
	if len(parts) == 2 && parts[0] == "sha256" {
		hexSig = parts[1]
	} else {
		hexSig = signatureHeader
	}

	expectedMAC := hmac.New(sha256.New, []byte(secret))
	expectedMAC.Write(payload)
	expectedBytes := expectedMAC.Sum(nil)
	expectedHex := hex.EncodeToString(expectedBytes)

	return subtle.ConstantTimeCompare([]byte(hexSig), []byte(expectedHex)) == 1
}
