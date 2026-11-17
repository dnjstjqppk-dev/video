package partners

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

func isoRouter(s *store.MemoryStore) http.Handler {
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/partners", NewHandler(s).Routes(auth))
	return r
}

func isoReq(t *testing.T, router http.Handler, method, path, token string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var body *bytes.Reader
	if payload != nil {
		raw, _ := json.Marshal(payload)
		body = bytes.NewReader(raw)
	} else {
		body = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// P1-18: partner surfaces require a partner role, not just any login.
func TestPartnerSurfaceRoleGated(t *testing.T) {
	s := store.NewMemoryStore()
	router := isoRouter(s)
	for _, path := range []string{"/partners/me", "/partners/channels", "/partners/contracts"} {
		rec := isoReq(t, router, http.MethodGet, path, "viewer-token", nil)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("GET %s viewer: expected 403, got %d", path, rec.Code)
		}
		rec = isoReq(t, router, http.MethodGet, path, "creator-token", nil)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("GET %s creator: expected 403, got %d", path, rec.Code)
		}
	}
}

// P1-18: partner overview reports the caller's own numbers, never globals.
func TestPartnerMeScoped(t *testing.T) {
	s := store.NewMemoryStore()
	s.Users["55555555-5555-5555-5555-555555555555"] = &store.User{
		ID: "55555555-5555-5555-5555-555555555555", Email: "adv2@test.local",
		Role: "advertiser", SubscriptionTier: "enterprise", IsActive: true,
	}
	router := isoRouter(s)

	rec := isoReq(t, router, http.MethodGet, "/partners/me", "advertiser-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		CampaignCount       int `json:"campaignCount"`
		ActiveChannelsCount int `json:"activeChannelsCount"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Seed store holds 2 campaigns for this advertiser and no owned channels.
	if body.CampaignCount != 2 || body.ActiveChannelsCount != 0 {
		t.Fatalf("overview must scope to caller, got %+v", body)
	}
}

// P1-18: contract revocation is role-gated, ownership-checked, and audited.
func TestRevokeContractGuarded(t *testing.T) {
	s := store.NewMemoryStore()
	s.Contracts["ctr-test-1"] = &store.PartnerContract{
		ID: "ctr-test-1", ChannelID: "ch-owned", ChannelName: "Owned",
		PartnerType: "MCN", Status: "active",
	}
	s.Channels["ch-owned"] = &store.Channel{ID: "ch-owned", OwnerID: "22222222-2222-2222-2222-222222222222", Handle: "@owned", Name: "Owned"}
	router := isoRouter(s)

	rec := isoReq(t, router, http.MethodPatch, "/partners/contracts/ctr-test-1", "viewer-token", map[string]string{"reason": "x"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 viewer revoke, got %d", rec.Code)
	}
	before := len(s.AuditLogs)
	rec = isoReq(t, router, http.MethodPatch, "/partners/contracts/ctr-test-1", "admin-token", map[string]string{"reason": "breach"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 admin revoke, got %d: %s", rec.Code, rec.Body.String())
	}
	if s.Contracts["ctr-test-1"].Status != "revoked" {
		t.Fatal("contract must be revoked")
	}
	if len(s.AuditLogs) != before+1 {
		t.Fatalf("revoke must audit, got %d -> %d", before, len(s.AuditLogs))
	}

	// MCN managers operate across creators' channels (role-trusted).
	s.Users["66666666-6666-6666-6666-666666666666"] = &store.User{
		ID: "66666666-6666-6666-6666-666666666666", Email: "mcn@test.local",
		Role: "partner_manager", SubscriptionTier: "enterprise", IsActive: true,
	}
	s.Contracts["ctr-test-2"] = &store.PartnerContract{
		ID: "ctr-test-2", ChannelID: "ch-owned", ChannelName: "Owned",
		PartnerType: "MCN", Status: "active",
	}
	rec = isoReq(t, router, http.MethodPatch, "/partners/contracts/ctr-test-2", "66666666-6666-6666-6666-666666666666", map[string]string{"reason": "breach"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 manager revoke, got %d: %s", rec.Code, rec.Body.String())
	}
}
