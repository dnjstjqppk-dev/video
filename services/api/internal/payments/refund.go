package payments

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

// RefundPayment refunds a succeeded transaction in full through its provider
// and posts reversing ledger entries so the books stay balanced (P1-17).
func (h *Handler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	orderID := chi.URLParam(r, "orderId")

	var payload struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	h.mu.Lock()
	defer h.mu.Unlock()

	tx, exists := h.Store.Transactions[orderID]
	if !exists {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Transaction not found for orderId")
		return
	}
	if tx.UserID != user.ID && user.Role != "super_admin" && user.Role != "platform_admin" && user.Role != "finance_admin" {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN", "Unauthorized to refund this transaction")
		return
	}
	if tx.Status == "refunded" {
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"idempotent_replay": true,
			"transaction":       tx,
		})
		return
	}
	if tx.Status != "succeeded" {
		identity.WriteError(w, r, http.StatusConflict, "INVALID_STATE", "Only succeeded transactions can be refunded")
		return
	}
	if tx.ProviderRef == "" {
		identity.WriteError(w, r, http.StatusConflict, "MISSING_PROVIDER_REF", "Transaction has no provider reference to refund")
		return
	}

	var refundID string
	switch adapter := h.Adapters[tx.Provider].(type) {
	case *TossPaymentsAdapter:
		if adapter == nil || adapter.getSecretKey() == "" {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Payment provider is not configured")
			return
		}
		id, err := adapter.CancelPayment(tx.ProviderRef, payload.Reason)
		if err != nil {
			identity.WriteError(w, r, http.StatusPaymentRequired, "REFUND_FAILED", fmt.Sprintf("Toss cancel failed: %v", err))
			return
		}
		refundID = id
	case *StripeAdapter:
		if adapter == nil || adapter.getAPIKey() == "" {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "UNAVAILABLE", "Payment provider is not configured")
			return
		}
		id, err := adapter.RefundSession(tx.ProviderRef, payload.Reason)
		if err != nil {
			identity.WriteError(w, r, http.StatusPaymentRequired, "REFUND_FAILED", fmt.Sprintf("Stripe refund failed: %v", err))
			return
		}
		refundID = id
	default:
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unsupported payment provider")
		return
	}

	now := time.Now()
	tx.Status = "refunded"
	tx.RefundedCents = tx.AmountCents
	tx.ProviderRefund = refundID
	tx.RefundedAt = &now

	reversals := make([]store.LedgerEntry, 0, len(tx.LedgerEntries))
	for _, e := range tx.LedgerEntries {
		direction := "DEBIT"
		if e.Direction == "DEBIT" {
			direction = "CREDIT"
		}
		reversals = append(reversals, store.LedgerEntry{
			ID:            uuid.NewString(),
			TransactionID: tx.ID,
			AccountType:   e.AccountType,
			Direction:     direction,
			AmountCents:   e.AmountCents,
			Currency:      e.Currency,
			CreatedAt:     now,
		})
	}
	tx.LedgerEntries = append(tx.LedgerEntries, reversals...)
	h.Store.Ledger = append(h.Store.Ledger, reversals...)
	h.Store.PersistTransaction(tx)

	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "REFUND_PAYMENT",
		"target":    tx.ID,
		"reason":    payload.Reason,
		"ip":        r.RemoteAddr,
		"timestamp": now.Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"transaction": tx,
		"refund_id":   refundID,
	})
}
