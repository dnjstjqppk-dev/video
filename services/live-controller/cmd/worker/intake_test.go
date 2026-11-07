package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newIntakeMux() *http.ServeMux {
	srv := newHealthServer("live-controller", NewLiveController())
	mux := http.NewServeMux()
	srv.registerIntakeRoutes(mux)
	return mux
}

func doIntake(t *testing.T, mux *http.ServeMux, method, path, body string) (int, map[string]interface{}) {
	t.Helper()
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var decoded map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &decoded)
	return rec.Code, decoded
}

func TestIntakeSessionLifecycle(t *testing.T) {
	mux := newIntakeMux()

	code, created := doIntake(t, mux, http.MethodPost, "/sessions", `{"channelId":"ch-1","dvrEnabled":true}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for provision, got %d: %v", code, created)
	}
	sessionID, _ := created["sessionId"].(string)
	if sessionID == "" || created["state"] != "CREATED" {
		t.Fatalf("invalid provisioned session: %v", created)
	}
	if created["rtmpsUrl"] == "" || created["playbackHlsUrl"] == "" {
		t.Fatalf("provisioned session missing ingest/playback URLs: %v", created)
	}

	code, fetched := doIntake(t, mux, http.MethodGet, "/sessions/"+sessionID, "")
	if code != http.StatusOK || fetched["sessionId"] != sessionID {
		t.Fatalf("expected 200 session fetch, got %d: %v", code, fetched)
	}

	code, live := doIntake(t, mux, http.MethodPost, "/sessions/"+sessionID+"/transition", `{"state":"LIVE"}`)
	if code != http.StatusOK || live["state"] != "LIVE" {
		t.Fatalf("expected LIVE transition, got %d: %v", code, live)
	}

	code, rotated := doIntake(t, mux, http.MethodPost, "/sessions/"+sessionID+"/rotate-key", `{}`)
	if code != http.StatusOK || rotated["streamKey"] == "" {
		t.Fatalf("expected rotated stream key, got %d: %v", code, rotated)
	}

	code, ended := doIntake(t, mux, http.MethodPost, "/sessions/"+sessionID+"/transition", `{"state":"ENDED"}`)
	if code != http.StatusOK || ended["state"] != "ENDED" {
		t.Fatalf("expected ENDED transition, got %d: %v", code, ended)
	}
	if ended["vodVideoId"] == "" {
		t.Fatalf("expected DVR session to mint a VOD id, got %v", ended)
	}

	// Terminal state rejects further transitions and key rotation.
	code, _ = doIntake(t, mux, http.MethodPost, "/sessions/"+sessionID+"/transition", `{"state":"LIVE"}`)
	if code != http.StatusConflict {
		t.Fatalf("expected 409 for transition out of ENDED, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/sessions/"+sessionID+"/rotate-key", `{}`)
	if code != http.StatusConflict {
		t.Fatalf("expected 409 for rotate-key on ENDED, got %d", code)
	}
}

func TestIntakeSessionValidation(t *testing.T) {
	mux := newIntakeMux()

	code, _ := doIntake(t, mux, http.MethodPost, "/sessions", `{"dvrEnabled":true}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing channelId, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodGet, "/sessions/does-not-exist", "")
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown session, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/sessions/does-not-exist/transition", `{"state":"LIVE"}`)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 for transition on unknown session, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/sessions", `{"channelId":"ch-1"}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for explicit-id-less provision, got %d", code)
	}

	code, health := doIntake(t, mux, http.MethodPost, "/sessions/health", `{"bitrateKbps":6000,"fps":60,"droppedPct":0.2}`)
	if code != http.StatusOK || health["status"] != "HEALTHY" {
		t.Fatalf("expected HEALTHY evaluation, got %d: %v", code, health)
	}
	code, health = doIntake(t, mux, http.MethodPost, "/sessions/health", `{"bitrateKbps":300,"fps":8,"droppedPct":30}`)
	if code != http.StatusOK || health["status"] != "UNHEALTHY" {
		t.Fatalf("expected UNHEALTHY evaluation, got %d: %v", code, health)
	}
}
