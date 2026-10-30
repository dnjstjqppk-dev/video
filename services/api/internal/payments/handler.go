package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

// PaymentAdapter abstracts external payment gateways (Stripe & TossPayments).
type PaymentAdapter interface {
	ProviderName() string
	CreateCheckoutSession(txID string, amountCents int64, currency string, txType string) (string, error)
	VerifyWebhookSignature(signature string, payload []byte) bool
}

type StripeAdapter struct {
	WebhookSecret string
	APIKey        string
	SuccessURL    string
	CancelURL     string
	Client        *http.Client
}

func (s *StripeAdapter) ProviderName() string { return "stripe" }
func (s *StripeAdapter) CreateCheckoutSession(txID string, amountCents int64, currency string, txType string) (string, error) {
	if txID == "" || amountCents <= 0 {
		return "", fmt.Errorf("invalid transaction parameters: amount must be positive")
	}
	return s.createSession(txID, amountCents, currency, txType)
}

func (s *StripeAdapter) VerifyWebhookSignature(signatureHeader string, payload []byte) bool {
	if signatureHeader == "" || signatureHeader == "invalid" {
		return false
	}
	secret := s.WebhookSecret
	if secret == "" {
		secret = os.Getenv("STRIPE_WEBHOOK_SECRET")
	}
	// If a production secret is configured, perform exact HMAC-SHA256 validation
	if secret != "" {
		// Parse Stripe signature format: t=timestamp,v1=hash
		var timestamp, sigHash string
		for _, part := range strings.Split(signatureHeader, ",") {
			kv := strings.SplitN(part, "=", 2)
			if len(kv) == 2 {
				if kv[0] == "t" {
					timestamp = kv[1]
				} else if kv[0] == "v1" {
					sigHash = kv[1]
				}
			}
		}
		if timestamp == "" || sigHash == "" {
			return false
		}
		seconds, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil || seconds < time.Now().Add(-5*time.Minute).Unix() || seconds > time.Now().Add(time.Minute).Unix() {
			return false
		}
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(timestamp + "." + string(payload)))
		expected := hex.EncodeToString(mac.Sum(nil))
		return subtle.ConstantTimeCompare([]byte(expected), []byte(sigHash)) == 1
	}

	return false
}

type TossPaymentsAdapter struct {
	SecretKey string
}

func (t *TossPaymentsAdapter) ProviderName() string { return "toss" }
func (t *TossPaymentsAdapter) CreateCheckoutSession(txID string, amountCents int64, currency string, txType string) (string, error) {
	if txID == "" || amountCents <= 0 {
		return "", fmt.Errorf("invalid transaction parameters")
	}
	return "", fmt.Errorf("Toss checkout integration is not configured")
}

func (t *TossPaymentsAdapter) VerifyWebhookSignature(signature string, payload []byte) bool {
	if signature == "" || signature == "invalid" {
		return false
	}
	secret := t.SecretKey
	if secret == "" {
		secret = os.Getenv("TOSSPAYMENTS_WEBHOOK_SECRET")
	}
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(payload)
		expected := hex.EncodeToString(mac.Sum(nil))
		return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
	}
	return false
}

type Handler struct {
	Store    *store.MemoryStore
	Adapters map[string]PaymentAdapter
	mu       sync.Mutex
}

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{
		Store: s,
		Adapters: map[string]PaymentAdapter{
			"stripe": &StripeAdapter{},
		},
	}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	// Webhooks are authenticated via provider signature & idempotency key
	r.Post("/webhooks/{provider}", h.HandleWebhook)

	// Authenticated user payment endpoints
	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/checkout", h.CreateCheckout)
		pr.Get("/transactions", h.ListTransactions)
	})

	return r
}

