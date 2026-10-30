package settlements

import (
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

	r.Get("/settlements", h.GetSettlements)
	r.Get("/earnings", h.GetEarnings)
	r.With(auth.RequireAuth).Post("/payout", h.RequestPayout)

	return r
}

func (h *Handler) GetSettlements(w http.ResponseWriter, r *http.Request) {
	// Aggregate from Ledger and stored payouts
	var accumulatedPayable int64 = 0
	for _, entry := range h.Store.Ledger {
		if entry.AccountType == "creator_payable" && entry.Direction == "CREDIT" {
			accumulatedPayable += entry.AmountCents
		}
	}

	grossOct := int64(3120000)
	if accumulatedPayable > 0 {
		grossOct += accumulatedPayable
	}
	taxWithheldOct := (grossOct * 33) / 1000 // 3.3% tax rate

	settlements := []map[string]interface{}{
		{
			"id":          "payout_2026_09",
			"channelId":   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"periodStart": "2026-09-01",
			"periodEnd":   "2026-09-30",
			"grossAmount": 2840000,
			"taxWithheld": 93720,
			"netAmount":   2746280,
			"currency":    "KRW",
			"status":      "paid",
			"paidAt":      time.Now().Add(-15 * 24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":          "payout_2026_10",
			"channelId":   "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"periodStart": "2026-10-01",
			"periodEnd":   "2026-10-31",
			"grossAmount": grossOct,
			"taxWithheld": taxWithheldOct,
			"netAmount":   grossOct - taxWithheldOct,
			"currency":    "KRW",
			"status":      "pending",
			"paidAt":      nil,
		},
	}
	identity.WriteJSON(w, http.StatusOK, settlements)
}

func (h *Handler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	var adRev int64 = 2250000
	var memberRev int64 = 620000
	var superRev int64 = 250000

	for _, tx := range h.Store.Transactions {
		if tx.Status == "succeeded" {
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

	payoutID := "payout_" + time.Now().Format("20060102150405")
	identity.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"payout_id": payoutID,
		"status":    "scheduled",
		"message":   "Payout request submitted successfully. Processing via Shinhan Bank KRW transfer.",
	})
}
