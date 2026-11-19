package channels

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

	r.Get("/{id}", h.GetChannel)
	r.Get("/{id}/videos", h.GetChannelVideos)

	// Protected routes
	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/", h.CreateChannel)
		pr.Patch("/{id}", h.UpdateChannel)
		pr.Get("/{id}/subscription", h.GetSubscription)
		pr.Post("/{id}/subscribe", h.Subscribe)
		pr.Patch("/{id}/subscribe", h.UpdateSubscription)
		pr.Delete("/{id}/subscribe", h.Unsubscribe)
	})

	return r
}

func (h *Handler) GetChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	channel, ok := h.Store.Channels[id]
	if !ok {
		// Try match by handle
		for _, c := range h.Store.Channels {
			if c.Handle == id || c.Handle == "@"+id {
				channel = c
				ok = true
				break
			}
		}
	}

	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}

	identity.WriteJSON(w, http.StatusOK, channel)
}

func (h *Handler) GetChannelVideos(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")
	page, limit := identity.ParsePagination(r, 20)
	var videos []*store.Video
	for _, v := range h.Store.Videos {
		if v.ChannelID == channelID {
			videos = append(videos, v)
		}
	}
	window, total := identity.PaginateSlice(videos, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var payload struct {
		Name        string `json:"name"`
		Handle      string `json:"handle"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	ch := &store.Channel{
		ID:                  uuid.NewString(),
		OwnerID:             user.ID,
		Name:                payload.Name,
		Handle:              payload.Handle,
		Description:         payload.Description,
		AvatarURL:           "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=150",
		SubscriberCount:     0,
		VideoCount:          0,
		IsVerified:          false,
		MonetizationEnabled: false,
		CreatedAt:           time.Now(),
	}

	h.Store.Channels[ch.ID] = ch
	h.Store.PersistChannel(ch)
	identity.WriteJSON(w, http.StatusCreated, ch)
}

func (h *Handler) UpdateChannel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	channel, ok := h.Store.Channels[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if channel.OwnerID != user.ID && user.Role != "super_admin" {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify channel of another user")
		return
	}

	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AvatarURL   string `json:"avatarUrl"`
		BannerURL   string `json:"bannerUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
		if payload.Name != "" {
			channel.Name = security.SanitizeUserText(payload.Name, 120)
		}
		if payload.Description != "" {
			channel.Description = security.SanitizeUserText(payload.Description, 2000)
		}
		if payload.AvatarURL != "" {
			channel.AvatarURL = security.SanitizeUserText(payload.AvatarURL, 2048)
		}
		if payload.BannerURL != "" {
			channel.BannerURL = security.SanitizeUserText(payload.BannerURL, 2048)
		}
	}

	h.Store.PersistChannel(channel)
	identity.WriteJSON(w, http.StatusOK, channel)
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")
	channel, ok := h.Store.Channels[channelID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}
	user := identity.GetUserFromContext(r.Context())
	subscribed := h.Store.Subscriptions[user.ID] != nil && h.Store.Subscriptions[user.ID][channelID]
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"channel_id":       channelID,
		"subscribed":       subscribed,
		"notify_level":     h.Store.SubNotifyLevel(user.ID, channelID),
		"subscriber_count": channel.SubscriberCount,
	})
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")
	channel, ok := h.Store.Channels[channelID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}

	var payload struct {
		NotifyLevel string `json:"notify_level"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
			return
		}
	}
	if payload.NotifyLevel != "" && !store.ValidNotifyLevel(payload.NotifyLevel) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "notify_level must be all, custom, or none")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if h.Store.Subscriptions[user.ID] == nil {
		h.Store.Subscriptions[user.ID] = make(map[string]bool)
	}

	if !h.Store.Subscriptions[user.ID][channelID] {
		h.Store.Subscriptions[user.ID][channelID] = true
		channel.SubscriberCount++
	}
	level := payload.NotifyLevel
	if level == "" {
		level = h.Store.SubNotifyLevel(user.ID, channelID)
	}
	h.Store.SetSubNotify(user.ID, channelID, level)

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"subscribed":       true,
		"notify_level":     level,
		"subscriber_count": channel.SubscriberCount,
	})
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")
	channel, ok := h.Store.Channels[channelID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}
	var payload struct {
		NotifyLevel string `json:"notify_level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || !store.ValidNotifyLevel(payload.NotifyLevel) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "notify_level must be all, custom, or none")
		return
	}
	user := identity.GetUserFromContext(r.Context())
	if h.Store.Subscriptions[user.ID] == nil || !h.Store.Subscriptions[user.ID][channelID] {
		identity.WriteError(w, r, http.StatusConflict, "INVALID_STATE", "Subscribe before changing notifications")
		return
	}
	h.Store.SetSubNotify(user.ID, channelID, payload.NotifyLevel)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"channel_id":       channelID,
		"subscribed":       true,
		"notify_level":     payload.NotifyLevel,
		"subscriber_count": channel.SubscriberCount,
	})
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "id")
	channel, ok := h.Store.Channels[channelID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if h.Store.Subscriptions[user.ID] != nil && h.Store.Subscriptions[user.ID][channelID] {
		delete(h.Store.Subscriptions[user.ID], channelID)
		h.Store.ClearSubNotify(user.ID, channelID)
		if channel.SubscriberCount > 0 {
			channel.SubscriberCount--
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"subscribed":       false,
		"notify_level":     "",
		"subscriber_count": channel.SubscriberCount,
	})
}
