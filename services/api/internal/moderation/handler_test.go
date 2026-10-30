package moderation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func TestModerationQueueAndResolve(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := h.Routes(auth)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from moderation queue, got %d", rec.Code)
	}

	resolveBody, _ := json.Marshal(map[string]string{
		"action": "approve",
		"notes":  "Verified safe content",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/mod-case-1/resolve", bytes.NewReader(resolveBody))
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 OK resolving moderation case, got %d", rec2.Code)
	}
}
