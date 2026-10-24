package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func TestAPIHealthzAndFeeds(t *testing.T) {
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Healthz
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for /healthz, got %d", rec.Code)
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
	memStore := store.NewMemoryStore()
	router := buildRouter(memStore)

	// 1. Checkout session
	checkoutBody := []byte(`{"provider":"toss","type":"super_chat","amount_cents":50000,"currency":"KRW","idempotency_key":"chk-key-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/payments/checkout", bytes.NewReader(checkoutBody))
	req.Header.Set("Authorization", "Bearer creator-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for checkout, got %d: %s", rec.Code, rec.Body.String())
	}
	var tx store.PaymentTransaction
	if err := json.Unmarshal(rec.Body.Bytes(), &tx); err != nil || tx.ID == "" {
		t.Fatalf("failed to parse checkout transaction: %v", err)
	}

	// 2. First Webhook delivery -> creates balanced double-entry ledger entries
	whPayload, _ := json.Marshal(map[string]interface{}{
		"idempotency_key": "wh-evt-1001",
		"transaction_id":  tx.ID,
		"amount_cents":    50000,
		"currency":        "KRW",
	})
	req = httptest.NewRequest(http.MethodPost, "/v1/payments/webhooks/toss", bytes.NewReader(whPayload))
	req.Header.Set("X-Idempotency-Key", "wh-evt-1001")
	req.Header.Set("X-Webhook-Signature", "valid_sig")
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
	req = httptest.NewRequest(http.MethodPost, "/v1/payments/webhooks/toss", bytes.NewReader(whPayload))
	req.Header.Set("X-Idempotency-Key", "wh-evt-1001")
	req.Header.Set("X-Webhook-Signature", "valid_sig")
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
