package users

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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
	r.Use(auth.RequireAuth)

	// User Profile
	r.Get("/me", h.GetMe)
	r.Patch("/me", h.UpdateMe)

	// Sessions (TASK-059)
	r.Get("/me/sessions", h.GetSessions)
	r.Post("/me/sessions/{id}/revoke", h.RevokeSession)

	// History
	r.Get("/me/history", h.GetHistory)
	r.Delete("/me/history", h.ClearHistory)

	// MFA & TOTP (TASK-008)
	r.Get("/me/mfa", h.GetMFAStatus)
	r.Post("/me/mfa/setup", h.SetupMFA)
	r.Post("/me/mfa/verify", h.VerifyMFA)
	r.Post("/me/mfa/disable", h.DisableMFA)

	// Blocked Users (TASK-070)
	r.Get("/me/blocked-users", h.GetBlockedUsers)
	r.Post("/me/blocked-users", h.BlockUser)
	r.Delete("/me/blocked-users/{targetId}", h.UnblockUser)

	// 3-Strikes System (TASK-075, TASK-076)
	r.Get("/me/strikes", h.GetStrikes)

	// GDPR Account Deletion (TASK-021)
	r.Post("/me/delete-request", h.RequestAccountDeletion)
	r.Delete("/me/delete-request", h.CancelAccountDeletion)

	return r
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}
	identity.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	var payload struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}

	if payload.DisplayName != "" {
		user.DisplayName = payload.DisplayName
	}
	if payload.Bio != "" {
		user.Bio = payload.Bio
	}

	identity.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) GetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := []map[string]interface{}{
		{
			"id":          "sess_current",
			"ip":          "127.0.0.1",
			"user_agent":  r.UserAgent(),
			"last_active": "Just now",
			"is_current":  true,
		},
	}
	identity.WriteJSON(w, http.StatusOK, sessions)
}

func (h *Handler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	if sessionID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Session ID required")
		return
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "revoked",
		"session_id": sessionID,
	})
}

func (h *Handler) GetHistory(w http.ResponseWriter, r *http.Request) {
	var list []*store.Video
	for _, v := range h.Store.Videos {
		list = append(list, v)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

func (h *Handler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}

func (h *Handler) GetMFAStatus(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"mfa_enabled":     user.MFAEnabled,
		"has_backup_keys": len(user.BackupCodes) > 0,
	})
}

func (h *Handler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	secret, err := security.GenerateTOTPSecret()
	if err != nil {
		identity.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate TOTP secret")
		return
	}

	plainCodes, hashedCodes, err := security.GenerateBackupCodes(8)
	if err != nil {
		identity.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to generate backup codes")
		return
	}

	// Temporarily stage secret pending verification
	user.TOTPSecret = secret
	user.BackupCodes = hashedCodes

	otpauthURI := security.FormatTOTPURI(secret, user.Email, "GlobalVideoPlatform")

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"secret":       secret,
		"otpauth_uri":  otpauthURI,
		"backup_codes": plainCodes,
		"issuer":       "GlobalVideoPlatform",
	})
}

func (h *Handler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	var payload struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Code == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "TOTP code required")
		return
	}

	if user.TOTPSecret == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_STATE", "MFA setup has not been initiated")
		return
	}

	if !security.ValidateTOTP(user.TOTPSecret, payload.Code) {
		identity.WriteError(w, r, http.StatusUnauthorized, "INVALID_CODE", "Invalid or expired 6-digit TOTP code")
		return
	}

	user.MFAEnabled = true
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "enabled",
		"mfa_enabled": true,
	})
}

func (h *Handler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	var payload struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Code == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "TOTP verification code required")
		return
	}

	// Validate TOTP or backup code before disabling
	valid := security.ValidateTOTP(user.TOTPSecret, payload.Code)
	if !valid {
		valid, user.BackupCodes = security.ValidateAndBurnBackupCode(payload.Code, user.BackupCodes)
	}

	if !valid {
		identity.WriteError(w, r, http.StatusUnauthorized, "INVALID_CODE", "Verification code invalid")
		return
	}

	user.MFAEnabled = false
	user.TOTPSecret = ""
	user.BackupCodes = nil

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":      "disabled",
		"mfa_enabled": false,
	})
}

func (h *Handler) GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	blocked := user.BlockedUsers
	if blocked == nil {
		blocked = []string{}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"blocked_users": blocked,
		"count":         len(blocked),
	})
}

func (h *Handler) BlockUser(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	var payload struct {
		TargetUserID string `json:"target_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.TargetUserID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Target user ID required")
		return
	}

	if payload.TargetUserID == user.ID {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Cannot block yourself")
		return
	}

	// Prevent duplicates
	for _, b := range user.BlockedUsers {
		if b == payload.TargetUserID {
			identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"status":        "already_blocked",
				"blocked_users": user.BlockedUsers,
			})
			return
		}
	}

	user.BlockedUsers = append(user.BlockedUsers, payload.TargetUserID)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "blocked",
		"blocked_users": user.BlockedUsers,
	})
}

func (h *Handler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	targetID := chi.URLParam(r, "targetId")
	if targetID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Target ID required")
		return
	}

	var newBlocked []string
	for _, b := range user.BlockedUsers {
		if b != targetID {
			newBlocked = append(newBlocked, b)
		}
	}
	user.BlockedUsers = newBlocked

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "unblocked",
		"blocked_users": user.BlockedUsers,
	})
}

func (h *Handler) GetStrikes(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	isSuspended := user.CopyrightStrikes >= 3 || user.CommunityStrikes >= 3

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"copyright_strikes":  user.CopyrightStrikes,
		"community_strikes":  user.CommunityStrikes,
		"max_strikes":        3,
		"is_suspended":       isSuspended,
		"account_status":     user.IsActive,
		"can_upload":         !isSuspended && user.IsActive,
		"can_live_stream":    !isSuspended && user.IsActive && user.CommunityStrikes == 0,
	})
}

func (h *Handler) RequestAccountDeletion(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	delReq := security.NewAccountDeletionRequest(user.ID)
	user.DeletionRequestedAt = &delReq.RequestedAt

	identity.WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":               "deletion_requested",
		"user_id":              user.ID,
		"requested_at":         delReq.RequestedAt.Format(time.RFC3339),
		"permanent_delete_at":  delReq.PermanentDeleteAt.Format(time.RFC3339),
		"grace_period_days":    30,
		"message":              "Your account is scheduled for permanent deletion in 30 days. You can cancel this request anytime before then by logging in.",
	})
}

func (h *Handler) CancelAccountDeletion(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "User session not found")
		return
	}

	user.DeletionRequestedAt = nil
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "deletion_cancelled",
		"message": "Account deletion request has been successfully cancelled.",
	})
}
