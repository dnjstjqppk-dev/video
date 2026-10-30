package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *StripeAdapter) createSession(txID string, amount int64, currency, kind string) (string, error) {
	key, success, cancel := s.APIKey, s.SuccessURL, s.CancelURL
	if key == "" {
		key = os.Getenv("STRIPE_SECRET_KEY")
	}
	if success == "" {
		success = os.Getenv("PAYMENT_SUCCESS_URL")
	}
	if cancel == "" {
		cancel = os.Getenv("PAYMENT_CANCEL_URL")
	}
	if key == "" {
		return "", fmt.Errorf("Stripe is not configured")
	}
	for _, raw := range []string{success, cancel} {
		u, err := url.Parse(raw)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
			return "", fmt.Errorf("HTTPS payment return URLs are required")
		}
	}
	form := url.Values{
		"mode": {"payment"}, "success_url": {success}, "cancel_url": {cancel},
		"client_reference_id": {txID}, "metadata[transaction_id]": {txID},
		"line_items[0][quantity]":                       {"1"},
		"line_items[0][price_data][currency]":           {strings.ToLower(currency)},
		"line_items[0][price_data][unit_amount]":        {strconv.FormatInt(amount, 10)},
		"line_items[0][price_data][product_data][name]": {kind},
	}
	ctx, cancelRequest := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelRequest()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", txID)
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Stripe request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("Stripe returned HTTP %d", resp.StatusCode)
	}
	var session struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&session); err != nil {
		return "", fmt.Errorf("Invalid Stripe response")
	}
	u, err := url.Parse(session.URL)
	if err != nil || session.ID == "" || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return "", fmt.Errorf("Missing Stripe checkout URL")
	}
	return session.URL, nil
}
