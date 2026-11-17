package notifications

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"time"
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

	r.Get("/", h.ListNotifications)
	r.Post("/{id}/read", h.MarkRead)

	return r
}

// Notify appends an inbox item for a user (newest first, capped).
func Notify(s *store.MemoryStore, userID, typ, title, body, refType, refID string) {
	if userID == "" {
		return
	}
	item := &store.Notification{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      typ,
		Title:     title,
		Body:      body,
		RefType:   refType,
		RefID:     refID,
		CreatedAt: time.Now(),
	}
	inbox := append([]*store.Notification{item}, s.Notifications[userID]...)
	if len(inbox) > 100 {
		inbox = inbox[:100]
	}
	s.Notifications[userID] = inbox
}

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	page, limit := identity.ParsePagination(r, 20)
	items := h.Store.Notifications[user.ID]
	if items == nil {
		items = []*store.Notification{}
	}
	unread := 0
	for _, n := range items {
		if !n.Read {
			unread++
		}
	}
	window, total := identity.PaginateSlice(items, page, limit)
	out := identity.PageEnvelope(window, total, page, limit)
	out["unread"] = unread
	identity.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	for _, n := range h.Store.Notifications[user.ID] {
		if n.ID == id {
			n.Read = true
			identity.WriteJSON(w, http.StatusOK, n)
			return
		}
	}
	identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Notification not found")
}
