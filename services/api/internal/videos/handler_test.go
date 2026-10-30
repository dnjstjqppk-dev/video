package videos_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/videos"
)

func newVideosTestServer() *httptest.Server {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	handler := videos.NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/videos", handler.Routes(auth))

	return httptest.NewServer(r)
}

func videosReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var rbody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rbody = bytes.NewReader(b)
	} else {
		rbody = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rbody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, httptest.NewRecorder()
}

func TestGetVideo_ReturnsVideo(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodGet, "/videos/"+videoID, "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetPlayback_ReturnsManifest(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodGet, "/videos/"+videoID+"/playback", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected JSON content type, got %s", rec.Header().Get("Content-Type"))
	}
}

func TestUpdateVideo_RequiresAuth(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodPatch, "/videos/"+videoID, "", map[string]string{"title": "hacked"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestUpdateVideo_AuthenticatedCanUpdate(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodPatch, "/videos/"+videoID, "creator-token", map[string]string{"title": "Updated Title"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteVideo_RequiresAuth(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodDelete, "/videos/"+videoID, "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestDeleteVideo_AuthenticatedCanDelete(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodDelete, "/videos/"+videoID, "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestLikeVideo_RequiresAuth(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodPost, "/videos/"+videoID+"/like", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestLikeVideo_Authenticated(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := videosReq(http.MethodPost, "/videos/"+videoID+"/like", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
