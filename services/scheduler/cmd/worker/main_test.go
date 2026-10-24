package main

import (
	"testing"
	"time"
)

func TestSchedulerOutboxPublishCleanupAndSettlementClosing(t *testing.T) {
	now := time.Now()
	eng := NewSchedulerEngine()

	// 1. Outbox relay
	eng.Outbox = append(eng.Outbox,
		&OutboxEvent{ID: "ob-1", Topic: "media.jobs.transcode", PayloadJSON: `{"videoId":"v1"}`},
		&OutboxEvent{ID: "ob-2", Topic: "payments.settled", PayloadJSON: `{"txId":"tx1"}`},
	)
	if n := eng.RelayOutboxEvents(); n != 2 {
		t.Fatalf("expected 2 outbox events relayed, got %d", n)
	}
	if n := eng.RelayOutboxEvents(); n != 0 {
		t.Fatalf("expected 0 outbox events on second poll, got %d", n)
	}

	// 2. Scheduled video publishing
	eng.Videos = append(eng.Videos,
		&ScheduledVideo{ID: "v-due", Visibility: "scheduled", PublishAt: now.Add(-5 * time.Minute)},
		&ScheduledVideo{ID: "v-future", Visibility: "scheduled", PublishAt: now.Add(2 * time.Hour)},
	)
	if n := eng.PublishDueVideos(now); n != 1 || eng.Videos[0].Visibility != "public" {
		t.Fatalf("expected 1 due video published to public, got %d", n)
	}

	// 3. Expired upload session cleanup
	eng.Uploads = append(eng.Uploads,
		&UploadSession{UploadID: "up-stale", Status: "pending", ExpiresAt: now.Add(-1 * time.Hour)},
		&UploadSession{UploadID: "up-active", Status: "pending", ExpiresAt: now.Add(1 * time.Hour)},
	)
	if n := eng.CleanupExpiredUploads(now); n != 1 || eng.Uploads[0].Status != "expired" {
		t.Fatalf("expected 1 stale upload cleaned up, got %d", n)
	}

	// 4. Monthly settlement period closing
	eng.Settlements["2026-10"] = &SettlementPeriod{PeriodID: "2026-10", Status: "open"}
	if !eng.CloseSettlementPeriod("2026-10") || eng.Settlements["2026-10"].Status != "closed" {
		t.Fatal("expected settlement period 2026-10 to transition to closed")
	}
}
