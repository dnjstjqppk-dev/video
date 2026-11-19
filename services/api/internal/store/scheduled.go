package store

import (
	"time"
)

// PublishDueVideos flips scheduled videos whose time has come to public.
// It returns the flipped video IDs. Callers persist via PersistVideo.
func (s *MemoryStore) PublishDueVideos(now time.Time) []string {
	var flipped []string
	for _, v := range s.Videos {
		if v == nil || v.Visibility != "scheduled" || v.Status == "DELETED" {
			continue
		}
		if v.ScheduledAt == nil || now.Before(*v.ScheduledAt) {
			continue
		}
		v.Visibility = "public"
		v.PublishedAt = now
		flipped = append(flipped, v.ID)
		s.PersistVideo(v)
	}
	return flipped
}
