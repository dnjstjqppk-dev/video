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

const (
	aiOwnerToken = "77777777-7777-7777-7777-777777777777"
	aiEpisode1   = "00000000-0000-0000-0000-0000000000a1"
	aiEpisode2   = "00000000-0000-0000-0000-0000000000a2"
	aiEpisode3   = "00000000-0000-0000-0000-0000000000a3"
	aiPending    = "00000000-0000-0000-0000-0000000000a4"
	humanVideo   = "00000000-0000-0000-0000-000000000001"
	aiSeriesID   = "5e5e5e5e-5e5e-4e5e-8e5e-5e5e5e5e5e5e"
)

func newVideosRouter() (http.Handler, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/videos", videos.NewHandler(memStore).Routes(auth))
	return r, memStore
}

func aiReq(t *testing.T, router http.Handler, method, path, token string, body interface{}) *httptest.ResponseRecorder {
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

func TestSubmitAIReview_RequiresAIFlag(t *testing.T) {
	router, _ := newVideosRouter()
	rec := aiReq(t, router, http.MethodPost, "/videos/"+humanVideo+"/ai-submit", "creator-token", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestSubmitAIReview_Flow(t *testing.T) {
	router, memStore := newVideosRouter()
	vid := memStore.Videos[aiPending]
	vid.AiReviewStatus = "" // fresh AI video, never submitted

	rec := aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-submit", aiOwnerToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("submit: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Videos[aiPending].AiReviewStatus != "pending" {
		t.Fatal("expected status pending after submit")
	}

	rec = aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-submit", aiOwnerToken, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("resubmit: expected 409, got %d", rec.Code)
	}

	// Non-owner cannot submit.
	rec = aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-submit", "creator-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("foreign submit: expected 403, got %d", rec.Code)
	}

	// Non-READY AI videos cannot enter review.
	vid.AiReviewStatus = ""
	vid.Status = "PROCESSING"
	rec = aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-submit", aiOwnerToken, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("processing submit: expected 409, got %d", rec.Code)
	}
	vid.Status = "READY"
}

func TestReviewAI_ApproveFlow(t *testing.T) {
	router, memStore := newVideosRouter()

	// Staff-only: creator cannot review.
	rec := aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-review", aiOwnerToken,
		map[string]string{"decision": "approved"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("creator review: expected 403, got %d: %s", rec.Code, rec.Body.String())
	}

	rec = aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-review", "admin-token",
		map[string]string{"decision": "approved", "note": "Clean provenance, strong pilot."})
	if rec.Code != http.StatusOK {
		t.Fatalf("approve: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if got := memStore.Videos[aiPending].AiReviewStatus; got != "approved" {
		t.Fatalf("status = %q, want approved", got)
	}

	// Reviewing twice is a conflict.
	rec = aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-review", "admin-token",
		map[string]string{"decision": "rejected"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("second review: expected 409, got %d", rec.Code)
	}
}

