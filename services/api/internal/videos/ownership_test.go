package videos_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/videos"
)

func TestVideoMutationOwnership(t *testing.T) {
	for _, method := range []string{http.MethodPatch, http.MethodDelete} {
		for _, scenario := range []struct {
			name           string
			token          string
			missingChannel bool
			status         int
		}{
			{"anonymous", "", false, http.StatusUnauthorized},
			{"owner", "creator-token", false, http.StatusOK},
			{"other_creator", "other-creator", false, http.StatusForbidden},
			{"administrator", "admin-token", false, http.StatusOK},
			{"missing_channel", "creator-token", true, http.StatusForbidden},
		} {
			t.Run(method+"/"+scenario.name, func(t *testing.T) {
				s := store.NewMemoryStore()
				s.Users["other-creator"] = &store.User{ID: "other-creator", Role: "creator", IsActive: true}
				video := s.Videos["00000000-0000-0000-0000-000000000001"]
				original := *video
				if scenario.missingChannel {
					delete(s.Channels, video.ChannelID)
				}
				auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
				router := chi.NewRouter()
				router.Use(auth.Authenticate)
				router.Mount("/videos", videos.NewHandler(s).Routes(auth))
				req, _ := videosReq(method, "/videos/"+video.ID, scenario.token, map[string]string{"title": "Changed"})
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				if rec.Code != scenario.status {
					t.Fatalf("expected %d, got %d: %s", scenario.status, rec.Code, rec.Body.String())
				}
				if scenario.status != http.StatusOK && (video.Title != original.Title || video.Status != original.Status) {
					t.Fatal("denied request mutated video")
				}
				if scenario.status == http.StatusOK {
					if method == http.MethodPatch && video.Title != "Changed" {
						t.Fatal("authorized update did not change title")
					}
					if method == http.MethodDelete && video.Status != "DELETED" {
						t.Fatal("authorized delete did not delete video")
					}
				}
			})
		}
	}
}
