package copyright

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
	// Seed demo claim
	claim := &store.CopyrightClaim{
		ID:                "claim-1",
		ClaimedVideoID:    "00000000-0000-0000-0000-000000000001",
		OwnerName:         "Sony Music Entertainment",
		ClaimAction:       "monetize",
		Status:            "active",
		MatchStartSeconds: 124,
		MatchEndSeconds:   152,
		CreatedAt:         time.Now().Add(-48 * time.Hour),
	}
	s.Claims[claim.ID] = claim
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(auth.RequireAuth)

	r.Get("/claims", h.ListClaims)
	r.Get("/claims/{id}", h.GetClaim)
	r.Post("/claims", h.CreateClaim)
	r.Patch("/claims/{id}", h.UpdateClaimAction)
	r.Post("/claims/{id}/dispute", h.DisputeClaim)
	r.Post("/claims/{id}/resolve", h.ResolveClaim)
	r.Post("/claims/dmca-takedown", h.SubmitDMCATakedown)
	r.Post("/claims/dmca-counter-notice", h.SubmitDMCACounterNotice)

	return r
}

func isOperator(user *store.User) bool {
	return user != nil && (user.Role == "super_admin" || user.Role == "platform_admin" || user.Role == "moderator")
}

func isRightsRole(user *store.User) bool {
	if user == nil {
		return false
	}
	switch user.Role {
	case "rights_holder", "rights_owner", "partner_manager", "mcn_manager", "partner_admin":
		return true
	default:
		return isOperator(user)
	}
}

// claimVisible scopes claims to their parties: the reporter, the claimed
// video's channel owner, rights roles handling the dispute, and operators.
func (h *Handler) claimVisible(claim *store.CopyrightClaim, user *store.User) bool {
	if user == nil || claim == nil {
		return false
	}
	if isOperator(user) || isRightsRole(user) {
		return true
	}
	if claim.ReporterUserID != "" && claim.ReporterUserID == user.ID {
		return true
	}
	if video, ok := h.Store.Videos[claim.ClaimedVideoID]; ok {
		if channel, ok := h.Store.Channels[video.ChannelID]; ok && channel != nil {
			return channel.OwnerID == user.ID
		}
	}
	return false
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
		return
	}
	if !h.claimVisible(claim, user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot access another party's claim")
		return
	}
	identity.WriteJSON(w, http.StatusOK, claim)
}

func (h *Handler) DisputeClaim(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
		return
	}

	if video, ok := h.Store.Videos[claim.ClaimedVideoID]; ok {
		if channel, ok := h.Store.Channels[video.ChannelID]; ok && channel != nil {
			if channel.OwnerID != user.ID && !isOperator(user) {
				identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Only the channel owner can dispute this claim")
				return
			}
		}
	}

	var payload struct {
		Reason      string `json:"reason"`      // fair_use, license_held, public_domain, original_content
		Explanation string `json:"explanation"` // statement
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Reason == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Dispute reason required")
		return
	}

	now := time.Now()
	claim.Status = "disputed"
	claim.DisputeReason = payload.Reason
	claim.DisputeText = payload.Explanation
	claim.DisputeFiledAt = &now

	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "DISPUTE_COPYRIGHT_CLAIM",
		"target":    id,
		"reason":    payload.Reason,
		"ip":        r.RemoteAddr,
		"timestamp": now.Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, claim)
}

func (h *Handler) ResolveClaim(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
		return
	}
	if claim.ReporterUserID != user.ID && !isOperator(user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Only the reporting rights holder or an operator can resolve this claim")
		return
	}

	var payload struct {
		Resolution string `json:"resolution"` // release, uphold, takedown
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Resolution == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Resolution action required")
		return
	}

	now := time.Now()
	claim.ResolvedAt = &now

	switch payload.Resolution {
	case "release":
		claim.Status = "released"
	case "uphold":
		claim.Status = "upheld"
	case "takedown":
		claim.Status = "takedown"
		// Block claimed video
		if vid, ok := h.Store.Videos[claim.ClaimedVideoID]; ok {
			vid.Status = "BLOCKED"
		}
	default:
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid resolution")
		return
	}

	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "RESOLVE_COPYRIGHT_CLAIM",
		"target":    id,
		"reason":    payload.Resolution,
		"ip":        r.RemoteAddr,
		"timestamp": now.Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, claim)
}

