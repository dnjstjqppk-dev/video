package copyright

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
	// Seed demo claim
	claim := &store.CopyrightClaim{
		ID:                "claim-1",
		ClaimedVideoID:    "00000000-0000-0000-0000-000000000001",
		OwnerName:         "Sony Music Entertainment",
		ClaimAction:       "monetize",
		Status:            "active",
		MatchStartSeconds: 124,
		MatchEndSeconds:   152,
		CreatedAt:         time.Now().Add(-48 * time.Hour),
	}
	s.Claims[claim.ID] = claim
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/claims", h.ListClaims)
	r.Post("/claims", h.CreateClaim)

	return r
}

func (h *Handler) ListClaims(w http.ResponseWriter, r *http.Request) {
	var list []*store.CopyrightClaim
	for _, c := range h.Store.Claims {
		list = append(list, c)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		VideoID string `json:"videoId"`
		AssetID string `json:"assetId"`
		Action  string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid copyright claim")
		return
	}

	claim := &store.CopyrightClaim{
		ID:                uuid.NewString(),
		ClaimedVideoID:    payload.VideoID,
		OwnerName:         "Independent Content Rights Group",
		ClaimAction:       payload.Action,
		Status:            "pending",
		MatchStartSeconds: 0,
		MatchEndSeconds:   30,
		CreatedAt:         time.Now(),
	}

	h.Store.Claims[claim.ID] = claim
	identity.WriteJSON(w, http.StatusCreated, claim)
}
