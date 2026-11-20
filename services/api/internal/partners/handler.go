package partners

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
	return &Handler{Store: s}
}

// partnerRoles gates the partner portal surface (P1-18 RBAC).
var partnerRoles = []string{
	"partner_manager", "advertiser", "rights_holder", "rights_owner",
	"partner_finance", "mcn_manager", "partner_admin", "platform_admin",
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(auth.RequireRole(partnerRoles...))

	r.Get("/me", h.GetPartnerMe)
	r.Get("/channels", h.GetChannels)
	r.Get("/contracts", h.GetContracts)
	r.Post("/contracts", h.CreateContract)
	r.Patch("/contracts/{id}", h.RevokeContract)

	return r
}

func isPlatformOperator(user *store.User) bool {
	return user != nil && (user.Role == "super_admin" || user.Role == "platform_admin")
}

// managerSeesAll covers MCN staff operating across creators' channels. Until
// organization membership lands, managers are role-trusted; every other
// partner role sees only its own records.
func managerSeesAll(user *store.User) bool {
	if user == nil {
		return false
	}
	if isPlatformOperator(user) {
		return true
	}
	switch user.Role {
	case "partner_manager", "mcn_manager", "partner_admin":
		return true
	default:
		return false
	}
}

func (h *Handler) GetPartnerMe(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	campaigns := 0
	for _, c := range h.Store.Campaigns {
		if c.Status != "active" {
			continue
		}
		if managerSeesAll(user) || c.AdvertiserID == user.ID {
			campaigns++
		}
	}
	channels := 0
	for _, c := range h.Store.Channels {
		if managerSeesAll(user) || c.OwnerID == user.ID {
			channels++
		}
	}
	var pending int64
	for _, tx := range h.Store.Transactions {
		if tx.Status != "succeeded" {
			continue
		}
		if !managerSeesAll(user) && tx.UserID != user.ID {
			continue
		}
		pending += (tx.AmountCents / 100) * 70
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"organization": map[string]interface{}{
			"id":   "org-" + user.ID,
			"name": user.DisplayName,
			"type": user.Role,
		},
		"campaignCount":           campaigns,
		"activeChannelsCount":     channels,
		"pendingSettlementAmount": pending,
	})
}

func (h *Handler) GetChannels(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	list := make([]*store.Channel, 0)
	for _, c := range h.Store.Channels {
		if managerSeesAll(user) || (user != nil && c.OwnerID == user.ID) {
			list = append(list, c)
		}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

func (h *Handler) GetContracts(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	contracts := make([]*store.PartnerContract, 0)
	for _, c := range h.Store.Contracts {
		if managerSeesAll(user) {
			contracts = append(contracts, c)
			continue
		}
		if channel, ok := h.Store.Channels[c.ChannelID]; ok && channel != nil && user != nil && channel.OwnerID == user.ID {
			contracts = append(contracts, c)
		}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": contracts,
		"total": len(contracts),
	})
}

func (h *Handler) RevokeContract(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	contract, ok := h.Store.Contracts[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Contract not found")
		return
	}
	allowed := managerSeesAll(user)
	if !allowed {
		if channel, ok := h.Store.Channels[contract.ChannelID]; ok && channel != nil && channel.OwnerID == user.ID {
			allowed = true
		}
	}
	if !allowed {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot revoke another party's contract")
		return
	}
	var payload struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	reason := security.SanitizeUserText(payload.Reason, 500)
	if reason == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Revocation reason is required")
		return
	}
	if contract.Status == "revoked" {
		identity.WriteError(w, r, http.StatusConflict, "ALREADY_REVOKED", "Contract is already revoked")
		return
	}
	now := time.Now()
	contract.Status = "revoked"
	contract.RevokedAt = &now
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "REVOKE_PARTNER_CONTRACT",
		"target":    id,
		"reason":    reason,
		"ip":        r.RemoteAddr,
		"timestamp": now.Format(time.RFC3339),
	})
	identity.WriteJSON(w, http.StatusOK, contract)
}

// CreateContract links a channel into the MCN roster (manager-side only).
func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if !managerSeesAll(user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_ROLE", "Linking channels requires a partner manager role")
		return
	}
	var payload struct {
		ChannelID       string `json:"channelId"`
		CreatorSplitPct *int   `json:"creatorSplitPct"`
		PartnerType     string `json:"partnerType"`
		ValidUntil      string `json:"validUntil"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.ChannelID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "channelId is required")
		return
	}
	channel, ok := h.Store.Channels[payload.ChannelID]
	if !ok || channel == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}
	split := 85
	if payload.CreatorSplitPct != nil {
		split = *payload.CreatorSplitPct
	}
	if split < 0 || split > 100 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "creatorSplitPct must be 0-100")
		return
	}
	for _, c := range h.Store.Contracts {
		if c.ChannelID == payload.ChannelID && c.Status == "active" {
			identity.WriteError(w, r, http.StatusConflict, "ALREADY_LINKED", "Channel already has an active contract")
			return
		}
	}
	now := time.Now()
	contract := &store.PartnerContract{
		ID:               uuid.NewString(),
		ChannelID:        channel.ID,
		ChannelName:      channel.Name,
		ChannelHandle:    channel.Handle,
		PartnerType:      "MCN",
		CreatorSplitPct:  split,
		PartnerSplitPct:  100 - split,
		RevenueShareRate: float64(split) / 100,
		Status:           "active",
		ValidUntil:       payload.ValidUntil,
		CreatedAt:        now,
	}
	if payload.PartnerType != "" {
		contract.PartnerType = payload.PartnerType
	}
	h.Store.Contracts[contract.ID] = contract
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   user.ID,
		"action":    "CREATE_PARTNER_CONTRACT",
		"target":    contract.ID,
		"ip":        r.RemoteAddr,
		"timestamp": now.Format(time.RFC3339),
	})
	identity.WriteJSON(w, http.StatusCreated, contract)
}
