package partners

import (
	"net/http"

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

	r.Get("/me", h.GetPartnerMe)
	r.Get("/channels", h.GetChannels)
	r.Get("/contracts", h.GetContracts)

	return r
}

func (h *Handler) GetPartnerMe(w http.ResponseWriter, r *http.Request) {
	partnerOverview := map[string]interface{}{
		"organization": map[string]interface{}{
			"id":   "org-acme-media",
			"name": "Acme Media & Entertainment Group",
			"type": "advertiser_and_mcn",
		},
		"campaignCount":           len(h.Store.Campaigns),
		"activeChannelsCount":     len(h.Store.Channels),
		"pendingSettlementAmount": 5860000,
	}

	identity.WriteJSON(w, http.StatusOK, partnerOverview)
}

func (h *Handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	var list []*store.Channel
	for _, c := range h.Store.Channels {
		list = append(list, c)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

func (h *Handler) GetContracts(w http.ResponseWriter, r *http.Request) {
	contracts := []map[string]interface{}{
		{
			"id":               "ct-2026-001",
			"partnerType":      "MCN",
			"revenueShareRate": 0.85,
			"validUntil":       "2027-12-31",
			"status":           "active",
		},
	}
	identity.WriteJSON(w, http.StatusOK, contracts)
}

