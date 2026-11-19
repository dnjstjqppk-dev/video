package videos_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/videos"
)

func scheduledRouter(s *store.MemoryStore) *chi.Mux {
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	router := chi.NewRouter()
	router.Use(auth.Authenticate)
	router.Mount("/videos", videos.NewHandler(s).Routes(auth))
	return router
}

func TestUpdateVideo_ScheduledValidation(t *testing.T) {
	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	past := time.Now().Add(-48 * time.Hour).Format(time.RFC3339)
	for _, tc := range []struct {
		name   string
		body   map[string]string
		status int
	}{
		{"scheduled_without_time", map[string]string{"visibility": "scheduled"}, http.StatusBadRequest},
		{"scheduled_past_time", map[string]string{"visibility": "scheduled", "scheduled_at": past}, http.StatusBadRequest},
		{"scheduled_bad_format", map[string]string{"visibility": "scheduled", "scheduled_at": "tomorrow"}, http.StatusBadRequest},
		{"scheduled_future", map[string]string{"visibility": "scheduled", "scheduled_at": future}, http.StatusOK},
		{"scheduled_camel_key", map[string]string{"visibility": "scheduled", "scheduledAt": future}, http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := store.NewMemoryStore()
			router := scheduledRouter(s)
			videoID := "00000000-0000-0000-0000-000000000001"
			req, rec := videosReq(http.MethodPatch, "/videos/"+videoID, "creator-token", tc.body)
			router.ServeHTTP(rec, req)
			if rec.Code != tc.status {
				t.Fatalf("expected %d, got %d: %s", tc.status, rec.Code, rec.Body.String())
			}
			v := s.Videos[videoID]
			if tc.status == http.StatusOK {
				if v.Visibility != "scheduled" || v.ScheduledAt == nil {
					t.Fatal("scheduled state was not stored")
				}
			} else if v.Visibility == "scheduled" {
				t.Fatal("rejected request must not flip visibility")
			}
		})
	}
}

func TestUpdateVideo_LeavingScheduledClearsTime(t *testing.T) {
	s := store.NewMemoryStore()
	router := scheduledRouter(s)
	videoID := "00000000-0000-0000-0000-000000000001"
	future := time.Now().Add(48 * time.Hour).Format(time.RFC3339)
	req, rec := videosReq(http.MethodPatch, "/videos/"+videoID, "creator-token",
		map[string]string{"visibility": "scheduled", "scheduled_at": future})
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("schedule failed: %s", rec.Body.String())
	}
	req, rec = videosReq(http.MethodPatch, "/videos/"+videoID, "creator-token",
		map[string]string{"visibility": "private"})
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("un-schedule failed: %s", rec.Body.String())
	}
	if v := s.Videos[videoID]; v.ScheduledAt != nil {
		t.Fatal("leaving scheduled must clear the scheduled time")
	}
}
