package recommendations

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func TestRecommendationsAndSearch(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := h.Routes(auth)

	for _, path := range []string{"/home", "/shorts", "/subscriptions"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for %s, got %d", path, rec.Code)
		}
	}

	fbBody, _ := json.Marshal(map[string]string{
		"video_id": "00000000-0000-0000-0000-000000000001",
		"action":   "not_interested",
	})
	reqFb := httptest.NewRequest(http.MethodPost, "/feedback", bytes.NewReader(fbBody))
	recFb := httptest.NewRecorder()
	router.ServeHTTP(recFb, reqFb)
	if recFb.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for /feedback, got %d", recFb.Code)
	}
}
