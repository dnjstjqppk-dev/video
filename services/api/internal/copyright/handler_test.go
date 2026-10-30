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

	// 1. List claims
	req, rec := copyReq(http.MethodGet, "/copyright/claims", "", nil)
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

	// 3. Rights holder releases claim
	req, rec = copyReq(http.MethodPost, "/copyright/claims/"+claimID+"/resolve", "advertiser-token", map[string]string{
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

func TestDMCATakedownAndCounterNoticeWorkflow(t *testing.T) {
	srv, memStore := newCopyrightTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	creatorUser := memStore.Users["22222222-2222-2222-2222-222222222222"]
	initialStrikes := creatorUser.CopyrightStrikes

	// 1. Submit DMCA Takedown
	req, rec := copyReq(http.MethodPost, "/copyright/claims/dmca-takedown", "advertiser-token", map[string]interface{}{
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
