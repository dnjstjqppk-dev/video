package security

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientWindow struct {
	timestamps []time.Time
}

type RateLimiter struct {
	mu           sync.Mutex
	windows      map[string]*clientWindow
	limit        int
	windowPeriod time.Duration
}

func NewRateLimiter(limit int, windowPeriod time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 120
	}
	if windowPeriod <= 0 {
		windowPeriod = time.Minute
	}
	rl := &RateLimiter{
		windows:      make(map[string]*clientWindow),
		limit:        limit,
		windowPeriod: windowPeriod,
	}

	// Background eviction cleanup every 2 minutes
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-rl.windowPeriod)
	for key, win := range rl.windows {
		var valid []time.Time
		for _, ts := range win.timestamps {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		if len(valid) == 0 {
			delete(rl.windows, key)
		} else {
			win.timestamps = valid
		}
	}
}

func (rl *RateLimiter) Allow(key string) (bool, int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.windowPeriod)

	win, exists := rl.windows[key]
	if !exists {
		win = &clientWindow{}
		rl.windows[key] = win
	}

	// Filter out expired timestamps
	var valid []time.Time
	for _, ts := range win.timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	win.timestamps = valid

	remaining := rl.limit - len(win.timestamps)
	if remaining <= 0 {
		// Calculate retry-after duration based on oldest timestamp
		retryAfter := time.Second
		if len(win.timestamps) > 0 {
			retryAfter = win.timestamps[0].Add(rl.windowPeriod).Sub(now)
			if retryAfter < time.Second {
				retryAfter = time.Second
			}
		}
		return false, 0, retryAfter
	}

	win.timestamps = append(win.timestamps, now)
	return true, remaining - 1, 0
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if rip := r.Header.Get("X-Real-IP"); rip != "" {
		return strings.TrimSpace(rip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// Middleware creates an HTTP middleware applying the sliding window rate limit (TASK-015)
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		allowed, remaining, retryAfter := rl.Allow(ip)

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retryAfter.Seconds())))
			http.Error(w, `{"error":{"code":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please slow down."}}`, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
