package main

import (
	"testing"
	"time"
)

func TestRecommendationPipelineFilteringRankingAndColdStart(t *testing.T) {
	now := time.Now()
	user := UserProfile{
		UserID:            "u-100",
		PreferredCategory: "Technology",
		AgeVerified:       false,
		BlockedChannels:   map[string]bool{"ch-blocked": true},
	}

	candidates := []VideoCandidate{
		{ID: "v-private", ChannelID: "ch-1", Category: "Technology", Status: "READY", Visibility: "private", PublishedAt: now},
		{ID: "v-age", ChannelID: "ch-1", Category: "Technology", Status: "READY", Visibility: "public", AgeRestricted: true, PublishedAt: now},
		{ID: "v-blocked", ChannelID: "ch-blocked", Category: "Technology", Status: "READY", Visibility: "public", PublishedAt: now},
		{ID: "v-old-other", ChannelID: "ch-2", Category: "Cooking", Status: "READY", Visibility: "public", Impressions: 50000, AvgWatchDurationSec: 120, PublishedAt: now.Add(-120 * time.Hour)},
		{ID: "v-coldstart-tech", ChannelID: "ch-3", Category: "Technology", Status: "READY", Visibility: "public", Impressions: 35, AvgWatchDurationSec: 200, PublishedAt: now.Add(-2 * time.Hour)},
	}

	ranked := RankCandidates(candidates, user, now)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 eligible videos after safety filtering, got %d", len(ranked))
	}
	if ranked[0].ID != "v-coldstart-tech" {
		t.Fatalf("expected cold-start relevant video ranked #1, got %s", ranked[0].ID)
	}
}
