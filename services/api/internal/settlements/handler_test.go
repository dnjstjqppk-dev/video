package settlements

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func newSettlementsTestServer() (*httptest.Server, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	settleH := NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/settlements", settleH.Routes(auth))

	return httptest.NewServer(r), memStore
}

func settleReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
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

func TestGetSettlementsAndEarnings(t *testing.T) {
	srv, memStore := newSettlementsTestServer()
	defer srv.Close()

	// Seed a succeeded transaction plus its payable leg for the creator.
	memStore.Transactions["tx_01"] = &store.PaymentTransaction{
		ID: "tx_01", UserID: "22222222-2222-2222-2222-222222222222",
		Provider: "stripe", Type: "membership", AmountCents: 10000,
		Currency: "KRW", Status: "succeeded", CreatedAt: time.Now(),
	}
	memStore.Ledger = append(memStore.Ledger, store.LedgerEntry{
		ID:            "led_01",
		TransactionID: "tx_01",
		AccountType:   "creator_payable",
		Direction:     "CREDIT",
		AmountCents:   7000,
		Currency:      "KRW",
		CreatedAt:     time.Now(),
	})

	// 1. Get settlements (P1-18: authenticated + scoped, no fake base)
	req, rec := settleReq(http.MethodGet, "/settlements/settlements", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 anonymous, got %d", rec.Code)
	}
	req, rec = settleReq(http.MethodGet, "/settlements/settlements", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for settlements, got %d", rec.Code)
	}

	var settlements []map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &settlements)
	if len(settlements) != 1 {
		t.Fatalf("expected exactly 1 accrued period, got %d", len(settlements))
	}

	// 2. Get earnings
	req, rec = settleReq(http.MethodGet, "/settlements/earnings", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for earnings, got %d", rec.Code)
	}
	var earnings map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &earnings)
	if earnings["currency"] != "KRW" {
		t.Fatalf("expected currency KRW, got %v", earnings["currency"])
	}

	// 3. Request payout (authenticated)
	req, rec = settleReq(http.MethodPost, "/settlements/payout", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for payout request, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Cross-owner channel payout is rejected
	req, rec = settleReq(http.MethodPost, "/settlements/payout", "viewer-token", map[string]string{
		"channelId": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-owner payout, got %d: %s", rec.Code, rec.Body.String())
	}
}
