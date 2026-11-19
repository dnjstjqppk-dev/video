package videos_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func TestThumbnailSelect_Validation(t *testing.T) {
	newSeeded := func() *store.MemoryStore {
		s := store.NewMemoryStore()
		v := s.Videos["00000000-0000-0000-0000-000000000001"]
		v.AutoFrameURLs = []string{"https://cdn.example/hls/v/auto_1.jpg", "https://cdn.example/hls/v/auto_2.jpg"}
		return s
	}
	selectSource := func(s *store.MemoryStore, token, source string) (int, map[string]interface{}) {
		router := scheduledRouter(s)
		req, rec := videosReq(http.MethodPost, "/videos/00000000-0000-0000-0000-000000000001/thumbnail/select", token,
			map[string]string{"source": source})
		router.ServeHTTP(rec, req)
		var out map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}

	if code, _ := selectSource(newSeeded(), "", "auto_1"); code != http.StatusUnauthorized {
		t.Fatalf("anonymous select must be 401, got %d", code)
	}
	if code, _ := selectSource(newSeeded(), "viewer-token", "auto_1"); code != http.StatusForbidden {
		t.Fatalf("non-owner select must be 403, got %d", code)
	}
	if code, _ := selectSource(newSeeded(), "creator-token", "auto_9"); code != http.StatusBadRequest {
		t.Fatalf("unknown frame must be 400, got %d", code)
	}
	if code, _ := selectSource(newSeeded(), "creator-token", "custom"); code != http.StatusConflict {
		t.Fatalf("missing custom must be 409, got %d", code)
	}
	s := newSeeded()
	if code, out := selectSource(s, "creator-token", "auto_2"); code != http.StatusOK {
		t.Fatalf("valid select must be 200, got %d: %v", code, out)
	} else if out["thumbnail_url"] != "https://cdn.example/hls/v/auto_2.jpg" {
		t.Fatalf("wrong thumbnail url: %v", out)
	}
	if got := s.Videos["00000000-0000-0000-0000-000000000001"].ThumbnailURL; got != "https://cdn.example/hls/v/auto_2.jpg" {
		t.Fatal("thumbnail was not applied to the video")
	}
}

func TestThumbnailUpload_RequiresStorage(t *testing.T) {
	s := store.NewMemoryStore()
	router := scheduledRouter(s)
	t.Setenv("R2_ACCESS_KEY_ID", "")
	t.Setenv("R2_SECRET_ACCESS_KEY", "")
	req, rec := videosReq(http.MethodPost, "/videos/00000000-0000-0000-0000-000000000001/thumbnail", "creator-token", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=x")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("unconfigured storage must be 503, got %d", rec.Code)
	}
}

func TestGetFrames_ListsKnownSources(t *testing.T) {
	s := store.NewMemoryStore()
	v := s.Videos["00000000-0000-0000-0000-000000000001"]
	v.AutoFrameURLs = []string{"https://cdn.example/a1.jpg"}
	v.CustomThumbnailURL = "https://cdn.example/custom.jpg"
	router := scheduledRouter(s)
	req, rec := videosReq(http.MethodGet, "/videos/"+v.ID+"/frames", "creator-token", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("frames must be 200, got %d", rec.Code)
	}
	var out struct {
		Current string `json:"current"`
		Frames  []struct {
			Source string `json:"source"`
			URL    string `json:"url"`
		} `json:"frames"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode frames: %v", err)
	}
	sources := map[string]bool{}
	for _, f := range out.Frames {
		sources[f.Source] = true
	}
	if !sources["auto_1"] || !sources["custom"] {
		t.Fatalf("expected auto_1 and custom frames, got %+v", out.Frames)
	}
}
