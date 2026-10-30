package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTossPaymentsCheckoutCallsProvider(t *testing.T) {
	adapter := &TossPaymentsAdapter{
		SecretKey:  "test_sk_sample12345",
		SuccessURL: "https://site.test/success",
		CancelURL:  "https://site.test/cancel",
	}

	adapter.Client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.String() != "https://api.tosspayments.com/v1/payments" || r.Method != "POST" {
				t.Fatalf("unexpected provider endpoint: %s %s", r.Method, r.URL.String())
			}
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
				t.Fatal("missing or invalid Toss basic authorization header")
			}
			if r.Header.Get("Idempotency-Key") != "tx_toss_1" {
				t.Fatal("missing idempotency key header")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"checkout":{"url":"https://api.tosspayments.com/v1/checkout/test1234"}}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	url, err := adapter.CreateCheckoutSession("tx_toss_1", 10000, "KRW", "SuperChat ₩10,000")
	if err != nil {
		t.Fatalf("unexpected error creating checkout session: %v", err)
	}
	if url != "https://api.tosspayments.com/v1/checkout/test1234" {
		t.Fatalf("expected checkout url https://api.tosspayments.com/v1/checkout/test1234, got %s", url)
	}
}

func TestTossPaymentsMissingConfigDoesNotInventCheckout(t *testing.T) {
	t.Setenv("TOSSPAYMENTS_SECRET_KEY", "")
	adapter := &TossPaymentsAdapter{}
	if u, err := adapter.CreateCheckoutSession("tx", 10000, "KRW", "donation"); err == nil || u != "" {
		t.Fatalf("expected failure for unconfigured Toss adapter, got %s, err: %v", u, err)
	}
}

func TestTossPaymentsSignatures(t *testing.T) {
	payload := []byte(`{"eventType":"PAYMENT_STATUS_CHANGED","data":{"paymentKey":"pk_123","status":"DONE"}}`)
	secret := "test_webhook_secret_key"

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSig := hex.EncodeToString(mac.Sum(nil))

	adapter := &TossPaymentsAdapter{WebhookSecret: secret}
	if !adapter.VerifyWebhookSignature(validSig, payload) {
		t.Fatal("valid Toss webhook signature rejected")
	}

	if adapter.VerifyWebhookSignature("invalid_signature", payload) {
		t.Fatal("invalid Toss webhook signature accepted")
	}
}

func TestTossPaymentsConfirm(t *testing.T) {
	adapter := &TossPaymentsAdapter{
		SecretKey: "test_sk_sample12345",
	}

	adapter.Client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.String() != "https://api.tosspayments.com/v1/payments/confirm" || r.Method != "POST" {
				t.Fatalf("unexpected provider endpoint: %s %s", r.Method, r.URL.String())
			}
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
				t.Fatal("missing or invalid Toss basic authorization header")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(`{"paymentKey":"pk_live_999","orderId":"tx_order_1","status":"DONE","totalAmount":15000,"method":"CARD"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	conf, err := adapter.ConfirmPayment("pk_live_999", "tx_order_1", 15000)
	if err != nil {
		t.Fatalf("unexpected error during confirm: %v", err)
	}
	if conf.PaymentKey != "pk_live_999" || conf.Status != "DONE" || conf.TotalAmount != 15000 {
		t.Fatalf("unexpected confirmation result: %+v", conf)
	}
}
