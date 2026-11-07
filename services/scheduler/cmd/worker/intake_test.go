package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newIntakeMux() *http.ServeMux {
	srv := newHealthServer("scheduler", NewSchedulerEngine())
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

func TestIntakeEnqueueAndTick(t *testing.T) {
	mux := newIntakeMux()

	code, ev := doIntake(t, mux, http.MethodPost, "/outbox", `{"topic":"video.published","payload":"{\"id\":\"v-1\"}"}`)
	if code != http.StatusCreated || ev["id"] == "" || ev["published"] != false {
		t.Fatalf("expected staged outbox event, got %d: %v", code, ev)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/scheduled-videos",
		`{"id":"v-due","publishAt":"2026-11-07T10:00:00Z"}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for scheduled video, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/scheduled-videos",
		`{"id":"v-future","publishAt":"2030-01-01T00:00:00Z"}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for future video, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/uploads",
		`{"uploadId":"up-stale","expiresAt":"2026-11-07T09:00:00Z"}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for tracked upload, got %d", code)
	}

	code, tick := doIntake(t, mux, http.MethodPost, "/tick", `{"now":"2026-11-07T12:00:00Z"}`)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for tick, got %d: %v", code, tick)
	}
	if tick["relayedEvents"] != float64(1) || tick["publishedVideos"] != float64(1) || tick["expiredUploads"] != float64(1) {
		t.Fatalf("unexpected tick counts: %v", tick)
	}
	// Second tick is idempotent: nothing left to sweep.
	code, tick = doIntake(t, mux, http.MethodPost, "/tick", `{"now":"2026-11-07T12:00:00Z"}`)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for second tick, got %d", code)
	}
	if tick["relayedEvents"] != float64(0) || tick["publishedVideos"] != float64(0) || tick["expiredUploads"] != float64(0) {
		t.Fatalf("expected zero counts on re-tick, got %v", tick)
	}
}

func TestIntakeSettlementsAndValidation(t *testing.T) {
	mux := newIntakeMux()

	code, _ := doIntake(t, mux, http.MethodPost, "/settlements", `{"periodId":"2026-11"}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for open settlement, got %d", code)
	}
	code, closed := doIntake(t, mux, http.MethodPost, "/settlements/2026-11/close", `{}`)
	if code != http.StatusOK || closed["status"] != "closed" {
		t.Fatalf("expected closed settlement, got %d: %v", code, closed)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/settlements/2026-11/close", `{}`)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 for double close, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/settlements/never-opened/close", `{}`)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown period, got %d", code)
	}

	code, _ = doIntake(t, mux, http.MethodPost, "/outbox", `{"payload":"x"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing topic, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/scheduled-videos", `{"id":"v-x","publishAt":"not-a-time"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad publishAt, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/tick", `{"now":"not-a-time"}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad tick time, got %d", code)
	}
}
