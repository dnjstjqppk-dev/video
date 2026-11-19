package store_test

import (
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/store"
)

func TestPublishDueVideos_FlipsOnlyDueScheduled(t *testing.T) {
	s := store.NewMemoryStore()
	future := time.Now().Add(2 * time.Hour)
	past := time.Now().Add(-2 * time.Hour)
	s.Videos["vid-due"] = &store.Video{ID: "vid-due", ChannelID: "ch1", Visibility: "scheduled", Status: "READY", ScheduledAt: &past}
	s.Videos["vid-future"] = &store.Video{ID: "vid-future", ChannelID: "ch1", Visibility: "scheduled", Status: "READY", ScheduledAt: &future}
	s.Videos["vid-unscheduled"] = &store.Video{ID: "vid-unscheduled", ChannelID: "ch1", Visibility: "scheduled", Status: "READY"}
	s.Videos["vid-public"] = &store.Video{ID: "vid-public", ChannelID: "ch1", Visibility: "public", Status: "READY", ScheduledAt: &past}

	flipped := s.PublishDueVideos(time.Now())
	if len(flipped) != 1 || flipped[0] != "vid-due" {
		t.Fatalf("expected only vid-due flipped, got %v", flipped)
	}
	if v := s.Videos["vid-due"]; v.Visibility != "public" || v.PublishedAt.IsZero() {
		t.Fatal("due video did not flip to public with publish time")
	}
	if v := s.Videos["vid-future"]; v.Visibility != "scheduled" {
		t.Fatal("future video must stay scheduled")
	}
	if v := s.Videos["vid-unscheduled"]; v.Visibility != "scheduled" {
		t.Fatal("scheduled video without time must stay scheduled")
	}
}
