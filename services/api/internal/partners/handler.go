package partners

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
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
	r.Patch("/contracts/{id}", h.RevokeContract)

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
	contracts := make([]*store.PartnerContract, 0, len(h.Store.Contracts))
	for _, c := range h.Store.Contracts {
		contracts = append(contracts, c)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": contracts,
		"total": len(contracts),
	})
}

func (h *Handler) RevokeContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	contract, ok := h.Store.Contracts[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Contract not found")
		return
	}
	var payload struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	reason := security.SanitizeUserText(payload.Reason, 500)
	if reason == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Revocation reason is required")
		return
	}
	if contract.Status == "revoked" {
		identity.WriteError(w, r, http.StatusConflict, "ALREADY_REVOKED", "Contract is already revoked")
		return
	}
	now := time.Now()
	contract.Status = "revoked"
	contract.RevokedAt = &now
	identity.WriteJSON(w, http.StatusOK, contract)
}
