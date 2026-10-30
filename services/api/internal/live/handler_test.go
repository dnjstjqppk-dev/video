package live

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func newLiveTestServer() (*httptest.Server, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	authMiddleware := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	liveH := NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(authMiddleware.Authenticate)
	r.Mount("/live-streams", liveH.Routes(authMiddleware))

	srv := httptest.NewServer(r)
	return srv, memStore
}

func liveReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	return req, rec
}

func TestLiveStreamLifecycle(t *testing.T) {
	srv, _ := newLiveTestServer()
	defer srv.Close()

	// 1. Create stream
	req, rec := liveReq(http.MethodPost, "/live-streams", "creator-token", map[string]string{
		"title":       "4K 60fps Live Tech Summit",
		"description": "Distributed HLS Transcoding",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d: %s", rec.Code, rec.Body.String())
	}

	var stream store.LiveStream
	json.Unmarshal(rec.Body.Bytes(), &stream)
	if stream.ID == "" || stream.Status != "idle" || stream.StreamKey == "" {
		t.Fatalf("invalid stream created: %+v", stream)
	}

	// 2. Start stream -> live
	req, rec = liveReq(http.MethodPost, "/live-streams/"+stream.ID+"/start", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for start stream, got %d", rec.Code)
	}
	json.Unmarshal(rec.Body.Bytes(), &stream)
	if stream.Status != "live" {
		t.Fatalf("expected status live, got %s", stream.Status)
	}

	// 3. Health check
	req, rec = liveReq(http.MethodGet, "/live-streams/"+stream.ID+"/health", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for stream health, got %d", rec.Code)
	}

	// 4. Stop stream -> ended
	req, rec = liveReq(http.MethodPost, "/live-streams/"+stream.ID+"/stop", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for stop stream, got %d", rec.Code)
	}
	json.Unmarshal(rec.Body.Bytes(), &stream)
	if stream.Status != "ended" {
		t.Fatalf("expected status ended, got %s", stream.Status)
	}
}

func TestSuperChat_TiersAndPinDuration(t *testing.T) {
	srv, _ := newLiveTestServer()
	defer srv.Close()

	streamID := "live_main_01"

	// Tier 1 ($2 / 200 cents) -> blue, 60s pin
	req, rec := liveReq(http.MethodPost, "/live-streams/"+streamID+"/superchat", "viewer-token", map[string]interface{}{
		"amount_cents": 200,
		"currency":     "USD",
		"message":      "Awesome stream!",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for superchat, got %d: %s", rec.Code, rec.Body.String())
	}
	var sc1 store.SuperChat
	json.Unmarshal(rec.Body.Bytes(), &sc1)
	if sc1.TierColor != "blue" || sc1.PinDurationSecs != 60 {
		t.Fatalf("expected blue/60s for $2 superchat, got color=%s pin=%d", sc1.TierColor, sc1.PinDurationSecs)
	}

	// Tier 6 ($100 / 10000 cents) -> red, 3600s pin
	req, rec = liveReq(http.MethodPost, "/live-streams/"+streamID+"/superchat", "viewer-token", map[string]interface{}{
		"amount_cents": 10000,
		"currency":     "USD",
		"message":      "Huge supporter from Korea!",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for $100 superchat, got %d", rec.Code)
	}
	var sc2 store.SuperChat
	json.Unmarshal(rec.Body.Bytes(), &sc2)
	if sc2.TierColor != "red" || sc2.PinDurationSecs != 3600 {
		t.Fatalf("expected red/3600s for $100 superchat, got color=%s pin=%d", sc2.TierColor, sc2.PinDurationSecs)
	}

	// Fetch active superchats list
	req, rec = liveReq(http.MethodGet, "/live-streams/"+streamID+"/superchats", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for get superchats, got %d", rec.Code)
	}
	var listResp struct {
		Items []*store.SuperChat `json:"items"`
		Total int                `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &listResp)
	if listResp.Total < 2 {
		t.Fatalf("expected at least 2 active superchats, got %d", listResp.Total)
	}
}

func TestLiveChatAndPollWorkflow(t *testing.T) {
	srv, _ := newLiveTestServer()
	defer srv.Close()

	streamID := "live_main_01"

	// 1. Send chat message
	req, rec := liveReq(http.MethodPost, "/live-streams/"+streamID+"/chat", "viewer-token", map[string]string{
		"message": "Hello from Seoul!",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for chat message, got %d", rec.Code)
	}

	// 2. Create Live Poll
	req, rec = liveReq(http.MethodPost, "/live-streams/"+streamID+"/polls", "creator-token", map[string]interface{}{
		"question": "Which codec should we benchmark next?",
		"options":  []string{"AV1 (SVT-AV1)", "VVC (H.266)", "HEVC (H.265)"},
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for poll create, got %d: %s", rec.Code, rec.Body.String())
	}
	var poll store.LivePoll
	json.Unmarshal(rec.Body.Bytes(), &poll)
	if len(poll.Options) != 3 {
		t.Fatalf("expected 3 poll options, got %d", len(poll.Options))
	}

	// 3. Vote on Poll (option index 0)
	req, rec = liveReq(http.MethodPost, "/live-streams/"+streamID+"/polls/vote", "viewer-token", map[string]int{
		"option_index": 0,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for poll vote, got %d", rec.Code)
	}
	var updatedPoll store.LivePoll
	json.Unmarshal(rec.Body.Bytes(), &updatedPoll)
	if updatedPoll.TotalVotes != 1 || updatedPoll.Options[0].Votes != 1 {
		t.Fatalf("vote count mismatch: total=%d, opt0=%d", updatedPoll.TotalVotes, updatedPoll.Options[0].Votes)
	}
}
