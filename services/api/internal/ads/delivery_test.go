package ads

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func adsRouter(s *store.MemoryStore) http.Handler {
	auth := identity.NewAuthMiddleware(s)
	return auth.Authenticate(NewHandler(s).Routes(auth))
}

func postAds(t *testing.T, router http.Handler, path, token string, payload interface{}) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func decide(t *testing.T, router http.Handler, userID string) (string, string) {
	t.Helper()
	rec := postAds(t, router, "/decision", "", map[string]interface{}{
		"slot_id": "watch_preroll", "user_id": userID, "category": "Technology", "platform": "WEB",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("decision failed: %d", rec.Code)
	}
	var resp struct {
		AdAvailable bool   `json:"ad_available"`
		DecisionID  string `json:"decision_id"`
		CampaignID  string `json:"campaign_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode decision: %v", err)
	}
	if !resp.AdAvailable {
		t.Fatalf("expected seeded eligible campaign, got %+v", resp)
	}
	return resp.DecisionID, resp.CampaignID
}

// P0-5: billing events must link to a real decision for the same campaign.
func TestAdEventDecisionLinkage(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	router := adsRouter(s)
	decisionID, campaignID := decide(t, router, "user-link-1")

	// Unknown decision id is rejected.
	rec := postAds(t, router, "/events", "", map[string]interface{}{
		"event_id": "evt-link-bad", "decision_id": "dec-nope",
		"campaign_id": campaignID, "event_type": "impression",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown decision, got %d", rec.Code)
	}

	// Decision bound to another campaign is rejected.
	s.Campaigns["camp-other"] = &store.AdCampaign{ID: "camp-other", AdvertiserID: "x", Name: "other", Status: "active"}
	rec = postAds(t, router, "/events", "", map[string]interface{}{
		"event_id": "evt-link-cross", "decision_id": decisionID,
		"campaign_id": "camp-other", "event_type": "click",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for cross-campaign event, got %d", rec.Code)
	}

	// Billing events require a decision id.
	rec = postAds(t, router, "/events", "", map[string]interface{}{
		"event_id": "evt-link-no-decision", "campaign_id": campaignID, "event_type": "impression",
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for impression without decision, got %d", rec.Code)
	}

	// Linked event is accepted and billed.
	before := s.Campaigns[campaignID].SpentCents
	rec = postAds(t, router, "/events", "", map[string]interface{}{
		"event_id": "evt-link-ok", "decision_id": decisionID,
		"campaign_id": campaignID, "event_type": "impression",
	})
	if rec.Code != http.StatusAccepted && rec.Code != http.StatusOK {
		t.Fatalf("expected accept for linked event, got %d: %s", rec.Code, rec.Body.String())
	}
	if s.Campaigns[campaignID].SpentCents <= before {
		t.Fatal("linked impression must accrue spend")
	}
}

// P0-5: preview can force-assign a visible campaign for verification.
func TestPreviewForcedAssignment(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	router := adsRouter(s)

	rec := postAds(t, router, "/preview", "advertiser-token", map[string]interface{}{
		"campaign_id": "camp-cloud-001", "slot_id": "watch_preroll",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for forced preview, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		PreviewMode bool `json:"preview_mode"`
		Creative    struct {
			CampaignID string `json:"campaign_id"`
		} `json:"creative"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if !body.PreviewMode || body.Creative.CampaignID != "camp-cloud-001" {
		t.Fatalf("forced preview must return the requested campaign creative: %+v", body)
	}

	// Unknown campaign id is a clean 404, not a silent fallback.
	rec = postAds(t, router, "/preview", "advertiser-token", map[string]interface{}{
		"campaign_id": "camp-nope",
	})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown preview campaign, got %d", rec.Code)
	}
}

// P0-5: frequency caps exclude over-served campaigns at decision time.
func TestFrequencyCapExcludesCampaign(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	// Single eligible campaign so the cap is observable.
	for id := range s.Campaigns {
		if id != "camp-cloud-001" {
			delete(s.Campaigns, id)
		}
	}
	router := adsRouter(s)

	var lastReason string
	served := 0
	for i := 0; i < hourlyCapPerUserCampaign+3; i++ {
		rec := postAds(t, router, "/decision", "", map[string]interface{}{
			"slot_id": "watch_preroll", "user_id": "user-cap-1", "category": "Technology", "platform": "WEB",
		})
		var resp struct {
			AdAvailable bool   `json:"ad_available"`
			Reason      string `json:"reason"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode decision: %v", err)
		}
		if resp.AdAvailable {
			served++
		} else {
			lastReason = resp.Reason
		}
	}
	if served != hourlyCapPerUserCampaign {
		t.Fatalf("expected %d served before cap, got %d", hourlyCapPerUserCampaign, served)
	}
	if lastReason != "FREQUENCY_CAPPED" {
		t.Fatalf("expected FREQUENCY_CAPPED after cap, got %q", lastReason)
	}
}

// P26: billing events stay valid across API restarts once decisions persist.
// Simulates a restart by dropping the in-memory decision map; the event must
// still validate against the persisted decision instead of 400 UNKNOWN_DECISION.
func TestAdEventSurvivesRestart(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL unset; skipping decision persistence test")
	}
	pg, err := store.NewPostgresStore(dsn)
	if err != nil {
		t.Fatalf("connect test postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Close() })
	s := store.NewMemoryStoreWithDB(pg)
	router := adsRouter(s)
	decisionID, campaignID := decide(t, router, "user-restart-1")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := pg.DB.ExecContext(ctx, "DELETE FROM ad_delivery_decisions WHERE decision_id = $1", decisionID); err != nil {
			t.Errorf("cleanup decision %s: %v", decisionID, err)
		}
	})

	// Simulate an API restart: memory is reseeded, decisions are gone.
	s.AdDecisions = map[string]*store.AdDeliveryDecision{}

	rec := postAds(t, router, "/events", "", map[string]interface{}{
		"event_id": "evt-restart-1", "decision_id": decisionID,
		"campaign_id": campaignID, "event_type": "impression",
	})
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for persisted decision after restart, got %d (%s)", rec.Code, rec.Body.String())
	}
}
