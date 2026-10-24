package main

import (
	"testing"
	"time"
)

func TestAnalyticsDeduplicationBotFilterAndClickHouseBatchSink(t *testing.T) {
	proc := NewAnalyticsProcessor(2)

	ev1 := TelemetryEvent{
		EventID:      "evt-1",
		EventType:    "valid_view",
		VideoID:      "v1",
		UserID:       "u1",
		UserAgent:    "Mozilla/5.0 Chrome/128.0",
		WatchSeconds: 45,
		PlaybackRate: 1.0,
		OccurredAt:   time.Now(),
	}

	// 1. First ingestion accepted
	ok, err := proc.Ingest(ev1)
	if err != nil || !ok {
		t.Fatalf("expected ev1 accepted, got ok=%v err=%v", ok, err)
	}

	// 2. Duplicate event_id deduplicated
	ok, err = proc.Ingest(ev1)
	if err != nil || ok {
		t.Fatalf("expected duplicate ev1 rejected by deduplicator")
	}

	// 3. Headless bot traffic filtered out
	botEv := TelemetryEvent{
		EventID:      "evt-bot",
		EventType:    "ad_click",
		VideoID:      "v1",
		UserAgent:    "Mozilla/5.0 HeadlessChrome/120.0",
		PlaybackRate: 1.0,
	}
	ok, err = proc.Ingest(botEv)
	if err != nil || ok {
		t.Fatalf("expected bot traffic filtered out")
	}

	// 4. Second valid event reaches batch size (2) and triggers automatic ClickHouse batch sink flush
	ev2 := TelemetryEvent{
		EventID:      "evt-2",
		EventType:    "ad_impression",
		VideoID:      "v1",
		UserID:       "u1",
		UserAgent:    "Mozilla/5.0 Safari/605.1",
		WatchSeconds: 15,
		PlaybackRate: 1.25,
	}
	ok, err = proc.Ingest(ev2)
	if err != nil || !ok {
		t.Fatalf("expected ev2 accepted")
	}

	if len(proc.FlushedRows) != 2 || len(proc.BatchBuffer) != 0 {
		t.Fatalf("expected 2 events flushed to ClickHouse sink, got flushed=%d buffered=%d", len(proc.FlushedRows), len(proc.BatchBuffer))
	}
}
