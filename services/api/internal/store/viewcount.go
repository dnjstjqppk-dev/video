package store

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

// ViewDedupWindow bounds view counting: one counted view per viewer per
// video per window (P1-14 valid-view definition). Overridable in tests.
var ViewDedupWindow = 24 * time.Hour

// ViewerKey derives a stable counting identity: the authenticated user ID
// when known, else a hash of the client address + user agent. The raw
// address is never stored, only the digest.
func ViewerKey(r *http.Request, userID string) string {
	if userID != "" {
		return "user:" + userID
	}
	addr := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first, _, _ := strings.Cut(forwarded, ","); strings.TrimSpace(first) != "" {
			addr = strings.TrimSpace(first)
		}
	}
	sum := sha256.Sum256([]byte(addr + "\n" + r.Header.Get("User-Agent")))
	return "anon:" + hex.EncodeToString(sum[:16])
}

// CountView records a view for (viewer, video), returning false when the
// same viewer already counted this video inside the dedup window. Both the
// playback-grant path and the analytics valid_view path consult this guard
// so one viewing session increments ViewCount exactly once.
func (s *MemoryStore) CountView(viewerKey, videoID string) bool {
	if viewerKey == "" || videoID == "" {
		return false
	}
	now := time.Now()
	key := viewerKey + "|" + videoID
	s.viewMu.Lock()
	defer s.viewMu.Unlock()
	if s.viewSeen == nil {
		s.viewSeen = make(map[string]time.Time)
	}
	if last, ok := s.viewSeen[key]; ok && now.Sub(last) < ViewDedupWindow {
		return false
	}
	s.viewSeen[key] = now
	// Opportunistic prune: drop expired entries on every ~256th insert.
	if len(s.viewSeen)%256 == 0 {
		for k, at := range s.viewSeen {
			if now.Sub(at) >= ViewDedupWindow {
				delete(s.viewSeen, k)
			}
		}
	}
	return true
}
