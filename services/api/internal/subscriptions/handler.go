package subscriptions

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

	r.Get("/", h.ListSubscriptions)
	r.Post("/{channelId}", h.Subscribe)
	r.Delete("/{channelId}", h.Unsubscribe)

	return r
}

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var subscribedChannels []*store.Channel

	if subs, ok := h.Store.Subscriptions[user.ID]; ok {
		for chID, active := range subs {
			if active {
				if ch, exists := h.Store.Channels[chID]; exists {
					subscribedChannels = append(subscribedChannels, ch)
				}
			}
		}
	}

	if subscribedChannels == nil {
		subscribedChannels = make([]*store.Channel, 0)
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": subscribedChannels,
		"total": len(subscribedChannels),
	})
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "channelId")
	channel, ok := h.Store.Channels[channelID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
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

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"channel_id":       channelID,
		"subscribed":       true,
		"subscriber_count": channel.SubscriberCount,
	})
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	channelID := chi.URLParam(r, "channelId")
	channel, ok := h.Store.Channels[channelID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if h.Store.Subscriptions[user.ID] != nil && h.Store.Subscriptions[user.ID][channelID] {
		delete(h.Store.Subscriptions[user.ID], channelID)
		if channel.SubscriberCount > 0 {
			channel.SubscriberCount--
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"channel_id":       channelID,
		"subscribed":       false,
		"subscriber_count": channel.SubscriberCount,
	})
}
