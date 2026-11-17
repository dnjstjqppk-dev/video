package notifications

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func notifRouter(s *store.MemoryStore) http.Handler {
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/notifications", NewHandler(s).Routes(auth))
	return r
}

// P1-16: comment activity notifies owners and parent authors, nobody else.
func TestCommentNotificationsFired(t *testing.T) {
	s := store.NewMemoryStore()

	// Top-level comment by the viewer on the creator's video.
	Notify(s, "22222222-2222-2222-2222-222222222222", "new_comment",
		"New comment on V", "hello", "comment", "c1")

	req := httptest.NewRequest(http.MethodGet, "/notifications/", nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec := httptest.NewRecorder()
	notifRouter(s).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct {
		Items  []store.Notification `json:"items"`
		Total  int                  `json:"total"`
		Unread int                  `json:"unread"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Total != 1 || body.Unread != 1 || body.Items[0].Type != "new_comment" {
		t.Fatalf("expected 1 unread new_comment, got %+v", body)
	}

	// Stranger's inbox is empty; marking read works.
	req2 := httptest.NewRequest(http.MethodGet, "/notifications/", nil)
	req2.Header.Set("Authorization", "Bearer viewer-token")
	rec2 := httptest.NewRecorder()
	notifRouter(s).ServeHTTP(rec2, req2)
	var stranger struct {
		Total int `json:"total"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &stranger)
	if stranger.Total != 0 {
		t.Fatalf("stranger inbox must be empty, got %+v", stranger)
	}

	mark := httptest.NewRequest(http.MethodPost, "/notifications/"+body.Items[0].ID+"/read", bytes.NewReader(nil))
	mark.Header.Set("Authorization", "Bearer creator-token")
	markRec := httptest.NewRecorder()
	notifRouter(s).ServeHTTP(markRec, mark)
	if markRec.Code != http.StatusOK {
		t.Fatalf("expected 200 mark read, got %d", markRec.Code)
	}
	if !s.Notifications["22222222-2222-2222-2222-222222222222"][0].Read {
		t.Fatal("notification must be marked read")
	}

	// Anonymous and cross-user reads are refused.
	anon := httptest.NewRequest(http.MethodGet, "/notifications/", nil)
	anonRec := httptest.NewRecorder()
	notifRouter(s).ServeHTTP(anonRec, anon)
	if anonRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 anonymous, got %d", anonRec.Code)
	}
	cross := httptest.NewRequest(http.MethodPost, "/notifications/"+body.Items[0].ID+"/read", bytes.NewReader(nil))
	cross.Header.Set("Authorization", "Bearer viewer-token")
	crossRec := httptest.NewRecorder()
	notifRouter(s).ServeHTTP(crossRec, cross)
	if crossRec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 cross-user read, got %d", crossRec.Code)
	}
}
