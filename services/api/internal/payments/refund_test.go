package payments

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func seedSucceededTx(s *store.MemoryStore, id, provider, ref string, amount int64) {
	s.Transactions[id] = &store.PaymentTransaction{
		ID:             id,
		UserID:         testCreatorID,
		Provider:       provider,
		ProviderRef:    ref,
		Type:           "membership",
		AmountCents:    amount,
		Currency:       "KRW",
		Status:         "succeeded",
		IdempotencyKey: "idem-" + id,
		CreatedAt:      time.Now(),
		LedgerEntries:  CreateDoubleEntryLedger(id, amount, "KRW"),
	}
	s.Ledger = append(s.Ledger, s.Transactions[id].LedgerEntries...)
}

func refundRequest(t *testing.T, router http.Handler, orderID, token string, payload map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/"+orderID+"/refund", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func refundRouter(s *store.MemoryStore, h *Handler) http.Handler {
	auth := identity.NewAuthMiddleware(s)
	return auth.Authenticate(h.Routes(auth))
}

// P1-17: refunds fail closed without provider credentials.
func TestRefundRefusesWithoutProviderSecret(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	t.Setenv("TOSSPAYMENTS_SECRET_KEY", "")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	h.Adapters["toss"] = &TossPaymentsAdapter{}
	router := refundRouter(s, h)

	seedSucceededTx(s, "tx_ref_nosecret", "toss", "pk_live_9", 15000)
	rec := refundRequest(t, router, "tx_ref_nosecret", "creator-token", map[string]interface{}{"reason": "changed mind"})
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 without secret, got %d (%s)", rec.Code, rec.Body.String())
	}
	if s.Transactions["tx_ref_nosecret"].Status != "succeeded" {
		t.Fatal("transaction must stay succeeded after refused refund")
	}
}

// P1-17: Toss full refund reverses the ledger and stays balanced.
func TestRefundTossFullSucceeds(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	toss := &TossPaymentsAdapter{SecretKey: "test_sk_sample12345"}
	toss.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/v1/payments/pk_live_9/cancel") {
			t.Fatalf("unexpected endpoint: %s %s", r.Method, r.URL.String())
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"paymentKey":"pk_live_9","orderId":"tx_ref_ok","status":"CANCELED","totalAmount":15000,"cancels":[{"cancelAmount":15000,"cancelReason":"test"}]}`)),
			Header:     make(http.Header),
		}, nil
	})}
	h.Adapters["toss"] = toss
	router := refundRouter(s, h)

	seedSucceededTx(s, "tx_ref_ok", "toss", "pk_live_9", 15000)
	rec := refundRequest(t, router, "tx_ref_ok", "creator-token", map[string]interface{}{"reason": "changed mind"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 refund, got %d: %s", rec.Code, rec.Body.String())
	}
	tx := s.Transactions["tx_ref_ok"]
	if tx.Status != "refunded" {
		t.Fatalf("expected refunded, got %s", tx.Status)
	}
	var debits, credits int64
	for _, e := range s.Ledger {
		if e.TransactionID != "tx_ref_ok" {
			continue
		}
		if e.Direction == "DEBIT" {
			debits += e.AmountCents
		} else {
			credits += e.AmountCents
		}
	}
	if debits != credits || debits != 30000 {
		t.Fatalf("ledger must balance at 30000 after reversal, got debit=%d credit=%d", debits, credits)
	}

	// Idempotent replay.
	rec = refundRequest(t, router, "tx_ref_ok", "creator-token", map[string]interface{}{"reason": "again"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 replay, got %d", rec.Code)
	}
}

// P1-17: strangers cannot refund; only pending/succeeded states refund.
func TestRefundOwnershipAndState(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	h.Adapters["toss"] = &TossPaymentsAdapter{SecretKey: "test_sk_sample12345"}
	router := refundRouter(s, h)

	seedSucceededTx(s, "tx_ref_own", "toss", "pk_live_9", 15000)
	rec := refundRequest(t, router, "tx_ref_own", "viewer-token", map[string]interface{}{"reason": "x"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 stranger refund, got %d", rec.Code)
	}

	s.Transactions["tx_ref_pending"] = &store.PaymentTransaction{
		ID: "tx_ref_pending", UserID: testCreatorID, Provider: "toss",
		Type: "membership", AmountCents: 15000, Currency: "KRW",
		Status: "pending", CreatedAt: time.Now(),
	}
	rec = refundRequest(t, router, "tx_ref_pending", "creator-token", map[string]interface{}{"reason": "x"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 pending refund, got %d", rec.Code)
	}
}

// P1-17: Stripe refund resolves the payment intent then refunds it.
func TestRefundStripeSucceeds(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	stripe := &StripeAdapter{APIKey: "sk_test_123"}
	calls := 0
	stripe.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if strings.HasSuffix(r.URL.Path, "/v1/payment_intents/pi_1/refunds") && r.Method == http.MethodPost {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"id":"re_1","status":"succeeded","amount":5000}`)),
				Header:     make(http.Header),
			}, nil
		}
		if strings.HasPrefix(r.URL.Path, "/v1/checkout/sessions/cs_test_paid") && r.Method == http.MethodGet {
			return &http.Response{
				StatusCode: 200,
				Body: io.NopCloser(strings.NewReader(
					`{"id":"cs_test_paid","payment_status":"paid","amount_total":5000,"currency":"krw","client_reference_id":"tx_ref_stripe","payment_intent":"pi_1"}`)),
				Header: make(http.Header),
			}, nil
		}
		t.Fatalf("unexpected endpoint: %s %s", r.Method, r.URL.String())
		return nil, nil
	})}
	h.Adapters["stripe"] = stripe
	router := refundRouter(s, h)

	seedSucceededTx(s, "tx_ref_stripe", "stripe", "cs_test_paid", 5000)
	rec := refundRequest(t, router, "tx_ref_stripe", "creator-token", map[string]interface{}{"reason": "changed mind"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 stripe refund, got %d: %s", rec.Code, rec.Body.String())
	}
	if s.Transactions["tx_ref_stripe"].Status != "refunded" {
		t.Fatalf("expected refunded, got %s", s.Transactions["tx_ref_stripe"].Status)
	}
	if calls != 2 {
		t.Fatalf("expected session+refund calls, got %d", calls)
	}
}
