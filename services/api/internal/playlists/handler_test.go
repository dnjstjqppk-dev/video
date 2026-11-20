package playlists

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func TestPlaylistCreateAndGet(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	body, _ := json.Marshal(map[string]string{
		"title":       "Go 1.24 Backend Masterclass",
		"description": "Deep dive into Go concurrency and WebSockets",
		"visibility":  "public",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer creator-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created creating playlist, got %d", rec.Code)
	}

	var created store.Playlist
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || created.ID == "" {
		t.Fatalf("failed to parse created playlist: %v", err)
	}

	reqGet := httptest.NewRequest(http.MethodGet, "/"+created.ID, nil)
	recGet := httptest.NewRecorder()
	router.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusOK {
		t.Fatalf("expected 200 OK getting playlist, got %d", recGet.Code)
	}
}

func TestPlaylistListAndDelete(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))
	call := func(method, path, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	body, _ := json.Marshal(map[string]string{"title": "QA Temp", "visibility": "private"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer viewer-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created store.Playlist
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Owner lists their own.
	rec = call(http.MethodGet, "/", "viewer-token")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), created.ID) {
		t.Fatalf("expected list to include own playlist, got %d %s", rec.Code, rec.Body.String())
	}
	// Others cannot delete it.
	if rec := call(http.MethodDelete, "/"+created.ID, "creator-token"); rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 deleting another user's playlist, got %d", rec.Code)
	}
	// Owner deletes; second delete 404s.
	if rec := call(http.MethodDelete, "/"+created.ID, "viewer-token"); rec.Code != http.StatusOK {
		t.Fatalf("expected 200 deleting own playlist, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec := call(http.MethodDelete, "/"+created.ID, "viewer-token"); rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 deleting twice, got %d", rec.Code)
	}
}
