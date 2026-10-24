package moderation

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
	// Seed demo moderation case
	mCase := &store.ModerationCase{
		ID:             "mod-case-1",
		TargetType:     "video",
		TargetID:       "00000000-0000-0000-0000-000000000001",
		Reason:         "Misleading thumbnail and copyright alert",
		PolicyCategory: "copyright",
		Status:         "pending",
		CreatedAt:      time.Now().Add(-6 * time.Hour),
	}
	s.Moderation[mCase.ID] = mCase
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.ListQueue)
	r.Post("/{id}/resolve", h.ResolveCase)

	return r
}

func (h *Handler) ListQueue(w http.ResponseWriter, r *http.Request) {
	var list []*store.ModerationCase
	for _, m := range h.Store.Moderation {
		list = append(list, m)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ResolveCase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	mCase, ok := h.Store.Moderation[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Moderation case not found")
		return
	}

	var payload struct {
		Action string `json:"action"` // approve, reject, remove
		Notes  string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	mCase.Status = "resolved"
	mCase.ActionTaken = payload.Action

	identity.WriteJSON(w, http.StatusOK, mCase)
}
