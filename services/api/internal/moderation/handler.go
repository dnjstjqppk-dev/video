package moderation

import (
	"encoding/json"
	"net/http"
	"time"

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
	// Seed demo moderation case
	mCase := &store.ModerationCase{
		ID:             "mod-case-1",
		TargetType:     "video",
		TargetID:       "00000000-0000-0000-0000-000000000001",
		Reason:         "Misleading thumbnail and copyright alert",
		PolicyCategory: "copyright",
		Status:         "pending",
		CreatedAt:      time.Now().Add(-6 * time.Hour),
	}
	s.Moderation[mCase.ID] = mCase
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.ListQueue)
	r.Post("/{id}/resolve", h.ResolveCase)

	return r
}

func (h *Handler) ListQueue(w http.ResponseWriter, r *http.Request) {
	var list []*store.ModerationCase
	for _, m := range h.Store.Moderation {
		list = append(list, m)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ResolveCase(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	id := chi.URLParam(r, "id")
	mCase, ok := h.Store.Moderation[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Moderation case not found")
		return
	}

	var payload struct {
		Action   string `json:"action"` // approve, reject, remove
		Notes    string `json:"notes"`
		TOTPCode string `json:"totpCode"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	// P0-7: sensitive Trust & Safety decisions require MFA step-up.
	if !identity.RequireStepUp(w, r, user, payload.TOTPCode) {
		return
	}
	switch payload.Action {
	case "approve", "reject", "remove":
	default:
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "action must be approve, reject, or remove")
		return
	}

	mCase.Status = "resolved"
	mCase.ActionTaken = payload.Action

	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   actorID(user),
		"action":    "RESOLVE_MODERATION_CASE",
		"target":    id,
		"reason":    security.SanitizeUserText(payload.Notes, 500),
		"ip":        r.RemoteAddr,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, mCase)
}

func actorID(user *store.User) string {
	if user == nil {
		return "unknown"
	}
	return user.ID
}

// ReportContent files a user report against a video, comment, or channel.
// The case lands in the admin moderation queue for human review.
func (h *Handler) ReportContent(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	var payload struct {
		TargetType     string `json:"targetType"`
		TargetID       string `json:"targetId"`
		Reason         string `json:"reason"`
		PolicyCategory string `json:"policyCategory"`
		Details        string `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	targetType := security.SanitizeUserText(payload.TargetType, 32)
	targetID := security.SanitizeUserText(payload.TargetID, 128)
	reason := security.SanitizeUserText(payload.Reason, 2000)
	if targetType == "" || targetID == "" || reason == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "targetType, targetId, and reason are required")
		return
	}
	switch targetType {
	case "video", "comment", "channel":
	default:
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "targetType must be video, comment, or channel")
		return
	}
	mCase := &store.ModerationCase{
		ID:             "mod-" + uuid.NewString()[:8],
		TargetType:     targetType,
		TargetID:       targetID,
		Reason:         reason,
		PolicyCategory: security.SanitizeUserText(payload.PolicyCategory, 64),
		Status:         "pending",
		CreatedAt:      time.Now(),
	}
	if mCase.PolicyCategory == "" {
		mCase.PolicyCategory = "general"
	}
	if details := security.SanitizeUserText(payload.Details, 2000); details != "" {
		mCase.Reason = reason + " | details: " + details
	}
	h.Store.Moderation[mCase.ID] = mCase
	identity.WriteJSON(w, http.StatusCreated, mCase)
}
