package policies

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

// Kill-switch keys enforced across the platform.
const (
	FlagSignupEnabled  = "GLOBAL_PUBLIC_SIGNUP_ENABLED"
	FlagUploadIngest   = "MULTIPART_R2_UPLOAD_INGEST"
	FlagLiveIngest     = "LIVE_STREAM_RTMPS_INGEST"
	FlagPayoutDispatch = "AUTOMATED_LEDGER_PAYOUT_DISPATCH"
)

// FlagEnabled reports whether a kill-switch key is on. Unknown keys default
// to enabled so enforcement can only restrict known breakers.
func FlagEnabled(s *store.MemoryStore, key string) bool {
	if flag, ok := s.FeatureFlags[key]; ok {
		return flag.Enabled
	}
	return true
}

type Handler struct {
	Store *store.MemoryStore
}

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.With(auth.RequireRole("super_admin", "platform_admin", "finance_admin", "moderator", "support_agent")).Get("/", h.GetPolicies)
	r.With(auth.RequireRole("super_admin", "platform_admin")).Patch("/", h.UpdatePolicyConfig)
	r.With(auth.RequireRole("super_admin", "platform_admin")).Patch("/flags/{key}", h.UpdateFlag)

	return r
}

func (h *Handler) GetPolicies(w http.ResponseWriter, r *http.Request) {
	flags := make([]*store.FeatureFlag, 0, len(h.Store.FeatureFlags))
	for _, f := range h.Store.FeatureFlags {
		flags = append(flags, f)
	}
	sort.SliceStable(flags, func(i, j int) bool { return flags[i].Key < flags[j].Key })

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"config": h.Store.Policy,
		"flags":  flags,
	})
}

func (h *Handler) UpdatePolicyConfig(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var payload struct {
		MaxUploadSizeGB        *int64 `json:"max_upload_size_gb"`
		MaxVideoDurationHours  *int   `json:"max_video_duration_hours"`
		CreatorRevenueSharePct *int   `json:"creator_revenue_share_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid policy payload")
		return
	}

	policy := h.Store.Policy
	if payload.MaxUploadSizeGB != nil {
		if *payload.MaxUploadSizeGB < 1 || *payload.MaxUploadSizeGB > 1024 {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "max_upload_size_gb must be between 1 and 1024")
			return
		}
		policy.MaxUploadSizeGB = *payload.MaxUploadSizeGB
	}
	if payload.MaxVideoDurationHours != nil {
		if *payload.MaxVideoDurationHours < 1 || *payload.MaxVideoDurationHours > 24 {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "max_video_duration_hours must be between 1 and 24")
			return
		}
		policy.MaxVideoDurationHours = *payload.MaxVideoDurationHours
	}
	if payload.CreatorRevenueSharePct != nil {
		if *payload.CreatorRevenueSharePct < 0 || *payload.CreatorRevenueSharePct > 100 {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "creator_revenue_share_pct must be between 0 and 100")
			return
		}
		policy.CreatorRevenueSharePct = *payload.CreatorRevenueSharePct
	}
	policy.UpdatedBy = user.ID
	policy.UpdatedAt = time.Now()
	h.Store.Policy = policy

	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit_" + time.Now().Format("20060102150405"),
		"actorId":   user.ID,
		"action":    "UPDATE_POLICY_CONFIG",
		"target":    "global_platform_config",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, policy)
}

func (h *Handler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	key := chi.URLParam(r, "key")
	flag, ok := h.Store.FeatureFlags[key]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Feature flag not found")
		return
	}

	var payload struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Enabled == nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "enabled is required")
		return
	}

	flag.Enabled = *payload.Enabled
	flag.UpdatedBy = user.ID
	flag.UpdatedAt = time.Now()

	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit_" + time.Now().Format("20060102150405"),
		"actorId":   user.ID,
		"action":    "TOGGLE_FEATURE_FLAG",
		"target":    key,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, flag)
}
