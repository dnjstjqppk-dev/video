package videos_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func TestVideoElements_CRUDAndValidation(t *testing.T) {
	s := store.NewMemoryStore()
	router := scheduledRouter(s)
	videoID := "00000000-0000-0000-0000-000000000001" // duration 754

	add := func(token string, body map[string]interface{}) (int, map[string]interface{}) {
		req, rec := videosReq(http.MethodPost, "/videos/"+videoID+"/elements", token, body)
		router.ServeHTTP(rec, req)
		var out map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}

	if code, _ := add("", map[string]interface{}{"kind": "card", "title": "x", "target_type": "video", "target": "v", "start_sec": 1, "end_sec": 5}); code != http.StatusUnauthorized {
		t.Fatalf("anonymous add must be 401, got %d", code)
	}
	if code, _ := add("viewer-token", map[string]interface{}{"kind": "card", "title": "x", "target_type": "video", "target": "v", "start_sec": 1, "end_sec": 5}); code != http.StatusForbidden {
		t.Fatalf("non-owner add must be 403, got %d", code)
	}
	for name, body := range map[string]map[string]interface{}{
		"bad_kind":    {"kind": "poll", "title": "x", "target_type": "video", "target": "v", "start_sec": 1, "end_sec": 5},
		"empty_title": {"kind": "card", "title": " ", "target_type": "video", "target": "v", "start_sec": 1, "end_sec": 5},
		"bad_target":  {"kind": "card", "title": "x", "target_type": "url", "target": "ftp://x", "start_sec": 1, "end_sec": 5},
		"bad_range":   {"kind": "card", "title": "x", "target_type": "video", "target": "v", "start_sec": 9, "end_sec": 9},
		"past_end":    {"kind": "card", "title": "x", "target_type": "video", "target": "v", "start_sec": 700, "end_sec": 9999},
	} {
		if code, _ := add("creator-token", body); code != http.StatusBadRequest {
			t.Fatalf("%s must be 400, got %d", name, code)
		}
	}

	code, created := add("creator-token", map[string]interface{}{
		"kind": "end_screen", "title": "Next up", "target_type": "video",
		"target": "vid-2", "start_sec": 700, "end_sec": 754,
	})
	if code != http.StatusCreated {
		t.Fatalf("valid add must be 201, got %d", code)
	}
	elID, _ := created["id"].(string)
	if elID == "" {
		t.Fatal("created element needs an id")
	}

	req, rec := videosReq(http.MethodGet, "/videos/"+videoID+"/elements", "creator-token", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list must be 200, got %d", rec.Code)
	}
	var listed struct {
		Items []map[string]interface{} `json:"items"`
		Total int                      `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listed); err != nil || listed.Total != 1 {
		t.Fatalf("expected 1 element listed: %v", rec.Body.String())
	}

	req, rec = videosReq(http.MethodDelete, "/videos/"+videoID+"/elements/does-not-exist", "creator-token", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown element delete must be 404, got %d", rec.Code)
	}
	req, rec = videosReq(http.MethodDelete, "/videos/"+videoID+"/elements/"+elID, "creator-token", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete must be 200, got %d", rec.Code)
	}
	if len(s.Videos[videoID].Elements) != 0 {
		t.Fatal("element was not removed")
	}
}
