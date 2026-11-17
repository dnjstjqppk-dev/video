package settlements

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/store"
)

const (
	creatorID = "22222222-2222-2222-2222-222222222222"
	viewerID  = "33333333-3333-3333-3333-333333333333"
)

func seedSucceededTx(s *store.MemoryStore, userID string, cents int64) {
	txID := "tx-settle-" + userID[:8]
	s.Transactions[txID] = &store.PaymentTransaction{
		ID: txID, UserID: userID, Provider: "stripe",
		Type: "membership", AmountCents: cents, Currency: "KRW",
		Status: "succeeded", CreatedAt: time.Now(),
	}
	share := (cents / 100) * 70
	s.Ledger = append(s.Ledger, store.LedgerEntry{
		ID: "led-" + userID[:8], TransactionID: txID,
		AccountType: "creator_payable", Direction: "CREDIT",
		AmountCents: share, Currency: "KRW", CreatedAt: time.Now(),
	})
}

// P1-18: settlement surfaces require authentication.
func TestSettlementsRequireAuth(t *testing.T) {
	srv, _ := newSettlementsTestServer()
	defer srv.Close()
	for _, path := range []string{"/settlements/settlements", "/settlements/earnings"} {
		req, rec := settleReq(http.MethodGet, path, "", nil)
		srv.Config.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("%s: expected 401 anonymous, got %d", path, rec.Code)
		}
	}
}

// P1-18: creators see only their own earnings; strangers see zero, never others'.
func TestEarningsScopedToCaller(t *testing.T) {
	srv, memStore := newSettlementsTestServer()
	defer srv.Close()
	seedSucceededTx(memStore, creatorID, 10000)

	req, rec := settleReq(http.MethodGet, "/settlements/earnings", "viewer-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var viewerBody struct {
		CurrentMonthEstimated int64 `json:"currentMonthEstimated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &viewerBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if viewerBody.CurrentMonthEstimated != 0 {
		t.Fatalf("stranger must see zero earnings, got %d", viewerBody.CurrentMonthEstimated)
	}

	req, rec = settleReq(http.MethodGet, "/settlements/earnings", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	var creatorBody struct {
		CurrentMonthEstimated int64 `json:"currentMonthEstimated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &creatorBody); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if creatorBody.CurrentMonthEstimated != 7000 {
		t.Fatalf("creator must see own 70%% share 7000, got %d", creatorBody.CurrentMonthEstimated)
	}
}

// P1-18: settlement rows derive from the caller's ledger only, no fake base.
func TestSettlementsScopedNoFakeBase(t *testing.T) {
	srv, memStore := newSettlementsTestServer()
	defer srv.Close()
	seedSucceededTx(memStore, creatorID, 10000)

	req, rec := settleReq(http.MethodGet, "/settlements/settlements", "viewer-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	var rows []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("stranger must see no settlement rows, got %d", len(rows))
	}

	req, rec = settleReq(http.MethodGet, "/settlements/settlements", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if err := json.Unmarshal(rec.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("creator must see exactly own accrued row, got %d", len(rows))
	}
	if gross, _ := rows[0]["grossAmount"].(float64); gross != 7000 {
		t.Fatalf("expected gross 7000 from own ledger, got %v", rows[0]["grossAmount"])
	}
}
