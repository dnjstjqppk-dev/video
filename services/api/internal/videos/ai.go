package videos

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
)

// SubmitAIReview queues an AI-made video for Originals review (owner only).
// Review gates the Originals shelf only; the video stays visible under its
// normal visibility rule while pending.
func (h *Handler) SubmitAIReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	if !h.canManageVideo(r, video) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot submit another user's video")
		return
	}
	if !video.AiGenerated {
		identity.WriteError(w, r, http.StatusBadRequest, "NOT_AI_VIDEO", "Only AI-generated videos can enter Originals review")
		return
	}
	if video.Status != "READY" {
		identity.WriteError(w, r, http.StatusConflict, "NOT_READY", "Only READY videos can enter Originals review")
		return
	}
	switch video.AiReviewStatus {
	case "pending":
		identity.WriteError(w, r, http.StatusConflict, "ALREADY_PENDING", "This video is already awaiting review")
		return
	case "approved":
		identity.WriteError(w, r, http.StatusConflict, "ALREADY_APPROVED", "This video is already an AI Original")
		return
	}
	video.AiReviewStatus = "pending"
	video.AiReviewNote = ""
	h.Store.PersistVideo(video)
	identity.WriteJSON(w, http.StatusOK, video)
}

// ReviewAI approves or rejects a pending AI video (trust & safety roles).
// Approving an episode also flags its parent series as an AI Original.
func (h *Handler) ReviewAI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	var payload struct {
		Decision string `json:"decision"` // approved, rejected
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	if payload.Decision != "approved" && payload.Decision != "rejected" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "decision must be approved or rejected")
		return
	}
	if video.AiReviewStatus != "pending" {
		identity.WriteError(w, r, http.StatusConflict, "NOT_PENDING", "Only pending videos can be reviewed")
		return
	}
	video.AiReviewStatus = payload.Decision
	video.AiReviewNote = security.SanitizeUserText(payload.Note, 2000)
	h.Store.PersistVideo(video)
	h.Store.RefreshSeriesOriginalFlag(video.SeriesID)
	identity.WriteJSON(w, http.StatusOK, video)
}
