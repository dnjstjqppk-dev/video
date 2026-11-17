package ads

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

func testRouter(s *store.MemoryStore) http.Handler {
	auth := identity.NewAuthMiddleware(s)
	return auth.Authenticate(NewHandler(s).Routes(auth))
}

func doAds(t *testing.T, router http.Handler, method, path, token string, payload interface{}) *httptest.ResponseRecorder {
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

func campaignPayload(name string) map[string]interface{} {
	return map[string]interface{}{
		"name":             name,
		"format":           "preroll",
		"totalBudget":      1000000,
		"dailyBudget":      100000,
		"cpmBidCents":      5000,
		"categories":       []string{"Technology"},
		"creativeMediaUrl": "https://media.test/ads/creative.mp4",
		"headline":         "Test creative",
		"destinationUrl":   "https://advertiser.test/landing",
		"callToAction":     "Learn More",
	}
}

func createCampaign(t *testing.T, router http.Handler, token, name string) map[string]interface{} {
	t.Helper()
	rec := doAds(t, router, http.MethodPost, "/campaigns", token, campaignPayload(name))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for campaign creation, got %d: %s", rec.Code, rec.Body.String())
	}
	var camp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &camp); err != nil {
		t.Fatalf("failed to decode campaign: %v", err)
	}
	return camp
}

// P0-8: anonymous campaign/report/review access must be refused.
func TestAdSurfaceAnonymousDenied(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := testRouter(store.NewMemoryStore())

	anon := []struct{ method, path string }{
		{http.MethodGet, "/campaigns"},
		{http.MethodPost, "/campaigns"},
		{http.MethodGet, "/campaigns/camp-1"},
		{http.MethodPatch, "/campaigns/camp-1/status"},
		{http.MethodGet, "/reports"},
		{http.MethodGet, "/advertiser/reports"},
		{http.MethodPost, "/preview"},
		{http.MethodGet, "/admin/diagnostics"},
		{http.MethodGet, "/admin/creatives"},
		{http.MethodPost, "/admin/creatives/cr-1/approve"},
		{http.MethodPost, "/admin/creatives/cr-1/reject"},
	}
	for _, c := range anon {
		rec := doAds(t, router, c.method, c.path, "", campaignPayload("anon"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s: expected 401 anonymous, got %d", c.method, c.path, rec.Code)
		}
	}
}

// P0-8: non-advertiser roles must be refused on campaign management.
func TestAdSurfaceRoleDenied(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := testRouter(store.NewMemoryStore())

	for _, token := range []string{"viewer-token", "creator-token"} {
		rec := doAds(t, router, http.MethodGet, "/campaigns", token, nil)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("token %s: expected 403 on /campaigns, got %d", token, rec.Code)
		}
		rec = doAds(t, router, http.MethodPost, "/campaigns", token, campaignPayload("x"))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("token %s: expected 403 on POST /campaigns, got %d", token, rec.Code)
		}
	}
}

// P0-8: creation must enter DRAFT with pending creative, never auto active/approved.
func TestCreateCampaignEntersDraft(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	router := testRouter(s)

	camp := createCampaign(t, router, "advertiser-token", "Draft state machine check")
	if camp["status"] != "draft" {
		t.Fatalf("expected status draft, got %v", camp["status"])
	}
	if camp["advertiser_id"] != "44444444-4444-4444-4444-444444444444" {
		t.Fatalf("campaign must be scoped to the calling advertiser, got %v", camp["advertiser_id"])
	}
	creatives, _ := camp["creatives"].([]interface{})
	if len(creatives) != 1 {
		t.Fatalf("expected 1 creative, got %v", camp["creatives"])
	}
	stored := s.AdCreatives[creatives[0].(string)]
	if stored == nil || stored.ReviewStatus != "pending" {
		t.Fatalf("expected creative review pending, got %+v", stored)
	}
}