func (h *Handler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var payload struct {
		Provider       string `json:"provider"`
		Type           string `json:"type"`
		AmountCents    int64  `json:"amount_cents"`
		Currency       string `json:"currency"`
		IdempotencyKey string `json:"idempotency_key"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.AmountCents <= 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Valid amount_cents is required")
		return
	}

	provider := strings.ToLower(payload.Provider)
	if provider == "" {
		provider = "stripe"
	}
	adapter, ok := h.Adapters[provider]
	if !ok {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unsupported payment provider")
		return
	}

	currency := strings.ToUpper(payload.Currency)
	if currency == "" {
		currency = "KRW"
	}
	txType := payload.Type
	if txType == "" {
		txType = "membership"
	}

	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey == "" {
		idemKey = payload.IdempotencyKey
	}
	if idemKey == "" {
		idemKey = uuid.NewString()
	}
	storeKey := "checkout:" + user.ID + ":" + idemKey

	h.mu.Lock()
	defer h.mu.Unlock()

	if existingTxID, exists := h.Store.IdempotencyKeys[storeKey]; exists {
		if existingTx, found := h.Store.Transactions[existingTxID]; found {
			if existingTx.AmountCents != payload.AmountCents || existingTx.Currency != currency || existingTx.Provider != provider || existingTx.Type != txType {
				identity.WriteError(w, r, http.StatusConflict, "CONFLICT", "Idempotency key reused with different checkout parameters")
				return
			}
			identity.WriteJSON(w, http.StatusOK, existingTx)
			return
		}
	}

	txID := uuid.NewString()
	checkoutURL, err := adapter.CreateCheckoutSession(txID, payload.AmountCents, currency, txType)
	if err != nil {
		identity.WriteError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Payment provider could not initialize checkout")
		return
	}

	tx := &store.PaymentTransaction{
		ID:             txID,
		UserID:         user.ID,
		Provider:       provider,
		IdempotencyKey: idemKey,
		Type:           txType,
		AmountCents:    payload.AmountCents,
		Currency:       currency,
		Status:         "pending",
		CheckoutURL:    checkoutURL,
		CreatedAt:      time.Now(),
	}

	h.Store.Transactions[txID] = tx
	h.Store.IdempotencyKeys[storeKey] = txID

	identity.WriteJSON(w, http.StatusCreated, tx)
}

// CreateDoubleEntryLedger generates balanced double-entry accounting records (70% creator, 30% platform).
func CreateDoubleEntryLedger(txID string, amountCents int64, currency string) []store.LedgerEntry {
	now := time.Now()
	creatorShare := (amountCents/100)*70 + ((amountCents%100)*70)/100
	platformFee := amountCents - creatorShare

	return []store.LedgerEntry{
		{
			ID:            uuid.NewString(),
			TransactionID: txID,
			AccountType:   "platform_escrow",
			Direction:     "DEBIT",
			AmountCents:   amountCents,
			Currency:      currency,
			CreatedAt:     now,
		},
		{
			ID:            uuid.NewString(),
			TransactionID: txID,
			AccountType:   "creator_payable",
			Direction:     "CREDIT",
			AmountCents:   creatorShare,
			Currency:      currency,
			CreatedAt:     now,
		},
		{
			ID:            uuid.NewString(),
			TransactionID: txID,
			AccountType:   "platform_revenue",
			Direction:     "CREDIT",
			AmountCents:   platformFee,
			Currency:      currency,
			CreatedAt:     now,
		},
	}
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	provider := strings.ToLower(chi.URLParam(r, "provider"))
	adapter, ok := h.Adapters[provider]
	if !ok {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unknown payment provider")
		return
	}

	sig := r.Header.Get("X-Webhook-Signature")
	if provider == "stripe" {
		sig = r.Header.Get("Stripe-Signature")
	}
	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		identity.WriteError(w, r, http.StatusRequestEntityTooLarge, "INVALID_ARGUMENT", "Invalid or oversized webhook body")
		return
	}
	if !adapter.VerifyWebhookSignature(sig, rawBody) {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Invalid webhook signature")
		return
	}

	var payload struct {
		IdempotencyKey string `json:"idempotency_key"`
		EventID        string `json:"event_id"`
		TransactionID  string `json:"transaction_id"`
		UserID         string `json:"user_id"`
		Type           string `json:"type"`
		AmountCents    int64  `json:"amount_cents"`
		Currency       string `json:"currency"`
		Status         string `json:"status"`
	}

	if err := json.Unmarshal(rawBody, &payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid webhook payload")
		return
	}
	if provider == "stripe" {
		var event struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Data struct {
				Object struct {
					ClientReferenceID string `json:"client_reference_id"`
					AmountTotal       int64  `json:"amount_total"`
					Currency          string `json:"currency"`
					PaymentStatus     string `json:"payment_status"`
				} `json:"object"`
			} `json:"data"`
		}
		if json.Unmarshal(rawBody, &event) != nil || event.ID == "" {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid Stripe event")
			return
		}
		if (event.Type != "checkout.session.completed" && event.Type != "checkout.session.async_payment_succeeded") || event.Data.Object.PaymentStatus != "paid" {
			identity.WriteJSON(w, http.StatusOK, map[string]bool{"ignored": true})
			return
		}
		payload.EventID = event.ID
		payload.TransactionID = event.Data.Object.ClientReferenceID
		payload.AmountCents = event.Data.Object.AmountTotal
		payload.Currency = event.Data.Object.Currency
		payload.Status = "succeeded"
	}

	idemKey := payload.EventID
	if idemKey == "" {
		idemKey = payload.IdempotencyKey
	}
	if idemKey == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Idempotency key or event_id is required")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	webhookKey := "wh:" + provider + ":" + idemKey
	if existingTxID, exists := h.Store.IdempotencyKeys[webhookKey]; exists {
		existingTx := h.Store.Transactions[existingTxID]
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"idempotent_replay": true,
			"transaction":       existingTx,
		})
		return
	}

	tx, exists := h.Store.Transactions[payload.TransactionID]
	if !exists || tx.Provider != provider {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Matching checkout transaction not found")
		return
	}
	if payload.AmountCents != tx.AmountCents || !strings.EqualFold(payload.Currency, tx.Currency) {
		identity.WriteError(w, r, http.StatusConflict, "CONFLICT", "Payment amount or currency mismatch")
		return
	}
	if payload.Status != "succeeded" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unsupported payment status")
		return
	}
	if tx.Status == "succeeded" {
		h.Store.IdempotencyKeys[webhookKey] = tx.ID
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{"idempotent_replay": true, "transaction": tx})
		return
	}
	tx.Status = "succeeded"
	entries := CreateDoubleEntryLedger(tx.ID, tx.AmountCents, tx.Currency)
	tx.LedgerEntries = entries
	h.Store.Ledger = append(h.Store.Ledger, entries...)
	h.Store.IdempotencyKeys[webhookKey] = tx.ID

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"idempotent_replay": false,
		"transaction":       tx,
	})
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var list []*store.PaymentTransaction
	for _, tx := range h.Store.Transactions {
		if tx.UserID == user.ID || user.Role == "super_admin" {
			list = append(list, tx)
		}
	}
	if list == nil {
		list = make([]*store.PaymentTransaction, 0)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}