func (h *Handler) SubmitDMCATakedown(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	if !isRightsRole(user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "DMCA takedowns require a rights-holder role")
		return
	}
	var payload struct {
		VideoID            string `json:"video_id"`
		ClaimantName       string `json:"claimant_name"`
		ClaimantEmail      string `json:"claimant_email"`
		CopyrightWorkTitle string `json:"copyright_work_title"`
		SignatureLegal     string `json:"signature_legal"`
		GoodFaithBelief    bool   `json:"good_faith_belief"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.VideoID == "" || !payload.GoodFaithBelief || payload.SignatureLegal == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Valid DMCA takedown with legal signature and good faith belief required")
		return
	}

	noticeID := uuid.NewString()
	notice := &store.DMCATakedownNotice{
		ID:                 noticeID,
		VideoID:            payload.VideoID,
		ClaimantName:       payload.ClaimantName,
		ClaimantEmail:      payload.ClaimantEmail,
		CopyrightWorkTitle: payload.CopyrightWorkTitle,
		InfringingURL:      "https://gvp.media/watch/" + payload.VideoID,
		SignatureLegal:     payload.SignatureLegal,
		GoodFaithBelief:    payload.GoodFaithBelief,
		Status:             "issued",
		CreatedAt:          time.Now(),
	}

	h.Store.DMCANotices[noticeID] = notice
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "SUBMIT_DMCA_TAKEDOWN",
		"target":    noticeID,
		"reason":    payload.CopyrightWorkTitle,
		"ip":        r.RemoteAddr,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	// Block video immediately
	var strikeIssuedToUserID string
	if vid, ok := h.Store.Videos[payload.VideoID]; ok {
		vid.Status = "BLOCKED"
		// Check channel owner
		if ch, ok := h.Store.Channels[vid.ChannelID]; ok {
			strikeIssuedToUserID = ch.OwnerID
			if owner, ok := h.Store.Users[ch.OwnerID]; ok {
				owner.CopyrightStrikes++
				if owner.CopyrightStrikes >= 3 {
					owner.IsActive = false
				}
			}
		}
	}

	identity.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"notice":         notice,
		"video_blocked":  true,
		"strike_issued":  strikeIssuedToUserID != "",
		"target_user_id": strikeIssuedToUserID,
	})
}

func (h *Handler) SubmitDMCACounterNotice(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	var payload struct {
		NoticeID         string `json:"notice_id"`
		LegalFullName    string `json:"legal_full_name"`
		Address          string `json:"address"`
		ConsentFederal   bool   `json:"consent_federal_jurisdiction"`
		StatementPerjury string `json:"statement_perjury"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.NoticeID == "" || !payload.ConsentFederal {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Valid counter notice with federal jurisdiction consent required")
		return
	}

	notice, ok := h.Store.DMCANotices[payload.NoticeID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "DMCA notice not found")
		return
	}
	if video, ok := h.Store.Videos[notice.VideoID]; ok {
		if channel, ok := h.Store.Channels[video.ChannelID]; ok && channel != nil {
			if channel.OwnerID != user.ID && !isOperator(user) {
				identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Only the channel owner can counter this notice")
				return
			}
		}
	}

	notice.Status = "counter_noticed"
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "SUBMIT_DMCA_COUNTER_NOTICE",
		"target":    notice.ID,
		"ip":        r.RemoteAddr,
		"timestamp": time.Now().Format(time.RFC3339),
	})

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"notice_id":         notice.ID,
		"status":            "counter_noticed",
		"waiting_period":    "10-14 business days before video restoration",
		"claimant_notified": true,
	})
}

func (h *Handler) ListClaims(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	list := make([]*store.CopyrightClaim, 0)
	for _, c := range h.Store.Claims {
		if h.claimVisible(c, user) {
			list = append(list, c)
		}
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func validClaimAction(action string) bool {
	switch action {
	case "monetize", "track", "block":
		return true
	default:
		return false
	}
}

func (h *Handler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	if !isRightsRole(user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_ROLE", "Filing claims requires a rights-holder role")
		return
	}
	var payload struct {
		VideoID string `json:"videoId"`
		AssetID string `json:"assetId"`
		Action  string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid copyright claim")
		return
	}
	if !validClaimAction(payload.Action) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Action must be monetize, track, or block")
		return
	}
	if _, ok := h.Store.Videos[payload.VideoID]; !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Claimed video not found")
		return
	}

	claim := &store.CopyrightClaim{
		ID:                uuid.NewString(),
		ClaimedVideoID:    payload.VideoID,
		ReporterUserID:    user.ID,
		OwnerName:         "Independent Content Rights Group",
		ClaimAction:       payload.Action,
		Status:            "pending",
		MatchStartSeconds: 0,
		MatchEndSeconds:   30,
		CreatedAt:         time.Now(),
	}

	h.Store.Claims[claim.ID] = claim
	identity.WriteJSON(w, http.StatusCreated, claim)
}

// UpdateClaimAction changes a claim's enforcement policy in place. Only the
// reporting rights holder or an operator may do so; everyone else keeps
// read-only visibility through claimVisible.
func (h *Handler) UpdateClaimAction(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
		return
	}
	if claim.ReporterUserID != user.ID && !isOperator(user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Only the reporting rights holder or an operator can change policy")
		return
	}
	var payload struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || !validClaimAction(payload.Action) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Action must be monetize, track, or block")
		return
	}
	claim.ClaimAction = payload.Action
	identity.WriteJSON(w, http.StatusOK, claim)
}
