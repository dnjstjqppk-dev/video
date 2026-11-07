package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newIntakeMux() *http.ServeMux {
	srv := newHealthServer("analytics-worker", NewAnalyticsProcessor(4))
	mux := http.NewServeMux()
	srv.registerIntakeRoutes(mux)
	return mux
}

func doIntake(t *testing.T, mux *http.ServeMux, method, path, body string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var decoded map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &decoded)
	return rec.Code, decoded
}

func TestIntakeIngestAndFlush(t *testing.T) {
	mux := newIntakeMux()

	code, res := doIntake(t, mux, http.MethodPost, "/ingest",
		`{"eventId":"ev-1","eventType":"valid_view","videoId":"v-1","userId":"u-1","userAgent":"Mozilla/5.0","watchSeconds":120,"playbackRate":1,"occurredAt":"2026-11-07T10:00:00Z"}`)
	if code != http.StatusAccepted || res["accepted"] != true {
		t.Fatalf("expected accepted ingest, got %d: %v", code, res)
	}
	// Duplicate event_id is deduped, not an error.
	code, res = doIntake(t, mux, http.MethodPost, "/ingest",
		`{"eventId":"ev-1","eventType":"valid_view","videoId":"v-1","userId":"u-1","userAgent":"Mozilla/5.0","watchSeconds":120,"playbackRate":1}`)
	if code != http.StatusAccepted || res["accepted"] != false {
		t.Fatalf("expected deduped ingest, got %d: %v", code, res)
	}
	// Bot traffic is filtered, not an error.
	code, res = doIntake(t, mux, http.MethodPost, "/ingest",
		`{"eventId":"ev-2","eventType":"valid_view","videoId":"v-1","userAgent":"HeadlessChrome/120","watchSeconds":5,"playbackRate":1}`)
	if code != http.StatusAccepted || res["accepted"] != false {
		t.Fatalf("expected filtered bot ingest, got %d: %v", code, res)
	}
	// Invalid payloads are rejected.
	code, _ = doIntake(t, mux, http.MethodPost, "/ingest",
		`{"eventId":"ev-3","eventType":"bogus","videoId":"v-1"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported event_type, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/ingest", `{"eventType":"valid_view"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing event_id, got %d", code)
	}

	code, flushed := doIntake(t, mux, http.MethodPost, "/flush", `{}`)
	if code != http.StatusOK || flushed["flushed"] != float64(1) {
		t.Fatalf("expected 1 flushed row, got %d: %v", code, flushed)
	}
	code, flushed = doIntake(t, mux, http.MethodPost, "/flush", `{}`)
	if code != http.StatusOK || flushed["flushed"] != float64(0) {
		t.Fatalf("expected 0 flushed rows on empty buffer, got %d: %v", code, flushed)
	}
}