// P0-8: advertisers see only their own campaigns and reports.
func TestAdvertiserTenantIsolation(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	s.Users["55555555-5555-5555-5555-555555555555"] = &store.User{
		ID: "55555555-5555-5555-5555-555555555555", Email: "adv2@test.local",
		Role: "advertiser", SubscriptionTier: "enterprise", IsActive: true,
	}
	router := testRouter(s)

	own := createCampaign(t, router, "advertiser-token", "own campaign")
	other := createCampaign(t, router, "55555555-5555-5555-5555-555555555555", "other campaign")

	rec := doAds(t, router, http.MethodGet, "/campaigns", "advertiser-token", nil)
	var list []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("failed to decode list: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("advertiser must list own campaigns, got none")
	}
	for _, item := range list {
		if item["advertiser_id"] != "44444444-4444-4444-4444-444444444444" {
			t.Fatalf("advertiser list leaked another tenant: %v", item["id"])
		}
		if item["id"] == other["id"] {
			t.Fatalf("advertiser list includes another tenant's campaign %v", other["id"])
		}
	}
	found := false
	for _, item := range list {
		if item["id"] == own["id"] {
			found = true
		}
	}
	if !found {
		t.Fatalf("advertiser list missing own campaign %v", own["id"])
	}

	rec = doAds(t, router, http.MethodGet, "/campaigns/"+other["id"].(string), "advertiser-token", nil)
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Fatalf("cross-advertiser read must be refused, got %d", rec.Code)
	}

	rec = doAds(t, router, http.MethodPatch, "/campaigns/"+other["id"].(string)+"/status", "advertiser-token",
		map[string]interface{}{"status": "paused"})
	if rec.Code != http.StatusForbidden && rec.Code != http.StatusNotFound {
		t.Fatalf("cross-advertiser mutation must be refused, got %d", rec.Code)
	}

	rec = doAds(t, router, http.MethodGet, "/reports", "advertiser-token", nil)
	var reports map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &reports); err != nil {
		t.Fatalf("failed to decode reports: %v", err)
	}
	items, _ := reports["campaigns"].([]interface{})
	if len(items) == 0 {
		t.Fatal("advertiser reports must include own campaigns, got none")
	}
	seenOwn := false
	for _, raw := range items {
		item, _ := raw.(map[string]interface{})
		if item["id"] == other["id"] {
			t.Fatalf("advertiser reports leaked another tenant's campaign %v", other["id"])
		}
		if item["id"] == own["id"] {
			seenOwn = true
		}
	}
	if !seenOwn {
		t.Fatalf("advertiser reports missing own campaign %v", own["id"])
	}
}

// P0-8: only platform admins may approve/reject creatives, with an audit record.
func TestCreativeReviewAdminOnlyWithAudit(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	router := testRouter(s)

	camp := createCampaign(t, router, "advertiser-token", "review flow")
	creativeID := camp["creatives"].([]interface{})[0].(string)

	rec := doAds(t, router, http.MethodPost, "/admin/creatives/"+creativeID+"/approve", "advertiser-token", nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("advertiser self-approval must be refused, got %d", rec.Code)
	}

	before := len(s.AuditLogs)
	rec = doAds(t, router, http.MethodPost, "/admin/creatives/"+creativeID+"/approve", "admin-token", nil)
	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 without MFA step-up, got %d", rec.Code)
	}
	admin := s.Users["11111111-1111-1111-1111-111111111111"]
	admin.MFAEnabled = true
	admin.TOTPSecret = "JBSWY3DPEHPK3PXP"
	code, err := security.ComputeTOTP(admin.TOTPSecret, time.Now())
	if err != nil {
		t.Fatalf("compute totp: %v", err)
	}
	rec = doAds(t, router, http.MethodPost, "/admin/creatives/"+creativeID+"/approve", "admin-token", map[string]string{"totpCode": code})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin approval, got %d: %s", rec.Code, rec.Body.String())
	}
	if s.AdCreatives[creativeID].ReviewStatus != "approved" {
		t.Fatalf("creative must be approved, got %s", s.AdCreatives[creativeID].ReviewStatus)
	}
	if len(s.AuditLogs) != before+1 {
		t.Fatalf("approval must append an audit log, got %d -> %d", before, len(s.AuditLogs))
	}
}

// P0-8: status transitions follow the review state machine.
func TestCampaignStatusTransitionsGuarded(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	router := testRouter(s)

	camp := createCampaign(t, router, "advertiser-token", "transitions")
	id := camp["id"].(string)

	rec := doAds(t, router, http.MethodPatch, "/campaigns/"+id+"/status", "advertiser-token",
		map[string]interface{}{"status": "active"})
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusForbidden {
		t.Fatalf("draft->active jump must be refused, got %d", rec.Code)
	}

	rec = doAds(t, router, http.MethodPatch, "/campaigns/"+id+"/status", "advertiser-token",
		map[string]interface{}{"status": "nonsense"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown status must be refused, got %d", rec.Code)
	}

	rec = doAds(t, router, http.MethodPatch, "/campaigns/"+id+"/status", "advertiser-token",
		map[string]interface{}{"status": "pending_review"})
	if rec.Code != http.StatusOK {
		t.Fatalf("draft->pending_review submit must succeed, got %d: %s", rec.Code, rec.Body.String())
	}
	if s.Campaigns[id].Status != "pending_review" {
		t.Fatalf("expected pending_review, got %s", s.Campaigns[id].Status)
	}
}
