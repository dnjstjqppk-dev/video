package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/store"
)

func TestAPIHealthzAndFeeds(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /healthz, got %d", rec.Code)
	}

	// 1.1 Ready probe (TASK-086)
	req = httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /ready, got %d", rec.Code)
	}

	// 1.2 Prometheus metrics (TASK-087)
	req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !bytes.Contains(rec.Body.Bytes(), []byte("gvp_uptime_seconds")) {
		t.Fatalf("expected 200 and prometheus metrics for /metrics, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Home Feed
	req = httptest.NewRequest(http.MethodGet, "/v1/feed/home", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /v1/feed/home, got %d", rec.Code)
	}
	var feedResp struct {
		Total int `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &feedResp); err != nil || feedResp.Total < 1 {
		t.Fatalf("expected non-empty home feed, got total=%d err=%v", feedResp.Total, err)
	}

	// 3. Video Playback
	req = httptest.NewRequest(http.MethodGet, "/v1/videos/00000000-0000-0000-0000-000000000001/playback", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for video playback, got %d", rec.Code)
	}
}

func TestUploadSessionStudioAndPartnerClaims(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Create Upload Session
	uploadBody := []byte(`{"title":"4K HDR Test Upload","description":"Upload test","fileSizeBytes":25000000,"mimeType":"video/mp4","videoType":"standard"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/uploads", bytes.NewReader(uploadBody))
	req.Header.Set("Authorization", "Bearer creator-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for upload session creation, got %d: %s", rec.Code, rec.Body.String())
	}
	var upResp struct {
		UploadID   string `json:"uploadId"`
		VideoID    string `json:"videoId"`
		TotalParts int    `json:"totalParts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &upResp); err != nil || upResp.UploadID == "" || upResp.TotalParts != 3 {
		t.Fatalf("unexpected upload response: %+v, err=%v", upResp, err)
	}

	// 2. Studio Dashboard
	req = httptest.NewRequest(http.MethodGet, "/v1/studio/dashboard", nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for studio dashboard, got %d", rec.Code)
	}

	// 3. Partner Claims
	claimBody := []byte(`{"videoId":"00000000-0000-0000-0000-000000000001","assetId":"asset-99","action":"monetize"}`)
	req = httptest.NewRequest(http.MethodPost, "/v1/partners/rights/claims", bytes.NewReader(claimBody))
	req.Header.Set("Authorization", "Bearer creator-token")
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for partner claim creation, got %d", rec.Code)
	}
}

func TestAdminRBACEnforcement(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// Non-admin user must receive 403 Forbidden
	req := httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for non-admin, got %d", rec.Code)
	}

	// super_admin user must receive 200 OK
	req = httptest.NewRequest(http.MethodGet, "/v1/admin/dashboard", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for super_admin, got %d", rec.Code)
	}
}

func TestPaymentWebhookIdempotencyAndDoubleEntryLedger(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// Seed a previously created checkout; provider HTTP behavior is tested separately.
	t.Setenv("STRIPE_WEBHOOK_SECRET", "test-webhook-secret")
	tx := store.PaymentTransaction{ID: "tx-test", Provider: "stripe", UserID: "22222222-2222-2222-2222-222222222222", AmountCents: 50000, Currency: "KRW", Status: "pending"}
	memStore.Transactions[tx.ID] = &tx
	var req *http.Request
	var rec *httptest.ResponseRecorder
	// 2. First Webhook delivery -> creates balanced double-entry ledger entries
	whPayload, _ := json.Marshal(map[string]interface{}{
		"id": "wh-evt-1001", "type": "checkout.session.completed",
		"data": map[string]interface{}{"object": map[string]interface{}{
			"client_reference_id": tx.ID, "amount_total": 50000, "currency": "krw", "payment_status": "paid",
		}},
	})
	timestamp := fmt.Sprint(time.Now().Unix())
	mac := hmac.New(sha256.New, []byte("test-webhook-secret"))
	mac.Write([]byte(timestamp + "." + string(whPayload)))
	signature := fmt.Sprintf("t=%s,v1=%x", timestamp, mac.Sum(nil))
	req = httptest.NewRequest(http.MethodPost, "/v1/payments/webhooks/stripe", bytes.NewReader(whPayload))
	req.Header.Set("X-Idempotency-Key", "wh-evt-1001")
	req.Header.Set("Stripe-Signature", signature)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for first webhook, got %d: %s", rec.Code, rec.Body.String())
	}
	var whResp1 struct {
		IdempotentReplay bool                     `json:"idempotent_replay"`
		Transaction      store.PaymentTransaction `json:"transaction"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &whResp1)
	if whResp1.IdempotentReplay {
		t.Fatalf("expected idempotent_replay=false on first webhook")
	}
	if len(memStore.Ledger) != 3 {
		t.Fatalf("expected 3 double-entry ledger entries, got %d", len(memStore.Ledger))
	}

	// Verify Debit == Credit balance
	var debitSum, creditSum int64
	for _, entry := range memStore.Ledger {
		if entry.Direction == "DEBIT" {
			debitSum += entry.AmountCents
		} else if entry.Direction == "CREDIT" {
			creditSum += entry.AmountCents
		}
	}
	if debitSum != 50000 || creditSum != 50000 {
		t.Fatalf("double-entry ledger imbalance: debit=%d credit=%d", debitSum, creditSum)
	}

	// 3. Duplicate Webhook delivery with same idempotency key -> no duplicate ledger entries
	req = httptest.NewRequest(http.MethodPost, "/v1/payments/webhooks/stripe", bytes.NewReader(whPayload))
	req.Header.Set("X-Idempotency-Key", "wh-evt-1001")
	req.Header.Set("Stripe-Signature", signature)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for duplicate webhook, got %d", rec.Code)
	}
	var whResp2 struct {
		IdempotentReplay bool `json:"idempotent_replay"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &whResp2)
	if !whResp2.IdempotentReplay {
		t.Fatalf("expected idempotent_replay=true on duplicate webhook")
	}
	if len(memStore.Ledger) != 3 {
		t.Fatalf("expected ledger count to remain 3 after duplicate webhook, got %d", len(memStore.Ledger))
	}
}

func TestPlaylistsAndSubscriptions(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Create Playlist
	plBody := []byte(`{"title":"Distributed Systems Series","description":"Architecture deep dives","visibility":"public"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/playlists", bytes.NewReader(plBody))
	req.Header.Set("Authorization", "Bearer creator-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for playlist creation, got %d", rec.Code)
	}
	var pl store.Playlist
	_ = json.Unmarshal(rec.Body.Bytes(), &pl)

	// 2. Add item to Playlist
	itemBody := []byte(`{"video_id":"00000000-0000-0000-0000-000000000001"}`)
	req = httptest.NewRequest(http.MethodPost, "/v1/playlists/"+pl.ID+"/items", bytes.NewReader(itemBody))
	req.Header.Set("Authorization", "Bearer creator-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for adding playlist item, got %d", rec.Code)
	}

	// 3. Patch Playlist
	patchBody := []byte(`{"title":"Updated Systems Series"}`)
	req = httptest.NewRequest(http.MethodPatch, "/v1/playlists/"+pl.ID, bytes.NewReader(patchBody))
	req.Header.Set("Authorization", "Bearer creator-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for patching playlist, got %d", rec.Code)
	}

	// 4. Get Playlist
	req = httptest.NewRequest(http.MethodGet, "/v1/playlists/"+pl.ID, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for getting playlist, got %d", rec.Code)
	}

	// 5. Subscribe / List / Unsubscribe via /v1/subscriptions
	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req = httptest.NewRequest(http.MethodPost, "/v1/subscriptions/"+channelID, nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for subscribing, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/subscriptions", nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for listing subscriptions, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, "/v1/subscriptions/"+channelID, nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for unsubscribing, got %d", rec.Code)
	}
}

