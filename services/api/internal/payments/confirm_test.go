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

const testCreatorID = "22222222-2222-2222-2222-222222222222"

func seedPendingTx(s *store.MemoryStore, id, provider string, amount int64) {
	s.Transactions[id] = &store.PaymentTransaction{
		ID:          id,
		UserID:      testCreatorID,
		Provider:    provider,
		Type:        "membership",
		AmountCents: amount,
		Currency:    "KRW",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
}

func confirmRequest(t *testing.T, router http.Handler, token string, payload map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/confirm", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// P0-3: without a Toss secret the confirmation must be refused, never marked succeeded.
func TestConfirmPaymentRefusesWithoutTossSecret(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	t.Setenv("TOSSPAYMENTS_SECRET_KEY", "")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	h.Adapters["toss"] = &TossPaymentsAdapter{}
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_no_secret", "toss", 15000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "toss", "paymentKey": "pk_any", "orderId": "tx_no_secret", "amount": 15000,
	})

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when Toss secret is unset, got %d (%s)", rec.Code, rec.Body.String())
	}
	if s.Transactions["tx_no_secret"].Status != "pending" {
		t.Fatalf("transaction must stay pending, got %s", s.Transactions["tx_no_secret"].Status)
	}
	if len(s.Ledger) != 0 {
		t.Fatalf("no ledger entries must be written on refused confirm, got %d", len(s.Ledger))
	}
}

// P0-3: unknown providers must be rejected, never marked succeeded.
func TestConfirmPaymentRejectsUnknownProvider(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_unknown", "toss", 15000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "bogus", "paymentKey": "pk_any", "orderId": "tx_unknown", "amount": 15000,
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown provider, got %d", rec.Code)
	}
	if s.Transactions["tx_unknown"].Status != "pending" {
		t.Fatalf("transaction must stay pending, got %s", s.Transactions["tx_unknown"].Status)
	}
}

// P0-3: client-asserted amount must match the recorded transaction.
func TestConfirmPaymentRejectsAmountMismatch(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	h.Adapters["toss"] = &TossPaymentsAdapter{SecretKey: "test_sk_sample12345"}
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_mismatch", "toss", 15000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "toss", "paymentKey": "pk_any", "orderId": "tx_mismatch", "amount": 100,
	})

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for amount mismatch, got %d", rec.Code)
	}
	if s.Transactions["tx_mismatch"].Status != "pending" {
		t.Fatalf("transaction must stay pending, got %s", s.Transactions["tx_mismatch"].Status)
	}
}

// P0-3: Toss DONE confirmation with matching amount succeeds (mock transport, no network).
func TestConfirmPaymentTossDoneSucceeds(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	toss := &TossPaymentsAdapter{SecretKey: "test_sk_sample12345"}
	toss.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"paymentKey":"pk_live_1","orderId":"tx_toss_ok","status":"DONE","totalAmount":15000,"method":"CARD"}`)),
			Header:     make(http.Header),
		}, nil
	})}
	h.Adapters["toss"] = toss
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_toss_ok", "toss", 15000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "toss", "paymentKey": "pk_live_1", "orderId": "tx_toss_ok", "amount": 15000,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for DONE confirmation, got %d (%s)", rec.Code, rec.Body.String())
	}
	if s.Transactions["tx_toss_ok"].Status != "succeeded" {
		t.Fatalf("transaction must be succeeded, got %s", s.Transactions["tx_toss_ok"].Status)
	}
}

// P0-3: non-DONE provider status must not confirm.
func TestConfirmPaymentTossNotDoneRejected(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	toss := &TossPaymentsAdapter{SecretKey: "test_sk_sample12345"}
	toss.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(`{"paymentKey":"pk_live_2","orderId":"tx_toss_pending","status":"WAITING_FOR_DEPOSIT","totalAmount":15000}`)),
			Header:     make(http.Header),
		}, nil
	})}
	h.Adapters["toss"] = toss
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_toss_pending", "toss", 15000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "toss", "paymentKey": "pk_live_2", "orderId": "tx_toss_pending", "amount": 15000,
	})

	if rec.Code == http.StatusOK {
		t.Fatal("non-DONE confirmation must not return 200")
	}
	if s.Transactions["tx_toss_pending"].Status != "pending" {
		t.Fatalf("transaction must stay pending, got %s", s.Transactions["tx_toss_pending"].Status)
	}
}

// P0-3: Stripe confirm without a secret must be refused, never marked succeeded.
func TestConfirmPaymentRefusesWithoutStripeSecret(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	t.Setenv("STRIPE_SECRET_KEY", "")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	h.Adapters["stripe"] = &StripeAdapter{}
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_stripe_no_secret", "stripe", 5000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "stripe", "paymentKey": "cs_test_1", "orderId": "tx_stripe_no_secret", "amount": 5000,
	})

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when Stripe secret is unset, got %d (%s)", rec.Code, rec.Body.String())
	}
	if s.Transactions["tx_stripe_no_secret"].Status != "pending" {
		t.Fatalf("transaction must stay pending, got %s", s.Transactions["tx_stripe_no_secret"].Status)
	}
}

// P0-3: Stripe paid session with matching amount/tx binding succeeds (mock transport).
func TestConfirmPaymentStripePaidSucceeds(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	stripe := &StripeAdapter{APIKey: "sk_test_123"}
	stripe.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/v1/checkout/sessions/cs_test_paid") {
			t.Fatalf("unexpected provider endpoint: %s %s", r.Method, r.URL.String())
		}
		if r.Header.Get("Authorization") != "Bearer "+stripe.APIKey {
			t.Fatal("missing Stripe authorization header")
		}
		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(
				`{"id":"cs_test_paid","payment_status":"paid","amount_total":5000,"currency":"krw","client_reference_id":"tx_stripe_ok"}`)),
			Header: make(http.Header),
		}, nil
	})}
	h.Adapters["stripe"] = stripe
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_stripe_ok", "stripe", 5000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "stripe", "paymentKey": "cs_test_paid", "orderId": "tx_stripe_ok", "amount": 5000,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for paid session, got %d (%s)", rec.Code, rec.Body.String())
	}
	if s.Transactions["tx_stripe_ok"].Status != "succeeded" {
		t.Fatalf("transaction must be succeeded, got %s", s.Transactions["tx_stripe_ok"].Status)
	}
}

// P0-3: Stripe session bound to a different transaction must not confirm this one.
func TestConfirmPaymentStripeSessionBindingEnforced(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	stripe := &StripeAdapter{APIKey: "sk_test_123"}
	stripe.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body: io.NopCloser(strings.NewReader(
				`{"id":"cs_test_other","payment_status":"paid","amount_total":5000,"currency":"krw","client_reference_id":"tx_someone_else"}`)),
			Header: make(http.Header),
		}, nil
	})}
	h.Adapters["stripe"] = stripe
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	seedPendingTx(s, "tx_stripe_bound", "stripe", 5000)
	rec := confirmRequest(t, router, "creator-token", map[string]interface{}{
		"provider": "stripe", "paymentKey": "cs_test_other", "orderId": "tx_stripe_bound", "amount": 5000,
	})

	if rec.Code == http.StatusOK {
		t.Fatal("session bound to another transaction must not confirm")
	}
	if s.Transactions["tx_stripe_bound"].Status != "pending" {
		t.Fatalf("transaction must stay pending, got %s", s.Transactions["tx_stripe_bound"].Status)
	}
}
