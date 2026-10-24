package ads

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

type Handler struct {
	Store *store.MemoryStore
}

func NewHandler(s *store.MemoryStore) *Handler {
	// Seed demo campaign
	c := &store.AdCampaign{
		ID:               "camp-1",
		AdvertiserID:     "55555555-5555-5555-5555-555555555555",
		Name:             "Q4 Global Tech Cloud Infrastructure Campaign",
		Format:           "preroll",
		Status:           "active",
		TotalBudgetCents: 500000000, // 5,000,000 KRW
		DailyBudgetCents: 20000000,  // 200,000 KRW
		SpentCents:       12450000,
		Impressions:      142800,
		Clicks:           3890,
		CreatedAt:        time.Now().Add(-14 * 24 * time.Hour),
	}
	s.Campaigns[c.ID] = c
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/campaigns", h.ListCampaigns)
	r.Post("/campaigns", h.CreateCampaign)

	return r
}

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	var list []*store.AdCampaign
	for _, c := range h.Store.Campaigns {
		list = append(list, c)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name             string `json:"name"`
		Format           string `json:"format"`
		TotalBudgetCents int64  `json:"totalBudget"`
		DailyBudgetCents int64  `json:"dailyBudget"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid campaign data")
		return
	}

	campaign := &store.AdCampaign{
		ID:               uuid.NewString(),
		AdvertiserID:     "55555555-5555-5555-5555-555555555555",
		Name:             payload.Name,
		Format:           payload.Format,
		Status:           "pending_review",
		TotalBudgetCents: payload.TotalBudgetCents,
		DailyBudgetCents: payload.DailyBudgetCents,
		SpentCents:       0,
		Impressions:      0,
		Clicks:           0,
		CreatedAt:        time.Now(),
	}

	h.Store.Campaigns[campaign.ID] = campaign
	identity.WriteJSON(w, http.StatusCreated, campaign)
}