func TestAdDecisionTelemetryAndDiagnostics(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Valid Ad Decision for watch_preroll
	decReqBody := []byte(`{"slot_id":"watch_preroll","user_id":"33333333-3333-3333-3333-333333333333","category":"Technology","platform":"WEB"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/ads/decision", bytes.NewReader(decReqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for ad decision, got %d: %s", rec.Code, rec.Body.String())
	}

	var decResp struct {
		AdAvailable bool              `json:"ad_available"`
		DecisionID  string            `json:"decision_id"`
		CampaignID  string            `json:"campaign_id"`
		Reason      string            `json:"reason"`
		Creative    *store.AdCreative `json:"creative"`
		Tracking    map[string]string `json:"tracking"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &decResp); err != nil {
		t.Fatalf("failed to decode ad decision response: %v", err)
	}
	if !decResp.AdAvailable {
		t.Fatalf("expected ad_available == true for seeded watch_preroll slot, got false (reason=%s)", decResp.Reason)
	}
	if decResp.Creative == nil || decResp.Creative.MediaURL == "" {
		t.Fatalf("expected creative with media URL, got %+v", decResp.Creative)
	}
	if decResp.DecisionID == "" || decResp.CampaignID == "" {
		t.Fatalf("expected decision_id and campaign_id to be populated, got decID=%s campID=%s", decResp.DecisionID, decResp.CampaignID)
	}

	// 2. Telemetry: Record Impression Event
	impBody, _ := json.Marshal(map[string]interface{}{
		"event_id":    "evt-imp-test-1",
		"decision_id": decResp.DecisionID,
		"campaign_id": decResp.CampaignID,
		"event_type":  "impression",
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/ads/events", bytes.NewReader(impBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted && rec.Code != http.StatusOK {
		t.Fatalf("expected 200 or 202 for ad impression event, got %d: %s", rec.Code, rec.Body.String())
	}

	// Duplicate event deduplication
	req = httptest.NewRequest(http.MethodPost, "/v1/ads/events", bytes.NewReader(impBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK && rec.Code != http.StatusAccepted {
		t.Fatalf("expected 200 or 202 for duplicate ad event, got %d", rec.Code)
	}
	var dupResp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &dupResp)
	if dupResp["status"] != "DUPLICATE_IGNORED" {
		t.Fatalf("expected DUPLICATE_IGNORED for duplicate event_id, got %v", dupResp)
	}

	// Record Click Event
	clkBody, _ := json.Marshal(map[string]interface{}{
		"event_id":    "evt-clk-test-1",
		"decision_id": decResp.DecisionID,
		"campaign_id": decResp.CampaignID,
		"event_type":  "click",
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/ads/events", bytes.NewReader(clkBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted && rec.Code != http.StatusOK {
		t.Fatalf("expected 200 or 202 for ad click event, got %d", rec.Code)
	}

	// 3. Ad Preview endpoint
	prevBody := []byte(`{"slot_id":"home_feed_banner","category":"Technology"}`)
	req = httptest.NewRequest(http.MethodPost, "/v1/ads/preview", bytes.NewReader(prevBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for ad preview, got %d", rec.Code)
	}

	// 4. Admin Delivery Diagnostics
	req = httptest.NewRequest(http.MethodGet, "/v1/admin/ads/delivery-diagnostics", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for ad delivery diagnostics, got %d: %s", rec.Code, rec.Body.String())
	}
	var diagResp struct {
		Summary struct {
			TotalRequests   int64   `json:"total_requests"`
			ServedDecisions int64   `json:"served_decisions"`
			FillRatePct     float64 `json:"fill_rate_pct"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &diagResp); err != nil {
		t.Fatalf("failed to decode diagnostics response: %v", err)
	}
	if diagResp.Summary.TotalRequests < 1 || diagResp.Summary.ServedDecisions < 1 {
		t.Fatalf("expected at least 1 request and 1 served decision, got requests=%d served=%d", diagResp.Summary.TotalRequests, diagResp.Summary.ServedDecisions)
	}

	// 5. Unmatched Slot should return ad_available = false
	unmatchedBody := []byte(`{"slot_id":"non_existent_slot_xyz"}`)
	req = httptest.NewRequest(http.MethodPost, "/v1/ads/decision", bytes.NewReader(unmatchedBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for unmatched slot decision, got %d", rec.Code)
	}
	var unmatchResp struct {
		AdAvailable bool   `json:"ad_available"`
		Reason      string `json:"reason"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &unmatchResp)
	if unmatchResp.AdAvailable {
		t.Fatalf("expected ad_available == false for non-existent slot")
	}
	if unmatchResp.Reason != "UNKNOWN_SLOT_ID" {
		t.Fatalf("expected reason == 'UNKNOWN_SLOT_ID', got '%s'", unmatchResp.Reason)
	}
}

func TestE2EAdLifecycleAndDeliveryLoop(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Advertiser creates a new Campaign with creative (ARTICLE 7.2 Step 1)
	campPayload := []byte(`{
		"name": "Hyundai Ioniq 9 Electric SUV Global Launch",
		"format": "preroll",
		"totalBudget": 50000000,
		"dailyBudget": 5000000,
		"cpmBidCents": 18000,
		"cpcBidCents": 950,
		"categories": ["Technology", "Automotive"],
		"creativeMediaUrl": "https://media.gvp.media/ads/creatives/ioniq9_preroll_4k.mp4",
		"headline": "The All-New 3-Row Electric SUV Experience",
		"destinationUrl": "https://www.hyundai.com/ioniq9",
		"callToAction": "Explore Design"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/partners/campaigns", bytes.NewReader(campPayload))
	req.Header.Set("Authorization", "Bearer creator-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for campaign creation, got %d: %s", rec.Code, rec.Body.String())
	}
	var createdCamp struct {
		ID        string   `json:"id"`
		Creatives []string `json:"creatives"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &createdCamp); err != nil || createdCamp.ID == "" || len(createdCamp.Creatives) == 0 {
		t.Fatalf("failed to decode created campaign: %v", err)
	}

	// 2. Creative Governance & Approval (ARTICLE 7.2 Step 5)
	creativeID := createdCamp.Creatives[0]
	apprReq := httptest.NewRequest(http.MethodPost, "/v1/ads/admin/creatives/"+creativeID+"/approve", nil)
	apprRec := httptest.NewRecorder()
	router.ServeHTTP(apprRec, apprReq)
	if apprRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for creative approval, got %d", apprRec.Code)
	}

	// 3. User requests Ad Decision for watch_preroll (ARTICLE 7.2 Step 8-10)
	decPayload := []byte(`{"slot_id":"watch_preroll","user_id":"33333333-3333-3333-3333-333333333333","category":"Technology","platform":"WEB"}`)
	decReq := httptest.NewRequest(http.MethodPost, "/v1/ads/decision", bytes.NewReader(decPayload))
	decReq.Header.Set("Content-Type", "application/json")
	decRec := httptest.NewRecorder()
	router.ServeHTTP(decRec, decReq)
	if decRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for ad decision, got %d", decRec.Code)
	}
	var decResp struct {
		AdAvailable bool              `json:"ad_available"`
		DecisionID  string            `json:"decision_id"`
		CampaignID  string            `json:"campaign_id"`
		Creative    *store.AdCreative `json:"creative"`
	}
	if err := json.Unmarshal(decRec.Body.Bytes(), &decResp); err != nil || !decResp.AdAvailable {
		t.Fatalf("expected ad_available == true, got %+v, err=%v", decResp, err)
	}

	// 4. Client VideoPlayer renders and records verified impression (ARTICLE 7.2 Step 11-13)
	impEventID := "evt-e2e-imp-" + decResp.DecisionID
	impPayload, _ := json.Marshal(map[string]interface{}{
		"event_id":    impEventID,
		"decision_id": decResp.DecisionID,
		"campaign_id": decResp.CampaignID,
		"event_type":  "impression",
	})
	impReq := httptest.NewRequest(http.MethodPost, "/v1/ads/events", bytes.NewReader(impPayload))
	impReq.Header.Set("Content-Type", "application/json")
	impRec := httptest.NewRecorder()
	router.ServeHTTP(impRec, impReq)
	if impRec.Code != http.StatusOK && impRec.Code != http.StatusAccepted {
		t.Fatalf("expected 200/202 for ad impression event, got %d", impRec.Code)
	}

	// 5. Viewer clicks CTA and registers verified engagement click (ARTICLE 7.2 Step 14)
	clkEventID := "evt-e2e-clk-" + decResp.DecisionID
	clkPayload, _ := json.Marshal(map[string]interface{}{
		"event_id":    clkEventID,
		"decision_id": decResp.DecisionID,
		"campaign_id": decResp.CampaignID,
		"event_type":  "click",
	})
	clkReq := httptest.NewRequest(http.MethodPost, "/v1/ads/events", bytes.NewReader(clkPayload))
	clkReq.Header.Set("Content-Type", "application/json")
	clkRec := httptest.NewRecorder()
	router.ServeHTTP(clkRec, clkReq)
	if clkRec.Code != http.StatusOK && clkRec.Code != http.StatusAccepted {
		t.Fatalf("expected 200/202 for ad click event, got %d", clkRec.Code)
	}

	// 6. Verify real-time audit via Delivery Inspector & Advertiser Reports
	diagReq := httptest.NewRequest(http.MethodGet, "/v1/admin/ads/delivery-diagnostics", nil)
	diagRec := httptest.NewRecorder()
	router.ServeHTTP(diagRec, diagReq)
	if diagRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for delivery diagnostics, got %d", diagRec.Code)
	}

	repReq := httptest.NewRequest(http.MethodGet, "/v1/advertiser/reports", nil)
	repRec := httptest.NewRecorder()
	router.ServeHTTP(repRec, repReq)
	if repRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for advertiser reports, got %d", repRec.Code)
	}
}
