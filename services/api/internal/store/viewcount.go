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

// HourPoint is one hour bucket of counted channel views.
type HourPoint struct {
	Hour  time.Time
	Views int64
}

// RecordHourlyView adds delta counted views to a channel's hour bucket.
func (s *MemoryStore) RecordHourlyView(channelID string, at time.Time, delta int64) {
	if channelID == "" || delta == 0 {
		return
	}
	hour := at.UTC().Truncate(time.Hour)
	s.hourlyMu.Lock()
	if s.HourlyViews == nil {
		s.HourlyViews = make(map[string]map[int64]int64)
	}
	buckets, ok := s.HourlyViews[channelID]
	if !ok {
		buckets = make(map[int64]int64)
		s.HourlyViews[channelID] = buckets
	}
	buckets[hour.Unix()] += delta
	s.hourlyMu.Unlock()
}

// HourlySeries returns the trailing N hour buckets for a channel, oldest
// first, zero-filling hours with no counted views.
func (s *MemoryStore) HourlySeries(channelID string, hours int) []HourPoint {
	if hours <= 0 {
		return nil
	}
	end := time.Now().UTC().Truncate(time.Hour)
	s.hourlyMu.Lock()
	buckets := s.HourlyViews[channelID]
	s.hourlyMu.Unlock()
	out := make([]HourPoint, 0, hours)
	for i := hours - 1; i >= 0; i-- {
		h := end.Add(time.Duration(-i) * time.Hour)
		var views int64
		if buckets != nil {
			views = buckets[h.Unix()]
		}
		out = append(out, HourPoint{Hour: h, Views: views})
	}
	return out
}

// NoteCountedView records one deduped view against the owning channel's
// current hour bucket, persisting best-effort to PostgreSQL.
func (s *MemoryStore) NoteCountedView(v *Video) {
	if v == nil || v.ChannelID == "" {
		return
	}
	now := time.Now().UTC().Truncate(time.Hour)
	s.RecordHourlyView(v.ChannelID, now, 1)
	s.PersistHourlyView(v.ChannelID, now, 1)
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