func TestReviewAI_RejectAndResubmit(t *testing.T) {
	router, memStore := newVideosRouter()
	rec := aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-review", "admin-token",
		map[string]string{"decision": "rejected", "note": "Needs human edit pass."})
	if rec.Code != http.StatusOK {
		t.Fatalf("reject: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Videos[aiPending].AiReviewNote == "" {
		t.Fatal("expected the rejection note to persist")
	}
	// Rejected videos may resubmit after fixes.
	rec = aiReq(t, router, http.MethodPost, "/videos/"+aiPending+"/ai-submit", aiOwnerToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("resubmit after reject: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestReviewAI_ApprovalFlagsParentSeries(t *testing.T) {
	router, memStore := newVideosRouter()
	memStore.Videos[aiEpisode1].AiReviewStatus = "pending"
	memStore.Series[aiSeriesID].AiOriginal = false
	rec := aiReq(t, router, http.MethodPost, "/videos/"+aiEpisode1+"/ai-review", "admin-token",
		map[string]string{"decision": "approved"})
	if rec.Code != http.StatusOK {
		t.Fatalf("approve: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !memStore.Series[aiSeriesID].AiOriginal {
		t.Fatal("approving an episode must flag its series as an AI Original")
	}
}

func TestReviewAI_RejectClearsFlagWhenLastApprovedLeaves(t *testing.T) {
	router, memStore := newVideosRouter()
	memStore.Videos[aiEpisode2].SeriesID = ""
	memStore.Videos[aiEpisode3].SeriesID = ""
	memStore.Videos[aiEpisode1].AiReviewStatus = "pending"
	rec := aiReq(t, router, http.MethodPost, "/videos/"+aiEpisode1+"/ai-review", "admin-token",
		map[string]string{"decision": "rejected", "note": "off-brief"})
	if rec.Code != http.StatusOK {
		t.Fatalf("reject: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Series[aiSeriesID].AiOriginal {
		t.Fatal("rejecting the last approved episode must clear the series flag")
	}
}

func TestUpdateVideo_UnflagClearsSeriesFlag(t *testing.T) {
	router, memStore := newVideosRouter()
	memStore.Videos[aiEpisode2].SeriesID = ""
	memStore.Videos[aiEpisode3].SeriesID = ""
	rec := aiReq(t, router, http.MethodPatch, "/videos/"+aiEpisode1, aiOwnerToken,
		map[string]interface{}{"ai_generated": false})
	if rec.Code != http.StatusOK {
		t.Fatalf("unflag: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Series[aiSeriesID].AiOriginal {
		t.Fatal("unflagging the last approved episode must clear the series flag")
	}
}

func TestDeleteVideo_ClearsSeriesFlag(t *testing.T) {
	router, memStore := newVideosRouter()
	memStore.Videos[aiEpisode2].SeriesID = ""
	memStore.Videos[aiEpisode3].SeriesID = ""
	rec := aiReq(t, router, http.MethodDelete, "/videos/"+aiEpisode1, aiOwnerToken, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Series[aiSeriesID].AiOriginal {
		t.Fatal("deleting the last approved episode must clear the series flag")
	}
}

func TestUpdateVideo_AIFlagTransitions(t *testing.T) {
	router, memStore := newVideosRouter()

	// Flagging a human video as AI clears any stale review trail.
	rec := aiReq(t, router, http.MethodPatch, "/videos/"+humanVideo, "creator-token",
		map[string]interface{}{"ai_generated": true, "ai_tool": "Test render"})
	if rec.Code != http.StatusOK {
		t.Fatalf("flag on: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	v := memStore.Videos[humanVideo]
	if !v.AiGenerated || v.AiTool != "Test render" || v.AiReviewStatus != "" {
		t.Fatalf("unexpected AI state after flag-on: %+v", v)
	}

	// Unflagging clears everything AI-related.
	rec = aiReq(t, router, http.MethodPatch, "/videos/"+humanVideo, "creator-token",
		map[string]interface{}{"ai_generated": false})
	if rec.Code != http.StatusOK {
		t.Fatalf("flag off: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	v = memStore.Videos[humanVideo]
	if v.AiGenerated || v.AiTool != "" || v.AiReviewStatus != "" || v.AiReviewNote != "" {
		t.Fatalf("AI trail not cleared: %+v", v)
	}
}

func TestGetVideo_SeriesContext(t *testing.T) {
	router, _ := newVideosRouter()
	rec := aiReq(t, router, http.MethodGet, "/videos/"+aiEpisode1, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	ctx, ok := body["series_context"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing series_context in %v", body)
	}
	if ctx["id"] != aiSeriesID || ctx["episode_total"].(float64) != 3 {
		t.Fatalf("bad series context: %v", ctx)
	}
	next, ok := ctx["next_episode"].(map[string]interface{})
	if !ok || next["video_id"] != aiEpisode2 {
		t.Fatalf("bad next episode: %v", ctx["next_episode"])
	}
}

func TestGetVideo_StandaloneHasNoSeriesContext(t *testing.T) {
	router, _ := newVideosRouter()
	rec := aiReq(t, router, http.MethodGet, "/videos/"+humanVideo, "", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, present := body["series_context"]; present {
		t.Fatalf("standalone video must omit series_context: %v", body["series_context"])
	}
}
