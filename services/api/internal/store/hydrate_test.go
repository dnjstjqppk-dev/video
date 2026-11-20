package store

import "testing"

// Persisted video rows are normalized (no channel columns); hydration must
// re-attach denormalized channel identity or feeds/search/playback lose
// channel names and avatars for every hydrated video.
func TestEnrichVideoChannelsRestoresIdentity(t *testing.T) {
	s := NewMemoryStore()
	const videoID = "00000000-0000-0000-0000-0000000000a1"
	v, ok := s.Videos[videoID]
	if !ok {
		t.Fatalf("seed video %s missing", videoID)
	}
	// Simulate a sparse PostgreSQL row overlaid by Hydrate.
	v.ChannelName = ""
	v.ChannelHandle = ""
	v.ChannelAvatar = ""
	v.ChannelSubscriberCount = 0
	v.ChannelVideoCount = 0
	v.ChannelIsVerified = false

	s.EnrichVideoChannels()

	ch := s.Channels[v.ChannelID]
	if ch == nil {
		t.Fatalf("seed channel %s missing", v.ChannelID)
	}
	if v.ChannelName != ch.Name || v.ChannelHandle != ch.Handle || v.ChannelAvatar != ch.AvatarURL {
		t.Fatalf("channel identity not restored: %+v", v)
	}
	if v.ChannelSubscriberCount != ch.SubscriberCount || v.ChannelVideoCount != ch.VideoCount || v.ChannelIsVerified != ch.IsVerified {
		t.Fatalf("channel stats not restored: %+v", v)
	}
}

func TestEnrichVideoChannelsSkipsOrphans(t *testing.T) {
	s := NewMemoryStore()
	orphan := &Video{ID: "orphan-video", ChannelID: "no-such-channel", Title: "orphan"}
	s.Videos[orphan.ID] = orphan
	s.EnrichVideoChannels()
	if orphan.ChannelName != "" || orphan.ChannelAvatar != "" {
		t.Fatalf("orphan video must be left untouched: %+v", orphan)
	}
}
