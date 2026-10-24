package events

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
	return &Handler{Store: s}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/analytics", h.RecordAnalyticsEvent)
	return r
}

func (h *Handler) RecordAnalyticsEvent(w http.ResponseWriter, r *http.Request) {
	var event struct {
		EventID                string `json:"event_id"`
		EventType              string `json:"event_type"`
		SessionID              string `json:"session_id"`
		VideoID                string `json:"video_id"`
		ChannelID              string `json:"channel_id"`
		DeviceType             string `json:"device_type"`
		WatchDurationSeconds   int    `json:"watch_duration_seconds"`
		CurrentPositionSeconds int    `json:"current_position_seconds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid event payload")
		return
	}

	// Increment view count if valid view event
	if event.EventType == "valid_view" {
		if v, ok := h.Store.Videos[event.VideoID]; ok {
			v.ViewCount++
		}
	}

	identity.WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":      "ACCEPTED",
		"received_at": time.Now().Format(time.RFC3339),
	})
}

