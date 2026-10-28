package security

import (
	"crypto/subtle"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type AdminGuard struct {
	mu           sync.RWMutex
	allowedCIDRs []*net.IPNet
	sudoSessions map[string]time.Time // key: userID -> sudo active until
}

func NewAdminGuard() *AdminGuard {
	ag := &AdminGuard{
		sudoSessions: make(map[string]time.Time),
	}

	rawCIDRs := os.Getenv("GVP_ADMIN_WHITELIST_CIDR")
	if rawCIDRs == "" {
		// Default secure loopback + local networks in dev
		rawCIDRs = "127.0.0.1/32,::1/128,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"
	}

	for _, cidr := range strings.Split(rawCIDRs, ",") {
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil && ipNet != nil {
			ag.allowedCIDRs = append(ag.allowedCIDRs, ipNet)
		}
	}

	return ag
}

// IPWhitelistMiddleware verifies incoming requests come from authorized network CIDRs (TASK-011)
func (ag *AdminGuard) IPWhitelistMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In non-production environments, bypass IP whitelist for testing/development convenience
		env := strings.ToLower(os.Getenv("GVP_ENV"))
		if env != "production" && env != "prod" {
			next.ServeHTTP(w, r)
			return
		}

		clientIPStr := getClientIP(r)
		parsedIP := net.ParseIP(clientIPStr)

		if parsedIP == nil {
			http.Error(w, `{"error":{"code":"FORBIDDEN_IP","message":"Unable to determine client IP"}}`, http.StatusForbidden)
			return
		}

		ag.mu.RLock()
		allowed := false
		for _, cidr := range ag.allowedCIDRs {
			if cidr.Contains(parsedIP) {
				allowed = true
				break
			}
		}
		ag.mu.RUnlock()

		if !allowed {
			http.Error(w, `{"error":{"code":"FORBIDDEN_IP","message":"Access denied: IP not on administrative whitelist"}}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ElevateSudo activates a 15-minute Sudo Mode for an administrator (TASK-012)
func (ag *AdminGuard) ElevateSudo(userID string) {
	ag.mu.Lock()
	defer ag.mu.Unlock()
	ag.sudoSessions[userID] = time.Now().Add(15 * time.Minute)
}

// IsSudoActive checks if user has active sudo privileges (TASK-012)
func (ag *AdminGuard) IsSudoActive(userID string) bool {
	ag.mu.RLock()
	defer ag.mu.RUnlock()

	validUntil, exists := ag.sudoSessions[userID]
	if !exists {
		return false
	}
	return time.Now().Before(validUntil)
}

// RequireSudo verifies sudo mode or valid sudo token for high-risk operations (TASK-012)
func (ag *AdminGuard) RequireSudo(getUserID func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := getUserID(r)
			if userID == "" {
				http.Error(w, `{"error":{"code":"UNAUTHENTICATED","message":"User identification required for sudo check"}}`, http.StatusUnauthorized)
				return
			}

			// Check 1: Active 15-minute sudo session
			if ag.IsSudoActive(userID) {
				next.ServeHTTP(w, r)
				return
			}

			// Check 2: 1-time Sudo Token Header
			sudoToken := r.Header.Get("X-Sudo-Token")
			if sudoToken != "" {
				expectedSecret := os.Getenv("ADMIN_SUDO_SECRET")
				if expectedSecret == "" {
					expectedSecret = "default-dev-sudo-secret-32-chars-long"
				}
				if subtle.ConstantTimeCompare([]byte(sudoToken), []byte(expectedSecret)) == 1 {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Sudo verification required
			http.Error(w, `{"error":{"code":"SUDO_REQUIRED","message":"Administrative Sudo re-authentication required within last 15 minutes"}}`, http.StatusForbidden)
		})
	}
}
