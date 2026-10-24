package comments

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

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	comments := h.Store.Comments[videoID]
	if comments == nil {
		comments = []*store.Comment{}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": comments,
		"total": len(comments),
	})
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	user := identity.GetUserFromContext(r.Context())

	var payload struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Content == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Comment content is required")
		return
	}

	comment := &store.Comment{
		ID:          uuid.NewString(),
		VideoID:     videoID,
		UserID:      user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Content:     payload.Content,
		LikeCount:   0,
		ReplyCount:  0,
		IsPinned:    false,
		CreatedAt:   time.Now(),
	}

	h.Store.Comments[videoID] = append([]*store.Comment{comment}, h.Store.Comments[videoID]...)

	// Increment video comment count
	if video, ok := h.Store.Videos[videoID]; ok {
		video.CommentCount++
	}

	identity.WriteJSON(w, http.StatusCreated, comment)
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	for vid, list := range h.Store.Comments {
		for i, c := range list {
			if c.ID == commentID {
				h.Store.Comments[vid] = append(list[:i], list[i+1:]...)
				identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
				return
			}
		}
	}
	identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Comment not found")
}

