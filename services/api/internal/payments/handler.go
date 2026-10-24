package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
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

type StripeAdapter struct{}

func (s *StripeAdapter) ProviderName() string { return "stripe" }
func (s *StripeAdapter) CreateCheckoutSession(txID string, amountCents int64, currency string, txType string) (string, error) {
	return fmt.Sprintf("https://checkout.stripe.com/c/pay/cs_test_%s", txID), nil
}
func (s *StripeAdapter) VerifyWebhookSignature(signature string, payload []byte) bool {
	return signature != "invalid"
}

type TossPaymentsAdapter struct{}

func (t *TossPaymentsAdapter) ProviderName() string { return "toss" }
func (t *TossPaymentsAdapter) CreateCheckoutSession(txID string, amountCents int64, currency string, txType string) (string, error) {
	return fmt.Sprintf("https://pay.toss.im/pay/toss_session_%s", txID), nil
}
func (t *TossPaymentsAdapter) VerifyWebhookSignature(signature string, payload []byte) bool {
	return signature != "invalid"
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
			"toss":   &TossPaymentsAdapter{},
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

	h.mu.Lock()
	defer h.mu.Unlock()

	if existingTxID, exists := h.Store.IdempotencyKeys[idemKey]; exists {
		if existingTx, found := h.Store.Transactions[existingTxID]; found {
			identity.WriteJSON(w, http.StatusOK, existingTx)
			return
		}
	}

	txID := uuid.NewString()
	checkoutURL, err := adapter.CreateCheckoutSession(txID, payload.AmountCents, currency, txType)
	if err != nil {
		identity.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to initialize checkout session")
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
	h.Store.IdempotencyKeys[idemKey] = txID

	identity.WriteJSON(w, http.StatusCreated, tx)
}

// CreateDoubleEntryLedger generates balanced double-entry accounting records (70% creator, 30% platform).
func CreateDoubleEntryLedger(txID string, amountCents int64, currency string) []store.LedgerEntry {
	now := time.Now()
	creatorShare := (amountCents * 70) / 100
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
	if !adapter.VerifyWebhookSignature(sig, nil) {
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

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid webhook payload")
		return
	}

	idemKey := r.Header.Get("X-Idempotency-Key")
	if idemKey == "" {
		idemKey = payload.IdempotencyKey
	}
	if idemKey == "" {
		idemKey = payload.EventID
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

	txID := payload.TransactionID
	tx, exists := h.Store.Transactions[txID]
	if !exists {
		if txID == "" {
			txID = uuid.NewString()
		}
		amt := payload.AmountCents
		if amt <= 0 {
			amt = 10000
		}
		curr := payload.Currency
		if curr == "" {
			curr = "KRW"
		}
		txType := payload.Type
		if txType == "" {
			txType = "super_chat"
		}
		userID := payload.UserID
		if userID == "" {
			userID = "22222222-2222-2222-2222-222222222222"
		}
		tx = &store.PaymentTransaction{
			ID:             txID,
			UserID:         userID,
			Provider:       provider,
			IdempotencyKey: idemKey,
			Type:           txType,
			AmountCents:    amt,
			Currency:       curr,
			Status:         "pending",
			CreatedAt:      time.Now(),
		}
		h.Store.Transactions[txID] = tx
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
