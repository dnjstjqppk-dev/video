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

	return r
}

func (h *Handler) GetSettlements(w http.ResponseWriter, r *http.Request) {
	settlements := []map[string]interface{}{
		{
			"id":           "payout_2026_09",
			"channelId":    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"periodStart":  "2026-09-01",
			"periodEnd":    "2026-09-30",
			"grossAmount":  2840000,
			"taxWithheld":  93720,
			"netAmount":    2746280,
			"currency":     "KRW",
			"status":       "paid",
			"paidAt":       time.Now().Add(-15 * 24 * time.Hour).Format(time.RFC3339),
		},
		{
			"id":           "payout_2026_10",
			"channelId":    "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			"periodStart":  "2026-10-01",
			"periodEnd":    "2026-10-31",
			"grossAmount":  3120000,
			"taxWithheld":  102960,
			"netAmount":    3017040,
			"currency":     "KRW",
			"status":       "pending",
			"paidAt":       nil,
		},
	}
	identity.WriteJSON(w, http.StatusOK, settlements)
}

func (h *Handler) GetEarnings(w http.ResponseWriter, r *http.Request) {
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"currentMonthEstimated": 3120000,
		"adRevenue":             2250000,
		"membershipRevenue":     620000,
		"superChatRevenue":      250000,
		"currency":              "KRW",
	})
}

