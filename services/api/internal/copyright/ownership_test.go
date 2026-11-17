package copyright

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

func ownRouter(s *store.MemoryStore) http.Handler {
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/rights", NewHandler(s).Routes(auth))
	return r
}

func ownReq(t *testing.T, router http.Handler, method, path, token string, payload interface{}) *httptest.ResponseRecorder {
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

// P1-18: claim surfaces require authentication (no public listing).
func TestClaimsRequireAuth(t *testing.T) {
	s := store.NewMemoryStore()
	router := ownRouter(s)
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/rights/claims"},
		{http.MethodGet, "/rights/claims/claim-1"},
		{http.MethodPost, "/rights/claims"},
	} {
		rec := ownReq(t, router, tc.method, tc.path, "", map[string]string{"videoId": "v", "assetId": "a", "action": "monetize"})
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: expected 401 anonymous, got %d", tc.method, tc.path, rec.Code)
		}
	}
}

// P1-18: reporters see own claims; strangers see none; admins see all.
func TestClaimsScopedToParties(t *testing.T) {
	s := store.NewMemoryStore()
	s.Users["55555555-5555-5555-5555-555555555555"] = &store.User{
		ID: "55555555-5555-5555-5555-555555555555", Email: "rh@test.local",
		Role: "rights_holder", SubscriptionTier: "enterprise", IsActive: true,
	}
	router := ownRouter(s)

	rec := ownReq(t, router, http.MethodPost, "/rights/claims", "55555555-5555-5555-5555-555555555555", map[string]string{
		"videoId": "00000000-0000-0000-0000-000000000001", "assetId": "ast-1", "action": "monetize",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 file claim, got %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Stranger (viewer) sees neither the claim in list nor by id.
	rec = ownReq(t, router, http.MethodGet, "/rights/claims", "viewer-token", nil)
	var list []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	for _, c := range list {
		if c["id"] == created.ID {
			t.Fatal("stranger must not list another reporter's claim")
		}
	}
	rec = ownReq(t, router, http.MethodGet, "/rights/claims/"+created.ID, "viewer-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 stranger read, got %d", rec.Code)
	}

	// Reporter reads own claim; admin reads everything.
	rec = ownReq(t, router, http.MethodGet, "/rights/claims/"+created.ID, "55555555-5555-5555-5555-555555555555", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 reporter read, got %d", rec.Code)
	}
	rec = ownReq(t, router, http.MethodGet, "/rights/claims/"+created.ID, "admin-token", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 admin read, got %d", rec.Code)
	}
}

// P1-18: disputes come from the video owner; resolutions from the rights
// holder or operators; both are audited.
func TestClaimDisputeResolveOwnership(t *testing.T) {
	s := store.NewMemoryStore()
	s.Users["55555555-5555-5555-5555-555555555555"] = &store.User{
		ID: "55555555-5555-5555-5555-555555555555", Email: "rh@test.local",
		Role: "rights_holder", SubscriptionTier: "enterprise", IsActive: true,
	}
	router := ownRouter(s)

	rec := ownReq(t, router, http.MethodPost, "/rights/claims", "55555555-5555-5555-5555-555555555555", map[string]string{
		"videoId": "00000000-0000-0000-0000-000000000001", "assetId": "ast-1", "action": "monetize",
	})
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	// Stranger cannot dispute.
	rec = ownReq(t, router, http.MethodPost, "/rights/claims/"+created.ID+"/dispute", "viewer-token", map[string]string{"reason": "fair_use"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 stranger dispute, got %d", rec.Code)
	}
	before := len(s.AuditLogs)
	// Channel owner (creator) can dispute.
	rec = ownReq(t, router, http.MethodPost, "/rights/claims/"+created.ID+"/dispute", "creator-token", map[string]string{"reason": "fair_use"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 owner dispute, got %d: %s", rec.Code, rec.Body.String())
	}

	// Stranger cannot resolve.
	rec = ownReq(t, router, http.MethodPost, "/rights/claims/"+created.ID+"/resolve", "viewer-token", map[string]string{"resolution": "release"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 stranger resolve, got %d", rec.Code)
	}
	// Reporter resolves own claim with an audit record.
	rec = ownReq(t, router, http.MethodPost, "/rights/claims/"+created.ID+"/resolve", "55555555-5555-5555-5555-555555555555", map[string]string{"resolution": "uphold"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 reporter resolve, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(s.AuditLogs) < before+2 {
		t.Fatalf("dispute+resolve must append audit logs, got %d -> %d", before, len(s.AuditLogs))
	}
}

// P1-18: DMCA takedowns require a rights role; counters come from the owner.
func TestDMCARolesEnforced(t *testing.T) {
	s := store.NewMemoryStore()
	router := ownRouter(s)

	dmca := map[string]interface{}{
		"video_id": "00000000-0000-0000-0000-000000000001", "claimant_name": "X",
		"claimant_email": "x@test.local", "copyright_work_title": "W",
		"signature_legal": "X", "good_faith_belief": true,
	}
	rec := ownReq(t, router, http.MethodPost, "/rights/claims/dmca-takedown", "viewer-token", dmca)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 viewer takedown, got %d", rec.Code)
	}
	rec = ownReq(t, router, http.MethodPost, "/rights/claims/dmca-takedown", "admin-token", dmca)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 admin takedown, got %d: %s", rec.Code, rec.Body.String())
	}
}
