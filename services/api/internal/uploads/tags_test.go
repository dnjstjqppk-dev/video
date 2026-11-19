package uploads_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/uploads"
)

func TestCreateUploadSession_StoresSanitizedTags(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	s := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	router := chi.NewRouter()
	router.Use(auth.Authenticate)
	router.Mount("/uploads", uploads.NewHandler(s).Routes(auth))

	body := map[string]interface{}{
		"title":         "Tagged upload",
		"description":   "d",
		"fileSizeBytes": 1024,
		"mimeType":      "video/mp4",
		"videoType":     "standard",
		"tags":          []string{" go ", "", "go", "<b>x</b>", "hls"},
	}
	req, rec := uploadReq(http.MethodPost, "/uploads", "creator-token", body)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var out struct {
		VideoID string `json:"videoId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode session: %v", err)
	}
	got := s.Videos[out.VideoID].Tags
	want := []string{"go", "x", "hls"}
	if len(got) != len(want) {
		t.Fatalf("expected tags %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected tags %v, got %v", want, got)
		}
	}
}
