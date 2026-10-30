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
