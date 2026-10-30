package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestStripeCheckoutCallsProvider(t *testing.T) {
	s := &StripeAdapter{APIKey: "test-key", SuccessURL: "https://site.test/success", CancelURL: "https://site.test/cancel"}
	s.Client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.stripe.com/v1/checkout/sessions" || r.Method != "POST" {
			t.Fatal("wrong provider endpoint")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" || r.Header.Get("Idempotency-Key") != "tx1" {
			t.Fatal("missing provider headers")
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.Form.Get("client_reference_id") != "tx1" || r.Form.Get("line_items[0][price_data][unit_amount]") != "5000" {
			t.Fatal("incorrect checkout parameters")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"id":"cs_1","url":"https://checkout.stripe.com/c/pay/cs_1"}`)), Header: make(http.Header)}, nil
	})}
	got, err := s.CreateCheckoutSession("tx1", 5000, "KRW", "donation")
	if err != nil || got != "https://checkout.stripe.com/c/pay/cs_1" {
		t.Fatalf("%s %v", got, err)
	}
}

func TestStripeMissingConfigurationDoesNotInventCheckout(t *testing.T) {
	t.Setenv("STRIPE_SECRET_KEY", "")
	if u, err := (&StripeAdapter{}).CreateCheckoutSession("tx", 100, "USD", "donation"); err == nil || u != "" {
		t.Fatal("invented checkout")
	}
}

func TestStripeSignatures(t *testing.T) {
	t.Setenv("STRIPE_WEBHOOK_SECRET", "")
	body := []byte(`{"id":"evt_1"}`)
	sign := func(at time.Time) string {
		timestamp := fmt.Sprint(at.Unix())
		mac := hmac.New(sha256.New, []byte("secret"))
		mac.Write(append([]byte(timestamp+"."), body...))
		return fmt.Sprintf("t=%s,v1=%x", timestamp, mac.Sum(nil))
	}
	s := &StripeAdapter{WebhookSecret: "secret"}
	if !s.VerifyWebhookSignature(sign(time.Now()), body) {
		t.Fatal("valid signature rejected")
	}
	if s.VerifyWebhookSignature(sign(time.Now()), []byte(`{"id":"tampered"}`)) {
		t.Fatal("tampered body accepted")
	}
	if s.VerifyWebhookSignature(sign(time.Now().Add(-10*time.Minute)), body) {
		t.Fatal("expired signature accepted")
	}
	if (&StripeAdapter{}).VerifyWebhookSignature("valid_sig", body) {
		t.Fatal("unconfigured signature accepted")
	}
}
