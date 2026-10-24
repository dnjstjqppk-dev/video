package users

import (
	"encoding/json"
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

	r.Get("/me", h.GetMe)
	r.Patch("/me", h.UpdateMe)
	r.Get("/me/sessions", h.GetSessions)
	r.Get("/me/history", h.GetHistory)
	r.Delete("/me/history", h.ClearHistory)

	return r
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}
	identity.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	var payload struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	if payload.DisplayName != "" {
		user.DisplayName = payload.DisplayName
	}
	if payload.Bio != "" {
		user.Bio = payload.Bio
	}

	identity.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []map[string]interface{}{
		{
			"id":          "sess_current",
			"ip":          "127.0.0.1",
			"user_agent":  r.UserAgent(),
			"last_active": "Just now",
			"is_current":  true,
		},
	}
	identity.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	var list []*store.Video
	for _, v := range h.Store.Videos {
		list = append(list, v)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

func (h *Handler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

