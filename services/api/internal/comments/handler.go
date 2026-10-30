package comments

import (
	"encoding/json"
	"net/http"
	"sort"
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

	r.Get("/videos/{id}/comments", h.GetComments)
	r.With(auth.RequireAuth).Post("/videos/{id}/comments", h.AddComment)
	r.With(auth.RequireAuth).Delete("/comments/{id}", h.DeleteComment)
	r.With(auth.RequireAuth).Post("/comments/{id}/like", h.LikeComment)
	r.With(auth.RequireAuth).Post("/comments/{id}/heart", h.HeartComment)
	r.With(auth.RequireAuth).Post("/comments/{id}/pin", h.PinComment)

	return r
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	sortBy := r.URL.Query().Get("sort") // "top" or "newest"

	rawList := h.Store.Comments[videoID]
	if rawList == nil {
		rawList = []*store.Comment{}
	}

	// Separate root comments and child replies
	rootComments := []*store.Comment{}
	childReplies := make(map[string][]*store.Comment)

	for _, c := range rawList {
		if c.ParentID == "" {
			rootComments = append(rootComments, c)
		} else {
			childReplies[c.ParentID] = append(childReplies[c.ParentID], c)
		}
	}

	// Attach replies to parents
	for _, root := range rootComments {
		if replies, ok := childReplies[root.ID]; ok {
			root.Replies = replies
			root.ReplyCount = len(replies)
		} else {
			root.Replies = []*store.Comment{}
			root.ReplyCount = 0
		}
	}

	// Sort root comments
	if sortBy == "newest" {
		sort.SliceStable(rootComments, func(i, j int) bool {
			if rootComments[i].IsPinned != rootComments[j].IsPinned {
				return rootComments[i].IsPinned
			}
			return rootComments[i].CreatedAt.After(rootComments[j].CreatedAt)
		})
	} else {
		// Default to Top (likes descending, pinned first)
		sort.SliceStable(rootComments, func(i, j int) bool {
			if rootComments[i].IsPinned != rootComments[j].IsPinned {
				return rootComments[i].IsPinned
			}
			if rootComments[i].LikeCount != rootComments[j].LikeCount {
				return rootComments[i].LikeCount > rootComments[j].LikeCount
			}
			return rootComments[i].CreatedAt.After(rootComments[j].CreatedAt)
		})
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": rootComments,
		"total": len(rawList),
	})
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "id")
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}

	var payload struct {
		Content  string `json:"content"`
		ParentID string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Content == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Comment content is required")
		return
	}

	// Check if author is video channel owner
	isChannelOwner := false
	if video, ok := h.Store.Videos[videoID]; ok {
		if channel, chOk := h.Store.Channels[video.ChannelID]; chOk && channel.OwnerID == user.ID {
			isChannelOwner = true
		}
	}

	comment := &store.Comment{
		ID:             uuid.NewString(),
		VideoID:        videoID,
		UserID:         user.ID,
		DisplayName:    user.DisplayName,
		AvatarURL:      user.AvatarURL,
		Content:        payload.Content,
		LikeCount:      0,
		ReplyCount:     0,
		IsPinned:       false,
		IsHearted:      false,
		IsChannelOwner: isChannelOwner,
		ParentID:       payload.ParentID,
		Replies:        []*store.Comment{},
		CreatedAt:      time.Now(),
	}

	h.Store.Comments[videoID] = append(h.Store.Comments[videoID], comment)

	// Increment video comment count
	if video, ok := h.Store.Videos[videoID]; ok {
		video.CommentCount++
	}

	identity.WriteJSON(w, http.StatusCreated, comment)
}

func (h *Handler) LikeComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	for _, list := range h.Store.Comments {
		for _, c := range list {
			if c.ID == commentID {
				c.LikeCount++
				identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
					"id":         c.ID,
					"like_count": c.LikeCount,
				})
				return
			}
		}
	}
	identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Comment not found")
}

func (h *Handler) HeartComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	user := identity.GetUserFromContext(r.Context())

	for _, list := range h.Store.Comments {
		for _, c := range list {
			if c.ID == commentID {
				// Only channel owner or admin can heart a comment (TASK-325)
				video, vOk := h.Store.Videos[c.VideoID]
				if vOk {
					channel, chOk := h.Store.Channels[video.ChannelID]
					if chOk && (channel.OwnerID == user.ID || user.Role == "super_admin") {
						c.IsHearted = !c.IsHearted
						identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
							"id":         c.ID,
							"is_hearted": c.IsHearted,
						})
						return
					}
				}
				identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN", "Only video creator can heart comments")
				return
			}
		}
	}
	identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Comment not found")
}

func (h *Handler) PinComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	user := identity.GetUserFromContext(r.Context())

	for _, list := range h.Store.Comments {
		for _, c := range list {
			if c.ID == commentID {
				// Only channel owner or admin can pin comments (TASK-325)
				video, vOk := h.Store.Videos[c.VideoID]
				if vOk {
					channel, chOk := h.Store.Channels[video.ChannelID]
					if chOk && (channel.OwnerID == user.ID || user.Role == "super_admin") {
						nextPinState := !c.IsPinned
						// Unpin any previously pinned comment in this video
						if nextPinState {
							for _, other := range list {
								other.IsPinned = false
							}
						}
						c.IsPinned = nextPinState
						identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
							"id":        c.ID,
							"is_pinned": c.IsPinned,
						})
						return
					}
				}
				identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN", "Only video creator can pin comments")
				return
			}
		}
	}
	identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Comment not found")
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	commentID := chi.URLParam(r, "id")
	user := identity.GetUserFromContext(r.Context())

	for vid, list := range h.Store.Comments {
		for i, c := range list {
			if c.ID == commentID {
				// Only author or video creator or admin can delete
				canDelete := c.UserID == user.ID || user.Role == "super_admin"
				if !canDelete {
					if video, vOk := h.Store.Videos[vid]; vOk {
						if channel, chOk := h.Store.Channels[video.ChannelID]; chOk && channel.OwnerID == user.ID {
							canDelete = true
						}
					}
				}

				if !canDelete {
					identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN", "Cannot delete another user's comment")
					return
				}

				h.Store.Comments[vid] = append(list[:i], list[i+1:]...)
				if video, ok := h.Store.Videos[vid]; ok && video.CommentCount > 0 {
					video.CommentCount--
				}
				identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
				return
			}
		}
	}
	identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Comment not found")
}
