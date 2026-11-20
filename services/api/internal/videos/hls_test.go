package videos_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/videos"
)

func hlsRouter(s *store.MemoryStore) *chi.Mux {
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	router := chi.NewRouter()
	router.Use(auth.Authenticate)
	router.Mount("/hls", videos.NewHandler(s).HLSRoutes())
	return router
}

func TestServeHLS_GatesAndStreams(t *testing.T) {
	fakeS3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/gvp-public-hls/hls/00000000-0000-0000-0000-000000000001/master.m3u8":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("#EXTM3U\n"))
		case "/gvp-public-hls/hls/00000000-0000-0000-0000-000000000099/poster.jpg":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("JPEG"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fakeS3.Close()
	t.Setenv("R2_ENDPOINT", fakeS3.URL)
	t.Setenv("R2_ACCESS_KEY_ID", "test")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test")

	s := store.NewMemoryStore()
	priv := *s.Videos["00000000-0000-0000-0000-000000000001"]
	priv.ID = "00000000-0000-0000-0000-000000000099"
	priv.Visibility = "private"
	s.Videos[priv.ID] = &priv
	router := hlsRouter(s)

	get := func(path, token string) (int, http.Header, string) {
		req, rec := videosReq(http.MethodGet, path, token, nil)
		router.ServeHTTP(rec, req)
		return rec.Code, rec.Header(), rec.Body.String()
	}

	if code, h, body := get("/hls/00000000-0000-0000-0000-000000000001/master.m3u8", ""); code != http.StatusOK || body != "#EXTM3U\n" {
		t.Fatalf("public master must stream, got %d %q", code, body)
	} else if ct := h.Get("Content-Type"); ct != "application/vnd.apple.mpegurl" {
		t.Fatalf("wrong content type %q", ct)
	}
	if code, _, _ := get("/hls/00000000-0000-0000-0000-000000000001/missing.ts", ""); code != http.StatusNotFound {
		t.Fatalf("missing artifact must be 404, got %d", code)
	}
	if code, _, _ := get("/hls/00000000-0000-0000-0000-000000000001/../secret", ""); code != http.StatusBadRequest {
		t.Fatalf("traversal must be 400, got %d", code)
	}
	if code, _, _ := get("/hls/does-not-exist/master.m3u8", ""); code != http.StatusNotFound {
		t.Fatalf("unknown video must be 404, got %d", code)
	}
	if code, _, _ := get("/hls/00000000-0000-0000-0000-000000000099/master.m3u8", ""); code != http.StatusNotFound {
		t.Fatalf("private video anon must be 404, got %d", code)
	}
	if code, _, body := get("/hls/00000000-0000-0000-0000-000000000099/poster.jpg", ""); code != http.StatusOK || body != "JPEG" {
		t.Fatalf("private video poster must stay reachable, got %d %q", code, body)
	}
	if _, h, _ := get("/hls/00000000-0000-0000-0000-000000000099/poster.jpg", ""); h.Get("Cross-Origin-Resource-Policy") != "cross-origin" {
		t.Fatalf("gateway stills must be embeddable cross-origin, got CORP %q", h.Get("Cross-Origin-Resource-Policy"))
	}
	if _, h, _ := get("/hls/00000000-0000-0000-0000-000000000001/master.m3u8", ""); h.Get("Cross-Origin-Resource-Policy") != "cross-origin" {
		t.Fatalf("gateway playlists must be fetchable cross-origin, got CORP %q", h.Get("Cross-Origin-Resource-Policy"))
	}
	if code, _, _ := get("/hls/00000000-0000-0000-0000-000000000099/master.m3u8", "creator-token"); code != http.StatusNotFound {
		t.Fatalf("private video missing key must be 404 even for owner, got %d", code)
	}
}

func TestServeHLS_PlaybackTokenAuthorizes(t *testing.T) {
	privID := "00000000-0000-0000-0000-000000000099"
	fakeS3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-MAP:URI=\"init.mp4\"\n240p/index.m3u8\nhttps://cdn.example/x.m4s\n"))
	}))
	defer fakeS3.Close()
	t.Setenv("R2_ENDPOINT", fakeS3.URL)
	t.Setenv("R2_ACCESS_KEY_ID", "test")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test")

	s := store.NewMemoryStore()
	base := *s.Videos["00000000-0000-0000-0000-000000000001"]
	base.ID = privID
	base.Visibility = "private"
	s.Videos[privID] = &base
	router := hlsRouter(s)

	token, err := videos.SignPlaybackToken(privID, time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	other, err := videos.SignPlaybackToken("00000000-0000-0000-0000-000000000001", time.Hour)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	get := func(target string) (int, string) {
		req, rec := videosReq(http.MethodGet, target, "", nil)
		router.ServeHTTP(rec, req)
		return rec.Code, rec.Body.String()
	}
	if code, _ := get("/hls/" + privID + "/master.m3u8"); code != http.StatusNotFound {
		t.Fatalf("bare private master must be 404, got %d", code)
	}
	if code, _ := get("/hls/" + privID + "/master.m3u8?token=garbage"); code != http.StatusNotFound {
		t.Fatalf("garbage token must be 404, got %d", code)
	}
	if code, _ := get("/hls/" + privID + "/master.m3u8?token=" + other); code != http.StatusNotFound {
		t.Fatalf("foreign token must be 404, got %d", code)
	}
	code, body := get("/hls/" + privID + "/master.m3u8?token=" + token)
	if code != http.StatusOK {
		t.Fatalf("token master must be 200, got %d", code)
	}
	for _, want := range []string{"240p/index.m3u8?token=" + token, `URI="init.mp4?token=` + token + `"`, "https://cdn.example/x.m4s\n"} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing %q in %q", want, body)
		}
	}
}
