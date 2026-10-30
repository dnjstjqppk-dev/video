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

	r.Get("/claims", h.ListClaims)
	r.Get("/claims/{id}", h.GetClaim)
	r.Post("/claims", h.CreateClaim)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/claims/{id}/dispute", h.DisputeClaim)
		pr.Post("/claims/{id}/resolve", h.ResolveClaim)
		pr.Post("/claims/dmca-takedown", h.SubmitDMCATakedown)
		pr.Post("/claims/dmca-counter-notice", h.SubmitDMCACounterNotice)
	})

	return r
}

func (h *Handler) GetClaim(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
		return
	}
	identity.WriteJSON(w, http.StatusOK, claim)
}

func (h *Handler) DisputeClaim(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
		return
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

	identity.WriteJSON(w, http.StatusOK, claim)
}

func (h *Handler) ResolveClaim(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claim, ok := h.Store.Claims[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Copyright claim not found")
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
			vid.Visibility = "BLOCKED"
		}
	default:
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid resolution")
		return
	}

	identity.WriteJSON(w, http.StatusOK, claim)
}

func (h *Handler) SubmitDMCATakedown(w http.ResponseWriter, r *http.Request) {
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

	// Block video immediately
	var strikeIssuedToUserID string
	if vid, ok := h.Store.Videos[payload.VideoID]; ok {
		vid.Visibility = "BLOCKED"
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
	var payload struct {
		NoticeID       string `json:"notice_id"`
		LegalFullName  string `json:"legal_full_name"`
		Address        string `json:"address"`
		ConsentFederal bool   `json:"consent_federal_jurisdiction"`
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

	notice.Status = "counter_noticed"

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"notice_id":         notice.ID,
		"status":            "counter_noticed",
		"waiting_period":    "10-14 business days before video restoration",
		"claimant_notified": true,
	})
}

func (h *Handler) ListClaims(w http.ResponseWriter, r *http.Request) {
	var list []*store.CopyrightClaim
	for _, c := range h.Store.Claims {
		list = append(list, c)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		VideoID string `json:"videoId"`
		AssetID string `json:"assetId"`
		Action  string `json:"action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid copyright claim")
		return
	}

	claim := &store.CopyrightClaim{
		ID:                uuid.NewString(),
		ClaimedVideoID:    payload.VideoID,
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
