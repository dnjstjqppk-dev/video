package events

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func TestRecordAnalyticsEvent(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	router := h.Routes()

	vidID := "00000000-0000-0000-0000-000000000001"
	beforeViews := s.Videos[vidID].ViewCount

	payload, _ := json.Marshal(map[string]interface{}{
		"event_id":               "evt_101",
		"event_type":             "valid_view",
		"video_id":               vidID,
		"watch_duration_seconds": 45,
	})

	req := httptest.NewRequest(http.MethodPost, "/analytics", bytes.NewReader(payload))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 Accepted, got %d", rec.Code)
	}
	if s.Videos[vidID].ViewCount != beforeViews+1 {
		t.Fatalf("expected view count increment to %d, got %d", beforeViews+1, s.Videos[vidID].ViewCount)
	}
}

// P1-14: repeat valid_view events from one viewer are accepted but not recounted.
func TestValidViewDeduped(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	router := h.Routes()

	vidID := "00000000-0000-0000-0000-000000000001"
	beforeViews := s.Videos[vidID].ViewCount
	send := func() *httptest.ResponseRecorder {
		payload, _ := json.Marshal(map[string]interface{}{
			"event_type": "valid_view",
			"video_id":   vidID,
		})
		req := httptest.NewRequest(http.MethodPost, "/analytics", bytes.NewReader(payload))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	if rec := send(); rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rec.Code)
	}

	second := send()
	if second.Code != http.StatusAccepted {
		t.Fatalf("repeat telemetry must still be accepted, got %d", second.Code)
	}
	var body struct {
		Counted bool `json:"counted"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Counted {
		t.Fatal("repeat valid_view must report counted=false")
	}
	if s.Videos[vidID].ViewCount != beforeViews+1 {
		t.Fatalf("expected single increment to %d, got %d", beforeViews+1, s.Videos[vidID].ViewCount)
	}
}
