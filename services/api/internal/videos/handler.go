package videos

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

	r.Get("/{id}", h.GetVideo)
	r.Get("/{id}/playback", h.GetPlayback)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Patch("/{id}", h.UpdateVideo)
		pr.Delete("/{id}", h.DeleteVideo)
		pr.Post("/{id}/like", h.LikeVideo)
		pr.Delete("/{id}/like", h.UnlikeVideo)
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
