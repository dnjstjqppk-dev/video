package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newIntakeMux() *http.ServeMux {
	srv := newHealthServer("recommendation-worker")
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

func TestIntakeRankOrdersAndFilters(t *testing.T) {
	mux := newIntakeMux()
	payload := `{"user":{"userId":"u-1","preferredCategory":"Tech","ageVerified":false,"blockedChannels":["ch-blocked"]},"candidates":[
		{"id":"v-old","channelId":"ch-a","category":"Cooking","status":"READY","visibility":"public","impressions":50000,"avgWatchDurationSec":200,"publishedAt":"2020-01-01T00:00:00Z"},
		{"id":"v-fresh","channelId":"ch-b","category":"Tech","status":"READY","visibility":"public","impressions":50,"avgWatchDurationSec":180,"publishedAt":"2026-11-07T10:00:00Z"},
		{"id":"v-private","channelId":"ch-c","category":"Tech","status":"READY","visibility":"private","impressions":10,"avgWatchDurationSec":300,"publishedAt":"2026-11-07T10:00:00Z"},
		{"id":"v-blocked","channelId":"ch-blocked","category":"Tech","status":"READY","visibility":"public","impressions":10,"avgWatchDurationSec":300,"publishedAt":"2026-11-07T10:00:00Z"},
		{"id":"v-age-gated","channelId":"ch-d","category":"Tech","status":"READY","visibility":"public","ageRestricted":true,"impressions":10,"avgWatchDurationSec":300,"publishedAt":"2026-11-07T10:00:00Z"}
	]}`
	code, res := doIntake(t, mux, http.MethodPost, "/rank", payload)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for rank, got %d: %v", code, res)
	}
	ranked, _ := res["ranked"].([]interface{})
	if len(ranked) != 2 {
		t.Fatalf("expected 2 eligible candidates, got %v", res)
	}
	first, _ := ranked[0].(map[string]interface{})
	second, _ := ranked[1].(map[string]interface{})
	if first["id"] != "v-fresh" || second["id"] != "v-old" {
		t.Fatalf("expected fresh Tech video first, got %v", ranked)
	}
	if first["score"] == nil || second["score"] == nil {
		t.Fatalf("expected scores on ranked items, got %v", ranked)
	}

	// Limit truncates the ranked list.
	code, res = doIntake(t, mux, http.MethodPost, "/rank", `{"limit":1,"user":{},"candidates":[
		{"id":"a","channelId":"c1","status":"READY","visibility":"public"},
		{"id":"b","channelId":"c2","status":"READY","visibility":"public"}]}`)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for limited rank, got %d", code)
	}
	if ranked, _ := res["ranked"].([]interface{}); len(ranked) != 1 {
		t.Fatalf("expected 1 ranked item with limit=1, got %v", res)
	}

	// Empty candidates are rejected.
	code, _ = doIntake(t, mux, http.MethodPost, "/rank", `{"user":{},"candidates":[]}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty candidates, got %d", code)
	}
}
