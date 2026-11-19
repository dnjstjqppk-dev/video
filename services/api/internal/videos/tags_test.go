package videos_test

import (
	"net/http"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func TestUpdateVideo_TagsReplaceSanitized(t *testing.T) {
	s := store.NewMemoryStore()
	router := scheduledRouter(s)
	videoID := "00000000-0000-0000-0000-000000000001"

	many := make([]interface{}, 0, 15)
	for i := 0; i < 15; i++ {
		many = append(many, "tag")
	}
	req, rec := videosReq(http.MethodPatch, "/videos/"+videoID, "creator-token", map[string]interface{}{
		"tags": append([]interface{}{" <i>hi</i> ", ""}, many...),
	})
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	got := s.Videos[videoID].Tags
	if len(got) != 2 || got[0] != "hi" || got[1] != "tag" {
		t.Fatalf("expected sanitized [hi tag], got %v", got)
	}

	// Absent tags leave the list untouched.
	req, rec = videosReq(http.MethodPatch, "/videos/"+videoID, "creator-token", map[string]string{"title": "Kept tags"})
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if len(s.Videos[videoID].Tags) != 2 {
		t.Fatalf("absent tags must not clear the list: %v", s.Videos[videoID].Tags)
	}
}
