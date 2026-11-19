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

	// P1-14: valid_view shares the playback counting guard so one session
	// increments ViewCount exactly once. Telemetry is always accepted.
	counted := false
	if event.EventType == "valid_view" {
		user := identity.GetUserFromContext(r.Context())
		viewerID := ""
		if user != nil {
			viewerID = user.ID
		}
		if v, ok := h.Store.Videos[event.VideoID]; ok && h.Store.CountView(store.ViewerKey(r, viewerID), event.VideoID) {
			v.ViewCount++
			h.Store.PersistVideo(v)
			h.Store.NoteCountedView(v)
			counted = true
		}
	}

	identity.WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":      "ACCEPTED",
		"counted":     counted,
		"received_at": time.Now().Format(time.RFC3339),
	})
}
