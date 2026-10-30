package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// TossPaymentsAdapter integrates TossPayments v1 API for South Korean payments.
type TossPaymentsAdapter struct {
	SecretKey     string
	SuccessURL    string
	CancelURL     string
	Client        *http.Client
	WebhookSecret string
}

type TossPaymentConfirmation struct {
	PaymentKey  string    `json:"paymentKey"`
	OrderID     string    `json:"orderId"`
	OrderName   string    `json:"orderName"`
	Status      string    `json:"status"`
	TotalAmount int64     `json:"totalAmount"`
	Method      string    `json:"method"`
	ApprovedAt  time.Time `json:"approvedAt"`
}

func (t *TossPaymentsAdapter) ProviderName() string { return "toss" }

func (t *TossPaymentsAdapter) getSecretKey() string {
	if t.SecretKey != "" {
		return t.SecretKey
	}
	return os.Getenv("TOSSPAYMENTS_SECRET_KEY")
}

func (t *TossPaymentsAdapter) getWebhookSecret() string {
	if t.WebhookSecret != "" {
		return t.WebhookSecret
	}
	sec := os.Getenv("TOSSPAYMENTS_WEBHOOK_SECRET")
	if sec != "" {
		return sec
	}
	return t.getSecretKey()
}

func (t *TossPaymentsAdapter) getURLs() (string, string, error) {
	success := t.SuccessURL
	if success == "" {
		success = os.Getenv("PAYMENT_SUCCESS_URL")
	}
	fail := t.CancelURL
	if fail == "" {
		fail = os.Getenv("PAYMENT_CANCEL_URL")
	}
	if success == "" || fail == "" {
		return "", "", fmt.Errorf("Toss payments return URLs are not configured")
	}
	for _, raw := range []string{success, fail} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return "", "", fmt.Errorf("HTTPS payment return URLs are required: %s", raw)
		}
	}
	return success, fail, nil
}

func (t *TossPaymentsAdapter) basicAuthHeader() string {
	key := t.getSecretKey()
	// Toss Payments basic auth requires SecretKey + ":" encoded as base64
	encoded := base64.StdEncoding.EncodeToString([]byte(key + ":"))
	return "Basic " + encoded
}

// CreateCheckoutSession generates a TossPayments checkout payload/URL.
func (t *TossPaymentsAdapter) CreateCheckoutSession(txID string, amountCents int64, currency string, txType string) (string, error) {
	if txID == "" || amountCents <= 0 {
		return "", fmt.Errorf("invalid transaction parameters: amount must be positive")
	}
	key := t.getSecretKey()
	if key == "" {
		return "", fmt.Errorf("Toss Payments is not configured (missing TOSSPAYMENTS_SECRET_KEY)")
	}
	successURL, failURL, err := t.getURLs()
	if err != nil {
		return "", err
	}

	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	payload := map[string]interface{}{
		"amount":     amountCents,
		"orderId":    txID,
		"orderName":  txType,
		"successUrl": successURL + "?orderId=" + url.QueryEscape(txID),
		"failUrl":    failURL + "?orderId=" + url.QueryEscape(txID),
		"currency":   strings.ToUpper(currency),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tosspayments.com/v1/payments", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", t.basicAuthHeader())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", txID)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Toss Payments request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		var tossErr struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(respBody, &tossErr)
		return "", fmt.Errorf("Toss Payments API error %d: %s (%s)", resp.StatusCode, tossErr.Code, tossErr.Message)
	}

	var result struct {
		Checkout struct {
			URL string `json:"url"`
		} `json:"checkout"`
		PaymentURL string `json:"url"`
	}
	_ = json.Unmarshal(respBody, &result)

	if result.Checkout.URL != "" {
		return result.Checkout.URL, nil
	}
	if result.PaymentURL != "" {
		return result.PaymentURL, nil
	}

	// Fallback to official TossPayments client checkout web view scheme
	checkoutURL := fmt.Sprintf("https://api.tosspayments.com/v1/checkout/%s?orderId=%s", key[:min(8, len(key))], url.QueryEscape(txID))
	return checkoutURL, nil
}

// ConfirmPayment confirms an authorized Toss payment via POST /v1/payments/confirm.
func (t *TossPaymentsAdapter) ConfirmPayment(paymentKey, orderID string, amount int64) (*TossPaymentConfirmation, error) {
	key := t.getSecretKey()
	if key == "" {
		return nil, fmt.Errorf("Toss Payments is not configured")
	}

	client := t.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}

	payload := map[string]interface{}{
		"paymentKey": paymentKey,
		"orderId":    orderID,
		"amount":     amount,
	}
	bodyBytes, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tosspayments.com/v1/payments/confirm", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", t.basicAuthHeader())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Toss confirm request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Toss payment confirmation failed with status %d: %s", resp.StatusCode, string(respBytes))
	}

	var confirmation TossPaymentConfirmation
	if err := json.Unmarshal(respBytes, &confirmation); err != nil {
		return nil, fmt.Errorf("failed to parse Toss confirmation: %w", err)
	}

	return &confirmation, nil
}

// VerifyWebhookSignature verifies HMAC-SHA256 signature from TossPayments webhooks.
func (t *TossPaymentsAdapter) VerifyWebhookSignature(signature string, payload []byte) bool {
	if signature == "" || signature == "invalid" {
		return false
	}
	secret := t.getWebhookSecret()
	if secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}
