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

func newCopyrightTestServer() (*httptest.Server, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	copyH := NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/copyright", copyH.Routes(auth))

	return httptest.NewServer(r), memStore
}

func copyReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
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

func TestContentIDClaimsAndDisputeWorkflow(t *testing.T) {
	srv, _ := newCopyrightTestServer()
	defer srv.Close()

	// 1. List claims (P1-18: authenticated; anonymous is refused)
	req, rec := copyReq(http.MethodGet, "/copyright/claims", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 anonymous list, got %d", rec.Code)
	}
	req, rec = copyReq(http.MethodGet, "/copyright/claims", "admin-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for list claims, got %d", rec.Code)
	}

	var claims []*store.CopyrightClaim
	json.Unmarshal(rec.Body.Bytes(), &claims)
	if len(claims) < 1 {
		t.Fatalf("expected at least 1 seed claim, got %d", len(claims))
	}
	claimID := claims[0].ID

	// 2. Creator files dispute for fair use
	req, rec = copyReq(http.MethodPost, "/copyright/claims/"+claimID+"/dispute", "creator-token", map[string]string{
		"reason":      "fair_use",
		"explanation": "Critical commentary and technical review under 17 U.S.C. 107",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for dispute, got %d: %s", rec.Code, rec.Body.String())
	}
	var disputed store.CopyrightClaim
	json.Unmarshal(rec.Body.Bytes(), &disputed)
	if disputed.Status != "disputed" || disputed.DisputeReason != "fair_use" {
		t.Fatalf("invalid dispute state: status=%s reason=%s", disputed.Status, disputed.DisputeReason)
	}

	// 3. Operator releases claim (P1-18: reporter or operator only)
	req, rec = copyReq(http.MethodPost, "/copyright/claims/"+claimID+"/resolve", "admin-token", map[string]string{
		"resolution": "release",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for resolve, got %d", rec.Code)
	}
	var resolved store.CopyrightClaim
	json.Unmarshal(rec.Body.Bytes(), &resolved)
	if resolved.Status != "released" {
		t.Fatalf("expected claim status released, got %s", resolved.Status)
	}
}

func TestDisputeOwnershipEnforced(t *testing.T) {
	srv, _ := newCopyrightTestServer()
	defer srv.Close()

	// A viewer who does not own the claimed video must be rejected.
	req, rec := copyReq(http.MethodPost, "/copyright/claims/claim-1/dispute", "viewer-token", map[string]string{
		"reason": "fair_use",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-owner dispute, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDMCATakedownAndCounterNoticeWorkflow(t *testing.T) {
	srv, memStore := newCopyrightTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	creatorUser := memStore.Users["22222222-2222-2222-2222-222222222222"]
	initialStrikes := creatorUser.CopyrightStrikes

	// 1. Submit DMCA Takedown (P1-18: rights roles and operators only)
	req, rec := copyReq(http.MethodPost, "/copyright/claims/dmca-takedown", "admin-token", map[string]interface{}{
		"video_id":             videoID,
		"claimant_name":        "Acme Rights Management Corp",
		"claimant_email":       "dmca@acme-rights.com",
		"copyright_work_title": "Original Film Soundtrack Master",
		"signature_legal":      "John Doe, General Counsel",
		"good_faith_belief":    true,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for DMCA takedown, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Notice struct {
			ID string `json:"id"`
		} `json:"notice"`
		VideoBlocked bool `json:"video_blocked"`
		StrikeIssued bool `json:"strike_issued"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if !resp.VideoBlocked || !resp.StrikeIssued {
		t.Fatalf("expected video blocked and strike issued, got blocked=%v strike=%v", resp.VideoBlocked, resp.StrikeIssued)
	}
	if creatorUser.CopyrightStrikes != initialStrikes+1 {
		t.Fatalf("expected creator strikes to increase from %d to %d, got %d", initialStrikes, initialStrikes+1, creatorUser.CopyrightStrikes)
	}

	// 2. Creator submits counter-notice
	req, rec = copyReq(http.MethodPost, "/copyright/claims/dmca-counter-notice", "creator-token", map[string]interface{}{
		"notice_id":                    resp.Notice.ID,
		"legal_full_name":              "Jane Creator",
		"address":                      "123 Creator Way, Seoul, KR",
		"consent_federal_jurisdiction": true,
		"statement_perjury":            "I swear under penalty of perjury that this video was removed as a result of mistake.",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for counter-notice, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateClaim_RequiresRightsRole(t *testing.T) {
	srv, _ := newCopyrightTestServer()
	defer srv.Close()
	body := map[string]string{
		"videoId": "00000000-0000-0000-0000-000000000001",
		"assetId": "asset-x",
		"action":  "block",
	}
	for _, token := range []string{"creator-token", "viewer-token"} {
		req, rec := copyReq(http.MethodPost, "/copyright/claims", token, body)
		srv.Config.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for %s filing a claim, got %d", token, rec.Code)
		}
	}
	req, rec := copyReq(http.MethodPost, "/copyright/claims", "admin-token", body)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for operator filing a claim, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateClaim_ValidatesPayload(t *testing.T) {
	srv, _ := newCopyrightTestServer()
	defer srv.Close()
	req, rec := copyReq(http.MethodPost, "/copyright/claims", "admin-token", map[string]string{
		"videoId": "99999999-9999-9999-9999-999999999999",
		"action":  "block",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown video, got %d", rec.Code)
	}
	req, rec = copyReq(http.MethodPost, "/copyright/claims", "admin-token", map[string]string{
		"videoId": "00000000-0000-0000-0000-000000000001",
		"action":  "nuke",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad action, got %d", rec.Code)
	}
}

func TestUpdateClaimAction_AuthzAndUpdate(t *testing.T) {
	srv, memStore := newCopyrightTestServer()
	defer srv.Close()
	req, rec := copyReq(http.MethodPost, "/copyright/claims", "admin-token", map[string]string{
		"videoId": "00000000-0000-0000-0000-000000000001",
		"action":  "track",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup: expected 201, got %d", rec.Code)
	}
	var created store.CopyrightClaim
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created claim: %v", err)
	}
	// Claimed channel owner can read but not change policy.
	req, rec = copyReq(http.MethodPatch, "/copyright/claims/"+created.ID, "creator-token",
		map[string]string{"action": "block"})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-reporter policy change, got %d", rec.Code)
	}
	req, rec = copyReq(http.MethodPatch, "/copyright/claims/"+created.ID, "admin-token",
		map[string]string{"action": "explode"})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for bad action, got %d", rec.Code)
	}
	req, rec = copyReq(http.MethodPatch, "/copyright/claims/"+created.ID, "admin-token",
		map[string]string{"action": "block"})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for reporter policy change, got %d: %s", rec.Code, rec.Body.String())
	}
	if memStore.Claims[created.ID].ClaimAction != "block" {
		t.Fatalf("claim action = %q, want block", memStore.Claims[created.ID].ClaimAction)
	}
	if len(memStore.Claims) != 2 {
		t.Fatalf("policy change must update in place, claims = %d", len(memStore.Claims))
	}
}
