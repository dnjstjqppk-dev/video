package studio

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

const realtimeCreatorToken = "creator-token"

func realtimeAuth(router http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("Authorization", "Bearer "+realtimeCreatorToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestGetRealtime_Returns48Buckets(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	ch := s.ChannelForUser("22222222-2222-2222-2222-222222222222")
	if ch == nil {
		t.Fatal("creator seed channel missing")
	}
	now := time.Now().UTC().Truncate(time.Hour)
	s.RecordHourlyView(ch.ID, now.Add(-3*time.Hour), 7)
	s.RecordHourlyView(ch.ID, now, 4)

	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	rec := realtimeAuth(router, "/analytics/realtime")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var out struct {
		ChannelID string `json:"channel_id"`
		Hours     []struct {
			Hour  string `json:"hour"`
			Views int64  `json:"views"`
		} `json:"hours"`
		Total int64 `json:"total_48h"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode realtime: %v", err)
	}
	if len(out.Hours) != 48 {
		t.Fatalf("expected 48 buckets, got %d", len(out.Hours))
	}
	if out.Total != 11 {
		t.Fatalf("expected total 11, got %d", out.Total)
	}
	if out.Hours[47].Views != 4 || out.Hours[44].Views != 7 {
		t.Fatal("buckets landed in the wrong hours")
	}

	anon := httptest.NewRequest(http.MethodGet, "/analytics/realtime", nil)
	anonRec := httptest.NewRecorder()
	router.ServeHTTP(anonRec, anon)
	if anonRec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous realtime must be 401, got %d", anonRec.Code)
	}
}
