package playlists

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
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/{id}", h.GetPlaylist)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/", h.CreatePlaylist)
		pr.Patch("/{id}", h.UpdatePlaylist)
		pr.Post("/{id}/items", h.AddItem)
	})

	return r
}

func (h *Handler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		ChannelID   string `json:"channel_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Title == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Playlist title is required")
		return
	}

	visibility := payload.Visibility
	if visibility == "" {
		visibility = "public"
	}

	now := time.Now()
	pl := &store.Playlist{
		ID:          uuid.NewString(),
		OwnerID:     user.ID,
		ChannelID:   payload.ChannelID,
		Title:       payload.Title,
		Description: payload.Description,
		Visibility:  visibility,
		Items:       make([]store.PlaylistItem, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.Store.Playlists[pl.ID] = pl
	identity.WriteJSON(w, http.StatusCreated, pl)
}

func (h *Handler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pl, ok := h.Store.Playlists[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Playlist not found")
		return
	}

	if pl.Visibility == "private" {
		user := identity.GetUserFromContext(r.Context())
		if user == nil || (user.ID != pl.OwnerID && user.Role != "super_admin") {
			identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Private playlist")
			return
		}
	}

	identity.WriteJSON(w, http.StatusOK, pl)
}

func (h *Handler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pl, ok := h.Store.Playlists[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Playlist not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if pl.OwnerID != user.ID && user.Role != "super_admin" {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another user's playlist")
		return
	}

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	if payload.Title != "" {
		pl.Title = payload.Title
	}
	if payload.Description != "" {
		pl.Description = payload.Description
	}
	if payload.Visibility != "" {
		pl.Visibility = payload.Visibility
	}
	pl.UpdatedAt = time.Now()

	identity.WriteJSON(w, http.StatusOK, pl)
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	pl, ok := h.Store.Playlists[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Playlist not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if pl.OwnerID != user.ID && user.Role != "super_admin" {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another user's playlist")
		return
	}

	var payload struct {
		VideoID string `json:"video_id"`
		VideoId string `json:"videoId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	vid := payload.VideoID
	if vid == "" {
		vid = payload.VideoId
	}
	if vid == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "video_id is required")
		return
	}

	item := store.PlaylistItem{
		VideoID:  vid,
		Position: len(pl.Items) + 1,
		AddedAt:  time.Now(),
	}
	pl.Items = append(pl.Items, item)
	pl.UpdatedAt = time.Now()

	identity.WriteJSON(w, http.StatusCreated, pl)
}
