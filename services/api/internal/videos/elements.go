package videos

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

const maxElementsPerVideo = 20

func validElementKind(kind string) bool {
	return kind == "end_screen" || kind == "card"
}

func validElementTarget(targetType, target string) bool {
	switch targetType {
	case "video", "playlist", "channel":
		return target != ""
	case "url":
		lower := strings.ToLower(target)
		return strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "http://")
	default:
		return false
	}
}

// ListElements returns the video's end screens and cards, oldest first.
func (h *Handler) ListElements(w http.ResponseWriter, r *http.Request) {
	video, ok := h.loadOwnedVideo(w, r)
	if !ok {
		return
	}
	elements := video.Elements
	if elements == nil {
		elements = []store.VideoElement{}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"video_id": video.ID,
		"items":    elements,
		"total":    len(elements),
	})
}

// AddElement appends one end screen or card after validating its timing.
func (h *Handler) AddElement(w http.ResponseWriter, r *http.Request) {
	video, ok := h.loadOwnedVideo(w, r)
	if !ok {
		return
	}
	var payload struct {
		Kind       string `json:"kind"`
		Title      string `json:"title"`
		TargetType string `json:"target_type"`
		Target     string `json:"target"`
		StartSec   int    `json:"start_sec"`
		EndSec     int    `json:"end_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid element payload")
		return
	}
	if !validElementKind(payload.Kind) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Kind must be end_screen or card")
		return
	}
	title := security.SanitizeUserText(payload.Title, 120)
	if title == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Title is required")
		return
	}
	if !validElementTarget(payload.TargetType, payload.Target) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Target type and target are required (urls must be http(s))")
		return
	}
	if payload.StartSec < 0 || payload.EndSec <= payload.StartSec {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "End second must exceed start second from 0")
		return
	}
	if video.DurationSeconds > 0 && payload.EndSec > video.DurationSeconds {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Element extends past the video duration")
		return
	}
	if len(video.Elements) >= maxElementsPerVideo {
		identity.WriteError(w, r, http.StatusBadRequest, "QUOTA_EXCEEDED", "A video holds at most 20 end screens and cards")
		return
	}
	el := store.VideoElement{
		ID:         uuid.NewString(),
		Kind:       payload.Kind,
		Title:      title,
		TargetType: payload.TargetType,
		Target:     security.SanitizeUserText(payload.Target, 2048),
		StartSec:   payload.StartSec,
		EndSec:     payload.EndSec,
	}
	video.Elements = append(video.Elements, el)
	h.Store.PersistVideo(video)
	identity.WriteJSON(w, http.StatusCreated, el)
}

// DeleteElement removes one end screen or card by id.
func (h *Handler) DeleteElement(w http.ResponseWriter, r *http.Request) {
	video, ok := h.loadOwnedVideo(w, r)
	if !ok {
		return
	}
	elID := chi.URLParam(r, "elId")
	kept := video.Elements[:0]
	found := false
	for _, el := range video.Elements {
		if el.ID == elID {
			found = true
			continue
		}
		kept = append(kept, el)
	}
	if !found {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Element not found")
		return
	}
	video.Elements = kept
	h.Store.PersistVideo(video)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"video_id": video.ID,
		"deleted":  elID,
	})
}
