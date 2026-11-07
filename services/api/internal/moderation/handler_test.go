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

func TestReportContent(t *testing.T) {
	setup := func() (*store.MemoryStore, *Handler, *identity.AuthMiddleware) {
		s := store.NewMemoryStore()
		h := NewHandler(s)
		auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
		return s, h, auth
	}
	call := func(auth *identity.AuthMiddleware, h *Handler, token string, body interface{}) *httptest.ResponseRecorder {
		b, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/moderation/reports", bytes.NewReader(b))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		auth.Authenticate(auth.RequireAuth(http.HandlerFunc(h.ReportContent))).ServeHTTP(rec, req)
		return rec
	}

	t.Run("files report and queues case", func(t *testing.T) {
		s, h, auth := setup()
		before := len(s.Moderation)
		rec := call(auth, h, "viewer-token", map[string]string{
			"targetType": "video",
			"targetId":   "vid_1",
			"reason":     "spam_or_misleading",
			"details":    "Fake giveaway links in description",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if len(s.Moderation) != before+1 {
			t.Fatalf("expected queue growth, got %d -> %d", before, len(s.Moderation))
		}
	})

	t.Run("rejects unauthenticated", func(t *testing.T) {
		_, h, auth := setup()
		rec := call(auth, h, "", map[string]string{"targetType": "video", "targetId": "v", "reason": "x"})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("rejects bad target type", func(t *testing.T) {
		_, h, auth := setup()
		rec := call(auth, h, "viewer-token", map[string]string{"targetType": "playlist", "targetId": "p", "reason": "x"})
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("strips markup from reason", func(t *testing.T) {
		s, h, auth := setup()
		rec := call(auth, h, "viewer-token", map[string]string{
			"targetType": "comment", "targetId": "c1", "reason": "<img src=x onerror=alert(1)>spam",
		})
		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d", rec.Code)
		}
		for _, c := range s.Moderation {
			if c.TargetID == "c1" && (c.Reason != "spam" || len(c.Reason) == 0) {
				t.Fatalf("expected sanitized reason, got %q", c.Reason)
			}
		}
	})
}
