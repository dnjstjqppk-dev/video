package videos

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

const (
	testCreatorIDP = "22222222-2222-2222-2222-222222222222"
	testViewerIDP  = "33333333-3333-3333-3333-333333333333"
	testChannelP   = "ch-playback-test"
)

func seedVisibilityStore() *store.MemoryStore {
	s := store.NewMemoryStore()
	s.Channels[testChannelP] = &store.Channel{ID: testChannelP, OwnerID: testCreatorIDP, Handle: "@playback-test", Name: "Playback Test"}
	mk := func(id, visibility string) {
		s.Videos[id] = &store.Video{
			ID: id, ChannelID: testChannelP, Title: "Visibility " + id,
			Description: "visibility fixture", Visibility: visibility, Status: "READY",
			DurationSeconds: 60, HlsMasterURL: "https://media.test/hls/" + id + "/master.m3u8",
		}
	}
	mk("vid-pub", "public")
	mk("vid-unlisted", "unlisted")
	mk("vid-private", "private")
	mk("vid-members", "members_only")
	mk("vid-drm", "public")
	drm := s.Videos["vid-drm"]
	drm.DrmProtected = true
	drm.DrmKeyID = "00112233445566778899aabbccddeeff"
	drm.DrmKeyHex = "ffeeddccbbaa99887766554433221100"
	return s
}

func playbackRouter(s *store.MemoryStore) http.Handler {
	auth := identity.NewAuthMiddleware(s)
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/videos", NewHandler(s).Routes(auth))
	return r
}

func getPlayback(t *testing.T, router http.Handler, videoID, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/videos/"+videoID+"/playback", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// P0-4: public playback issues a verifiable signed token, not a static prefix.
func TestPlaybackTokenSignedAndVerifiable(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := playbackRouter(seedVisibilityStore())

	rec := getPlayback(t, router, "vid-pub", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for public playback, got %d", rec.Code)
	}
	var body struct {
		PlaybackToken string `json:"playback_token"`
		HasAccess     bool   `json:"has_access"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode playback: %v", err)
	}
	if body.PlaybackToken == "pb_token_vid-pub" || !strings.HasPrefix(body.PlaybackToken, "v1.") {
		t.Fatalf("expected signed v1 token, got %q", body.PlaybackToken)
	}
	if !body.HasAccess {
		t.Fatal("expected has_access true for public video")
	}
	if err := VerifyPlaybackToken(body.PlaybackToken, "vid-pub"); err != nil {
		t.Fatalf("issued token must verify: %v", err)
	}
	if err := VerifyPlaybackToken(body.PlaybackToken, "vid-other"); err == nil {
		t.Fatal("token must be bound to its video id")
	}
}

func TestPlaybackDrmKeysDelivered(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := playbackRouter(seedVisibilityStore())

	rec := getPlayback(t, router, "vid-drm", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for drm playback, got %d", rec.Code)
	}
	var body struct {
		DrmProtected bool `json:"drm_protected"`
		Drm          *struct {
			KeyID string `json:"key_id"`
			Key   string `json:"key"`
		} `json:"drm"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode playback: %v", err)
	}
	if !body.DrmProtected || body.Drm == nil || body.Drm.KeyID == "" || body.Drm.Key == "" {
		t.Fatalf("expected drm keys delivered: %+v", body)
	}

	rec = getPlayback(t, router, "vid-pub", "")
	var plain struct {
		DrmProtected bool      `json:"drm_protected"`
		Drm          *struct{} `json:"drm"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &plain); err != nil {
		t.Fatalf("failed to decode playback: %v", err)
	}
	if plain.DrmProtected || plain.Drm != nil {
		t.Fatalf("plain video must not carry drm: %+v", plain)
	}
}

// P0-4: private videos are hidden from everyone except the owner and admins.
func TestPlaybackPrivateHidden(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := playbackRouter(seedVisibilityStore())

	for _, token := range []string{"", "viewer-token"} {
		rec := getPlayback(t, router, "vid-private", token)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("token %q: expected 404 for private playback, got %d", token, rec.Code)
		}
		req := httptest.NewRequest(http.MethodGet, "/videos/vid-private", nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec2 := httptest.NewRecorder()
		router.ServeHTTP(rec2, req)
		if rec2.Code != http.StatusNotFound {
			t.Fatalf("token %q: expected 404 for private metadata, got %d", token, rec2.Code)
		}
	}

	rec := getPlayback(t, router, "vid-private", "creator-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for owner private playback, got %d", rec.Code)
	}
	rec = getPlayback(t, router, "vid-private", "admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin private playback, got %d", rec.Code)
	}
}

// P0-4: members-only playback requires entitlement; metadata stays public.
func TestPlaybackMembersOnlyGated(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := playbackRouter(seedVisibilityStore())

	rec := getPlayback(t, router, "vid-members", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 anonymous members-only, got %d", rec.Code)
	}
	rec = getPlayback(t, router, "vid-members", "viewer-token")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 free-tier members-only, got %d", rec.Code)
	}
	// Creator seed carries a premium tier.
	rec = getPlayback(t, router, "vid-members", "creator-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 premium members-only, got %d", rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/videos/vid-members", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 members-only metadata, got %d", rec2.Code)
	}
}

// P0-4: forged and expired tokens are rejected.
func TestPlaybackTokenForgeryRejected(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("PLAYBACK_SIGNING_SECRET", "test-playback-secret-0123456789")
	if err := VerifyPlaybackToken("v1.vid-pub.9999999999.deadbeef", "vid-pub"); err == nil {
		t.Fatal("forged signature must be rejected")
	}
	expired, err := SignPlaybackToken("vid-pub", -time.Hour)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	if err := VerifyPlaybackToken(expired, "vid-pub"); err == nil {
		t.Fatal("expired token must be rejected")
	}
	if err := VerifyPlaybackToken("not-a-token", "vid-pub"); err == nil {
		t.Fatal("malformed token must be rejected")
	}
}

// P1-14: one viewer counts once per video; access itself is unaffected.
func TestPlaybackCountsOncePerViewer(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := seedVisibilityStore()
	router := playbackRouter(s)
	before := s.Videos["vid-pub"].ViewCount

	if rec := getPlayback(t, router, "vid-pub", "creator-token"); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec := getPlayback(t, router, "vid-pub", "creator-token"); rec.Code != http.StatusOK {
		t.Fatalf("repeat playback must still grant access, got %d", rec.Code)
	}
	if rec := getPlayback(t, router, "vid-pub", "viewer-token"); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := s.Videos["vid-pub"].ViewCount - before; got != 2 {
		t.Fatalf("expected 2 counted views (creator once + viewer once), got %d", got)
	}
}

// P0-4: production refuses to issue tokens without a configured secret.
func TestPlaybackSigningRefusesWithoutSecretInProduction(t *testing.T) {
	t.Setenv("GVP_ENV", "production")
	t.Setenv("PLAYBACK_SIGNING_SECRET", "")
	t.Setenv("JWT_SECRET", "")
	if _, err := SignPlaybackToken("vid-pub", time.Hour); err == nil {
		t.Fatal("production signing without secret must fail")
	}
}
