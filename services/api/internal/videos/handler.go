package videos

import (
	"encoding/json"
	"net/http"

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

	r.Get("/{id}", h.GetVideo)
	r.Get("/{id}/playback", h.GetPlayback)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Patch("/{id}", h.UpdateVideo)
		pr.Delete("/{id}", h.DeleteVideo)
		pr.Post("/{id}/like", h.LikeVideo)
		pr.Delete("/{id}/like", h.UnlikeVideo)
		pr.Get("/{id}/captions", h.ListCaptions)
		pr.Post("/{id}/captions", h.AddCaption)
		pr.Delete("/{id}/captions/{capId}", h.DeleteCaption)
	})

	return r
}

func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}

	identity.WriteJSON(w, http.StatusOK, video)
}

func (h *Handler) GetPlayback(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}

	// Increment view count
	video.ViewCount++

	playback := map[string]interface{}{
		"video_id":        video.ID,
		"title":           video.Title,
		"hls_master_url":  video.HlsMasterURL,
		"duration":        video.DurationSeconds,
		"sources":         video.Sources,
		"chapters":        video.Chapters,
		"captions":        video.Captions,
		"playback_token":  "pb_token_" + video.ID,
		"ad_break_points": []int{180, 420}, // Ad cues in seconds
		"is_members_only": video.Visibility == "members_only",
		"has_access":      true,
	}

	identity.WriteJSON(w, http.StatusOK, playback)
}

func (h *Handler) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}

	if !h.canManageVideo(r, video) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify video of another user")
		return
	}

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Visibility  string `json:"visibility"`
		Category    string `json:"category"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err == nil {
		if payload.Title != "" {
			video.Title = payload.Title
		}
		if payload.Description != "" {
			video.Description = payload.Description
		}
		if payload.Visibility != "" {
			video.Visibility = payload.Visibility
		}
		if payload.Category != "" {
			video.Category = payload.Category
		}
	}
	h.Store.PersistVideo(video)

	identity.WriteJSON(w, http.StatusOK, video)
}

func (h *Handler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}

	if !h.canManageVideo(r, video) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify video of another user")
		return
	}

	video.Status = "DELETED"
	identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) canManageVideo(r *http.Request, video *store.Video) bool {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		return false
	}
	if user.Role == "super_admin" {
		return true
	}
	channel, ok := h.Store.Channels[video.ChannelID]
	return ok && channel != nil && channel.OwnerID == user.ID
}

func (h *Handler) LikeVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if h.Store.Likes[user.ID] == nil {
		h.Store.Likes[user.ID] = make(map[string]bool)
	}

	if !h.Store.Likes[user.ID][id] {
		h.Store.Likes[user.ID][id] = true
		video.LikeCount++
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"liked":      true,
		"like_count": video.LikeCount,
	})
}

func (h *Handler) UnlikeVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if h.Store.Likes[user.ID] != nil && h.Store.Likes[user.ID][id] {
		delete(h.Store.Likes[user.ID], id)
		if video.LikeCount > 0 {
			video.LikeCount--
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"liked":      false,
		"like_count": video.LikeCount,
	})
}

func (h *Handler) ListCaptions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	captions := video.Captions
	if captions == nil {
		captions = []store.VideoCaption{}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": captions,
		"total": len(captions),
	})
}

func (h *Handler) AddCaption(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	if !h.canManageVideo(r, video) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify video of another user")
		return
	}
	var payload struct {
		Language string `json:"language"`
		Label    string `json:"label"`
		VttURL   string `json:"vttUrl"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	language := security.SanitizeUserText(payload.Language, 16)
	vttURL := security.SanitizeUserText(payload.VttURL, 2048)
	if language == "" || vttURL == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "language and vttUrl are required")
		return
	}
	cap := store.VideoCaption{
		ID:       "cap-" + uuid.NewString()[:8],
		Language: language,
		Label:    security.SanitizeUserText(payload.Label, 120),
		VttURL:   vttURL,
	}
	if cap.Label == "" {
		cap.Label = language
	}
	video.Captions = append(video.Captions, cap)
	identity.WriteJSON(w, http.StatusCreated, cap)
}

func (h *Handler) DeleteCaption(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	capID := chi.URLParam(r, "capId")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	if !h.canManageVideo(r, video) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify video of another user")
		return
	}
	kept := make([]store.VideoCaption, 0, len(video.Captions))
	found := false
	for _, c := range video.Captions {
		if c.ID == capID {
			found = true
			continue
		}
		kept = append(kept, c)
	}
	if !found {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Caption track not found")
		return
	}
	video.Captions = kept
	w.WriteHeader(http.StatusNoContent)
}
