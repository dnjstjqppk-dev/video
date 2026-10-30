package security

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CachedResponse struct {
	StatusCode  int
	Header      http.Header
	Body        []byte
	CreatedAt   time.Time
	Fingerprint [32]byte
}

type IdempotencyManager struct {
	mu      sync.RWMutex
	records map[string]*CachedResponse
	ttl     time.Duration
	pending map[string]chan struct{}
}

func NewIdempotencyManager(ttl time.Duration) *IdempotencyManager {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	im := &IdempotencyManager{
		records: make(map[string]*CachedResponse),
		ttl:     ttl,
		pending: make(map[string]chan struct{}),
	}

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
	statusCode  int
	body        bytes.Buffer
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
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

		// Anonymous operations need their own domain-specific deduplication.
		if r.Header.Get("Authorization") == "" {
			next.ServeHTTP(w, r)
			return
		}
		if len(idempotencyKey) > 256 {
			http.Error(w, "Idempotency key too long", http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			http.Error(w, "Invalid or oversized request body", http.StatusRequestEntityTooLarge)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(body))
		fingerprint := sha256.Sum256(append([]byte(r.Header.Get("Content-Type")+"\x00"), body...))
		scope := sha256.Sum256([]byte(r.Header.Get("Authorization") + "\x00" + r.Method + "\x00" + r.URL.RequestURI() + "\x00" + idempotencyKey))
		cacheKey := hex.EncodeToString(scope[:])
		im.cleanup()
		for {
			im.mu.Lock()
			if cached, exists := im.records[cacheKey]; exists && time.Since(cached.CreatedAt) < im.ttl {
				im.mu.Unlock()
				if cached.Fingerprint != fingerprint {
					http.Error(w, "Idempotency key reused with different request", http.StatusConflict)
					return
				}
				for k, values := range cached.Header {
					// Keep current request's security headers and request identifier.
					if k == "Content-Type" || k == "Location" {
						w.Header()[k] = append([]string(nil), values...)
					}
				}
				w.Header().Set("X-Cache-Lookup", "HIT-IDEMPOTENT")
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
				return
			}
			if done, exists := im.pending[cacheKey]; exists {
				im.mu.Unlock()
				select {
				case <-done:
					continue
				case <-r.Context().Done():
					return
				}
			}
			if len(im.records)+len(im.pending) >= 10000 {
				im.mu.Unlock()
				http.Error(w, "Idempotency cache capacity reached", http.StatusServiceUnavailable)
				return
			}
			im.pending[cacheKey] = make(chan struct{})
			im.mu.Unlock()
			break
		}
		defer func() {
			im.mu.Lock()
			close(im.pending[cacheKey])
			delete(im.pending, cacheKey)
			im.mu.Unlock()
		}()

		// Record new response
		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		// Failed operations must remain retryable.
		if rec.statusCode < 200 || rec.statusCode >= 300 {
			return
		}
		im.mu.Lock()
		im.records[cacheKey] = &CachedResponse{
			StatusCode:  rec.statusCode,
			Header:      rec.Header().Clone(),
			Body:        bytes.Clone(rec.body.Bytes()),
			CreatedAt:   time.Now(),
			Fingerprint: fingerprint,
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
