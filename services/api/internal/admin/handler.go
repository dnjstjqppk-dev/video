package admin

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

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(auth.RequireRole("super_admin", "platform_admin"))

	r.Get("/dashboard", h.GetDashboard)
	r.Get("/users", h.ListUsers)
	r.Patch("/users/{id}/status", h.UpdateUserStatus)
	r.Get("/audit-logs", h.GetAuditLogs)

	return r
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	totalViews := int64(0)
	for _, v := range h.Store.Videos {
		totalViews += v.ViewCount
	}

	dashboard := map[string]interface{}{
		"metrics": map[string]interface{}{
			"totalUsers":          len(h.Store.Users),
			"activeUsersToday":    1240,
			"totalVideos":         len(h.Store.Videos),
			"liveStreamsActive":   len(h.Store.LiveStreams),
			"totalAdRevenueKrw":   18450000,
			"totalViewsAllVideos": totalViews,
			"systemHealth":        "HEALTHY",
		},
		"recentAuditLogs": []map[string]interface{}{
			{
				"id":        "audit-1",
				"action":    "POLICY_UPDATE",
				"actor":     "admin_master",
				"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			{
				"id":        "audit-2",
				"action":    "PAYOUT_BATCH_APPROVED",
				"actor":     "admin_master",
				"timestamp": time.Now().Add(-4 * time.Hour).Format(time.RFC3339),
			},
		},
	}
	identity.WriteJSON(w, http.StatusOK, dashboard)
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var list []*store.User
	for _, u := range h.Store.Users {
		list = append(list, u)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

func (h *Handler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	user, ok := h.Store.Users[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	var payload struct {
		Status string `json:"status"` // active, suspended, banned
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid payload")
		return
	}

	user.IsActive = (payload.Status == "active")

	// Audit log
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"actorId":   "admin",
		"action":    "USER_STATUS_UPDATE",
		"target":    id,
		"reason":    payload.Reason,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": h.Store.AuditLogs,
		"total": len(h.Store.AuditLogs),
	})
}
