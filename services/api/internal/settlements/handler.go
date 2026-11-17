package settlements

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

type Handler struct {
	Store *store.MemoryStore
}

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(auth.RequireAuth)

	r.Get("/settlements", h.GetSettlements)
	r.Get("/earnings", h.GetEarnings)
	r.Post("/payout", h.RequestPayout)

	return r
}

// financeSeesAll reports whether the user may view platform-wide money.
func financeSeesAll(user *store.User) bool {
	if user == nil {
		return false
	}
	switch user.Role {
	case "super_admin", "platform_admin", "finance_admin":
		return true
	default:
		return false
	}
}

// ownTransactions returns succeeded transactions attributable to the caller.
// Finance operators see everything; everyone else sees only their own.
func (h *Handler) ownTransactions(user *store.User) []*store.PaymentTransaction {
	all := financeSeesAll(user)
	var out []*store.PaymentTransaction
	for _, tx := range h.Store.Transactions {
		if tx.Status != "succeeded" {
			continue
		}
		if all || (user != nil && tx.UserID == user.ID) {
			out = append(out, tx)
		}
	}
	return out
}

func (h *Handler) GetSettlements(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	// P1-18: aggregate only the caller's payable ledger (no fake base).
	mine := make(map[string]bool)
	for _, tx := range h.ownTransactions(user) {
		mine[tx.ID] = true
	}
	byMonth := make(map[string]int64)
	for _, entry := range h.Store.Ledger {
		if entry.AccountType != "creator_payable" || entry.Direction != "CREDIT" {
			continue
		}
		if !mine[entry.TransactionID] {
			continue
		}
		month := entry.CreatedAt.UTC().Format("2006-01")
		byMonth[month] += entry.AmountCents
	}

	channelID := ""
	if ch := h.Store.ChannelForUser(user.ID); ch != nil {
		channelID = ch.ID
	}
	settlements := make([]map[string]interface{}, 0, len(byMonth))
	for month, gross := range byMonth {
		tax := (gross * 33) / 1000 // 3.3% withholding
		settlements = append(settlements, map[string]interface{}{
			"id":          "payout_" + month,
			"channelId":   channelID,
			"periodStart": month + "-01",
			"periodEnd":   month + "-31",
			"grossAmount": gross,
			"taxWithheld": tax,
			"netAmount":   gross - tax,
			"currency":    "KRW",
			"status":      "accrued",
			"paidAt":      nil,
		})
	}
	identity.WriteJSON(w, http.StatusOK, settlements)
}

func (h *Handler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	var adRev, memberRev, superRev int64

	for _, tx := range h.ownTransactions(user) {
		creatorShare := (tx.AmountCents / 100) * 70
		switch tx.Type {
		case "super_chat":
			superRev += creatorShare
		case "membership":
			memberRev += creatorShare
		case "ad_topup":
			adRev += (tx.AmountCents / 100) * 55
		}
	}

	total := adRev + memberRev + superRev

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"currentMonthEstimated": total,
		"adRevenue":             adRev,
		"membershipRevenue":     memberRev,
		"superChatRevenue":      superRev,
		"currency":              "KRW",
	})
}

func (h *Handler) RequestPayout(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}

	var payload struct {
		ChannelID string `json:"channelId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	if payload.ChannelID != "" {
		channel, ok := h.Store.Channels[payload.ChannelID]
		if !ok {
			identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
			return
		}
		isFinanceOperator := user.Role == "super_admin" || user.Role == "platform_admin" || user.Role == "finance_admin"
		if channel.OwnerID != user.ID && !isFinanceOperator {
			identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot request payouts for another creator's channel")
			return
		}
	} else if h.Store.ChannelForUser(user.ID) == nil && !financeSeesAll(user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Payout requests require a channel")
		return
	}

	payoutID := "payout_" + time.Now().Format("20060102150405")
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        payoutID,
		"actorId":   user.ID,
		"action":    "REQUEST_PAYOUT",
		"target":    payload.ChannelID,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	identity.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"payout_id": payoutID,
		"status":    "scheduled",
		"message":   "Payout request submitted successfully. Processing via Shinhan Bank KRW transfer.",
	})
}
