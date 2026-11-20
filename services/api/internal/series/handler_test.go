package series_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/series"
	"globalvideoplatform/services/api/internal/store"
)

const (
	testChannelID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	testVideoID   = "00000000-0000-0000-0000-000000000001"
	testAIVideoID = "00000000-0000-0000-0000-0000000000a1"
)

func newSeriesRouter() (http.Handler, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/series", series.NewHandler(memStore).Routes(auth))
	return r, memStore
}

func seriesReq(t *testing.T, router http.Handler, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var rbody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rbody = bytes.NewReader(b)
	} else {
		rbody = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rbody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decodeBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v: %s", err, rec.Body.String())
	}
	return out
}

func TestCreateSeries_RequiresAuth(t *testing.T) {
	router, _ := newSeriesRouter()
	rec := seriesReq(t, router, http.MethodPost, "/series/", "", map[string]string{"title": "x"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateSeries_ForbiddenOnForeignChannel(t *testing.T) {
	router, _ := newSeriesRouter()
	rec := seriesReq(t, router, http.MethodPost, "/series/", "viewer-token", map[string]string{
		"channel_id": testChannelID, "title": "Hijack",
	})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateSeries_ValidatesTitle(t *testing.T) {
	router, _ := newSeriesRouter()
	rec := seriesReq(t, router, http.MethodPost, "/series/", "creator-token", map[string]string{
		"channel_id": testChannelID, "title": "  ",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func createSeries(t *testing.T, router http.Handler) string {
	t.Helper()
	rec := seriesReq(t, router, http.MethodPost, "/series/", "creator-token", map[string]string{
		"channel_id": testChannelID, "title": "Go Deep Dives", "status": "ongoing",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	s, ok := body["series"].(map[string]interface{})
	if !ok || s["id"] == "" {
		t.Fatalf("missing series id in %v", body)
	}
	return s["id"].(string)
}

func TestSeries_CRUDAndEpisodes(t *testing.T) {
	router, memStore := newSeriesRouter()
	id := createSeries(t, router)

	// Attach two episodes out of order; detail must sort by episode_no.
	for _, ep := range []map[string]interface{}{
		{"video_id": testVideoID, "season_no": 1, "episode_no": 2},
	} {
		rec := seriesReq(t, router, http.MethodPost, "/series/"+id+"/episodes", "creator-token", ep)
		if rec.Code != http.StatusOK {
			t.Fatalf("attach: expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	}
	secondID := "00000000-0000-0000-0000-000000000002"
	memStore.Videos[secondID] = &store.Video{
		ID: secondID, ChannelID: testChannelID, Title: "Second",
		Visibility: "public", Status: "READY",
	}
	rec := seriesReq(t, router, http.MethodPost, "/series/"+id+"/episodes", "creator-token",
		map[string]interface{}{"video_id": secondID, "season_no": 1, "episode_no": 1})
	if rec.Code != http.StatusOK {
		t.Fatalf("attach second: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	body := decodeBody(t, rec)
	eps := body["episodes"].([]interface{})
	if len(eps) != 2 {
		t.Fatalf("expected 2 episodes, got %d", len(eps))
	}
	if eps[0].(map[string]interface{})["video_id"] != secondID {
		t.Fatalf("episodes not sorted: %v", eps)
	}
	if body["episode_count"].(float64) != 2 {
		t.Fatalf("episode_count = %v, want 2", body["episode_count"])
	}

	// Public detail works without auth.
	rec = seriesReq(t, router, http.MethodGet, "/series/"+id, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("detail: expected 200, got %d", rec.Code)
	}

	// Channel filter lists the new series.
	rec = seriesReq(t, router, http.MethodGet, "/series/?channel_id="+testChannelID, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	if decodeBody(t, rec)["total"].(float64) < 1 {
		t.Fatal("expected at least one series for the channel")
	}

	// Status update.
	rec = seriesReq(t, router, http.MethodPatch, "/series/"+id, "creator-token", map[string]string{"status": "completed"})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if decodeBody(t, rec)["series"].(map[string]interface{})["status"] != "completed" {
		t.Fatal("status did not update")
	}

	// Detach one episode.
	rec = seriesReq(t, router, http.MethodDelete, "/series/"+id+"/episodes/"+secondID, "creator-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("detach: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if decodeBody(t, rec)["episode_count"].(float64) != 1 {
		t.Fatal("expected 1 episode after detach")
	}
	if memStore.Videos[secondID].SeriesID != "" {
		t.Fatal("detached video still references the series")
	}

	// Delete detaches the rest.
	rec = seriesReq(t, router, http.MethodDelete, "/series/"+id, "creator-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Videos[testVideoID].SeriesID != "" {
		t.Fatal("episode still references the deleted series")
	}
	rec = seriesReq(t, router, http.MethodGet, "/series/"+id, "", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", rec.Code)
	}
}

func TestAttachEpisode_RejectsCrossChannel(t *testing.T) {
	router, _ := newSeriesRouter()
	id := createSeries(t, router)
	rec := seriesReq(t, router, http.MethodPost, "/series/"+id+"/episodes", "creator-token",
		map[string]interface{}{"video_id": testAIVideoID, "season_no": 1, "episode_no": 1})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAttachEpisode_ValidatesSlot(t *testing.T) {
	router, _ := newSeriesRouter()
	id := createSeries(t, router)
	rec := seriesReq(t, router, http.MethodPost, "/series/"+id+"/episodes", "creator-token",
		map[string]interface{}{"video_id": testVideoID, "season_no": 0, "episode_no": 1})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for season 0, got %d", rec.Code)
	}
}

func TestSeries_ForbiddenForOtherOwner(t *testing.T) {
	router, _ := newSeriesRouter()
	id := createSeries(t, router)
	rec := seriesReq(t, router, http.MethodPatch, "/series/"+id, "viewer-token", map[string]string{"title": "x"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	rec = seriesReq(t, router, http.MethodDelete, "/series/"+id, "viewer-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}
