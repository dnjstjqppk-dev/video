package ads

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

type Handler struct {
	Store *store.MemoryStore
	mu    sync.Mutex
	// freqImpressions tracks SERVED decisions per user+campaign for hourly
	// frequency capping (P0-5). Key: userID + "|" + campaignID.
	freqImpressions map[string][]time.Time
	// CreativeDir overrides the creative-asset root (tests; else env/default).
	CreativeDir string
	// creativeFiles maps creative IDs to stored files; creativeMIME to sniffed types.
	creativeFiles map[string]string
	creativeMIME  map[string]string
}

// hourlyCapPerUserCampaign bounds serves per user per campaign per hour until
// per-campaign caps land with the persistent ad engine (P0-1).
const hourlyCapPerUserCampaign = 3

const freqWindow = time.Hour

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{
		Store:           s,
		freqImpressions: make(map[string][]time.Time),
		creativeFiles:   make(map[string]string),
		creativeMIME:    make(map[string]string),
	}
}

// capped reports whether the user already hit the hourly cap for the campaign.
func (h *Handler) capped(userID, campaignID string) bool {
	if userID == "" {
		return false
	}
	key := userID + "|" + campaignID
	cutoff := time.Now().Add(-freqWindow)
	kept := h.freqImpressions[key][:0]
	for _, at := range h.freqImpressions[key] {
		if at.After(cutoff) {
			kept = append(kept, at)
		}
	}
	h.freqImpressions[key] = kept
	return len(kept) >= hourlyCapPerUserCampaign
}

func (h *Handler) recordServe(userID, campaignID string) {
	if userID == "" {
		return
	}
	key := userID + "|" + campaignID
	h.freqImpressions[key] = append(h.freqImpressions[key], time.Now())
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	// Public / Client Ad Runtime endpoints
	r.Post("/decision", h.MakeAdDecision)
	r.Post("/events", h.RecordAdEvent)

	// Public creative asset delivery (handler gates non-approved creatives)
	r.Get("/creatives/{id}/asset", h.ServeCreativeAsset)

	// Advertiser Portal endpoints (P0-8: authenticated advertiser roles only)
	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireRole("advertiser", "partner_manager", "platform_admin"))
		pr.Post("/preview", h.PreviewAdDecision)
		pr.Get("/campaigns", h.ListCampaigns)
		pr.Post("/campaigns", h.CreateCampaign)
		pr.Get("/campaigns/{id}", h.GetCampaign)
		pr.Patch("/campaigns/{id}/status", h.UpdateCampaignStatus)
		pr.Post("/campaigns/{id}/creatives", h.UploadCreative)
		pr.Get("/advertiser/reports", h.GetAdvertiserReports)
		pr.Get("/reports", h.GetAdvertiserReports)
		pr.Get("/alerts", h.GetDeliveryAlerts)
	})

	// Admin Governance & Diagnostic endpoints (P0-8: platform admins only)
	r.Group(func(ar chi.Router) {
		ar.Use(auth.RequireRole("super_admin", "platform_admin"))
		ar.Get("/admin/diagnostics", h.GetDeliveryDiagnostics)
		ar.Get("/admin/creatives", h.ListCreatives)
		ar.Post("/admin/creatives/{id}/approve", h.ApproveCreative)
		ar.Post("/admin/creatives/{id}/reject", h.RejectCreative)
	})

	return r
}

// AdvertiserRoutes serves the /v1/advertiser mount: the advertiser-portal
// surface without the admin governance paths (P0-8 duplicate-mount fix).
func (h *Handler) AdvertiserRoutes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	// Public / Client Ad Runtime endpoints
	r.Post("/decision", h.MakeAdDecision)
	r.Post("/events", h.RecordAdEvent)

	// Advertiser Portal endpoints (authenticated advertiser roles only)
	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireRole("advertiser", "partner_manager", "platform_admin"))
		pr.Post("/preview", h.PreviewAdDecision)
		pr.Get("/campaigns", h.ListCampaigns)
		pr.Post("/campaigns", h.CreateCampaign)
		pr.Get("/campaigns/{id}", h.GetCampaign)
		pr.Patch("/campaigns/{id}/status", h.UpdateCampaignStatus)
		pr.Post("/campaigns/{id}/creatives", h.UploadCreative)
		pr.Get("/reports", h.GetAdvertiserReports)
		pr.Get("/alerts", h.GetDeliveryAlerts)
	})

	return r
}

// DeliveryAlert flags a campaign whose spend pace deviates from its time
// pace (P1-17 under/over-delivery monitoring).
type DeliveryAlert struct {
	CampaignID string  `json:"campaign_id"`
	Name       string  `json:"name"`
	Kind       string  `json:"kind"` // UNDER_DELIVERY | OVER_DELIVERY
	SpendPace  float64 `json:"spend_pace"`
	TimePace   float64 `json:"time_pace"`
}

// GetDeliveryAlerts lists pacing alerts for the caller's campaigns.
func (h *Handler) GetDeliveryAlerts(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	alerts := make([]DeliveryAlert, 0)
	for _, camp := range h.Store.Campaigns {
		if !campaignVisible(camp, user) {
			continue
		}
		if camp.Status != "active" || camp.TotalBudgetCents <= 0 {
			continue
		}
		if camp.StartAt.IsZero() || camp.EndAt.IsZero() || !camp.EndAt.After(camp.StartAt) {
			continue
		}
		if now.Before(camp.StartAt) || now.After(camp.EndAt) {
			continue
		}
		timePace := now.Sub(camp.StartAt).Hours() / camp.EndAt.Sub(camp.StartAt).Hours()
		spendPace := float64(camp.SpentCents) / float64(camp.TotalBudgetCents)
		deviation := spendPace - timePace
		if deviation <= -0.25 {
			alerts = append(alerts, DeliveryAlert{CampaignID: camp.ID, Name: camp.Name, Kind: "UNDER_DELIVERY", SpendPace: spendPace, TimePace: timePace})
		} else if deviation >= 0.25 {
			alerts = append(alerts, DeliveryAlert{CampaignID: camp.ID, Name: camp.Name, Kind: "OVER_DELIVERY", SpendPace: spendPace, TimePace: timePace})
		}
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{"alerts": alerts})
}

func isPlatformAdmin(user *store.User) bool {
	return user != nil && (user.Role == "super_admin" || user.Role == "platform_admin")
}

// campaignVisible enforces advertiser tenant isolation: advertisers see only
// their own campaigns, platform admins see everything.
func campaignVisible(camp *store.AdCampaign, user *store.User) bool {
	if user == nil || camp == nil {
		return false
	}
	return isPlatformAdmin(user) || camp.AdvertiserID == user.ID
}

// adStatusTransitions is the ARTICLE 7.1 review state machine.
var adStatusTransitions = map[string][]string{
	"draft":          {"pending_review", "archived"},
	"pending_review": {"approved", "rejected", "draft"},
	"approved":       {"active", "scheduled", "paused"},
	"scheduled":      {"active", "paused", "expired"},
	"active":         {"paused", "completed", "exhausted", "suspended"},
	"paused":         {"active", "completed", "archived"},
	"rejected":       {"draft", "archived"},
	"suspended":      {"active", "archived"},
	"completed":      {"archived"},
	"exhausted":      {"archived"},
	"expired":        {"archived"},
	"archived":       {},
}

// advertiserTargets are the statuses a non-admin advertiser may set directly.
var advertiserTargets = map[string]bool{
	"pending_review": true,
	"paused":         true,
	"active":         true,
	"draft":          true,
	"archived":       true,
}

func transitionAllowed(from, to string) bool {
	for _, next := range adStatusTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// 1. AD DECISION API (ARTICLE 7.2 & 7.3)
// -----------------------------------------------------------------------------

type AdDecisionRequest struct {
	SlotID   string `json:"slot_id"` // watch_preroll, home_feed_banner, shorts_sponsored
	UserID   string `json:"user_id"`
	Category string `json:"category"`
	Platform string `json:"platform"` // WEB, ANDROID, IOS
}

type AdDecisionResponse struct {
	AdAvailable bool              `json:"ad_available"`
	DecisionID  string            `json:"decision_id,omitempty"`
	CampaignID  string            `json:"campaign_id,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	Creative    *store.AdCreative `json:"creative,omitempty"`
	Tracking    map[string]string `json:"tracking,omitempty"`
}

func (h *Handler) MakeAdDecision(w http.ResponseWriter, r *http.Request) {
	var req AdDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SlotID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Valid slot_id is required")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	decisionID := "dec-" + uuid.NewString()

	// Find slot definition
	slot, slotExists := h.Store.AdSlots[req.SlotID]
	if !slotExists {
		h.recordDecision(decisionID, req.SlotID, "", "", req.UserID, "UNFILLED", "UNKNOWN_SLOT_ID", 0, nil)
		identity.WriteJSON(w, http.StatusOK, AdDecisionResponse{
			AdAvailable: false,
			Reason:      "UNKNOWN_SLOT_ID",
		})
		return
	}

	// Filter candidate campaigns
	type candidate struct {
		campaign *store.AdCampaign
		creative *store.AdCreative
		bidScore int64
	}
	var candidates []candidate
	cappedOut := false

	for _, camp := range h.Store.Campaigns {
		// 1. Status check
		if camp.Status != "active" {
			continue
		}
		// 2. Budget exhaustion check
		if camp.SpentCents >= camp.TotalBudgetCents {
			camp.Status = "exhausted"
			continue
		}
		// 3. Schedule flight check
		if (!camp.StartAt.IsZero() && now.Before(camp.StartAt)) || (!camp.EndAt.IsZero() && now.After(camp.EndAt)) {
			continue
		}
		// 3b. Frequency cap check (P0-5)
		if h.capped(req.UserID, camp.ID) {
			cappedOut = true
			continue
		}
		// 4. Category targeting check
		if req.Category != "" && len(camp.TargetCategories) > 0 {
			categoryMatched := false
			for _, cat := range camp.TargetCategories {
				if strings.EqualFold(cat, req.Category) || strings.EqualFold(cat, "all") {
					categoryMatched = true
					break
				}
			}
			if !categoryMatched {
				continue
			}
		}

		// 5. Approved Creative format matching slot format
		var matchedCreative *store.AdCreative
		for _, creativeID := range camp.Creatives {
			if c, ok := h.Store.AdCreatives[creativeID]; ok {
				if c.ReviewStatus == "approved" && (c.Format == slot.Format || camp.Format == slot.Format) {
					matchedCreative = c
					break
				}
			}
		}
		// Fallback check: any creative with this campaign ID
		if matchedCreative == nil {
			for _, c := range h.Store.AdCreatives {
				if c.CampaignID == camp.ID && c.ReviewStatus == "approved" && c.Format == slot.Format {
					matchedCreative = c
					break
				}
			}
		}

		if matchedCreative != nil {
			candidates = append(candidates, candidate{
				campaign: camp,
				creative: matchedCreative,
				bidScore: camp.CpmBidCents,
			})
		}
	}

	if len(candidates) == 0 {
		reason := "NO_ELIGIBLE_CAMPAIGN"
		if cappedOut {
			reason = "FREQUENCY_CAPPED"
		}
		h.recordDecision(decisionID, req.SlotID, "", "", req.UserID, "UNFILLED", reason, 0, nil)
		identity.WriteJSON(w, http.StatusOK, AdDecisionResponse{
			AdAvailable: false,
			Reason:      reason,
		})
		return
	}

	// Sort by bid score descending (highest CPM bid gets precedence)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].bidScore > candidates[j].bidScore
	})

	winner := candidates[0]
	h.recordDecision(decisionID, req.SlotID, winner.campaign.ID, winner.creative.ID, req.UserID, "SERVED", "MATCHED", winner.campaign.CpmBidCents, winner.creative)
	h.recordServe(req.UserID, winner.campaign.ID)

	identity.WriteJSON(w, http.StatusOK, AdDecisionResponse{
		AdAvailable: true,
		DecisionID:  decisionID,
		CampaignID:  winner.campaign.ID,
		Reason:      "MATCHED",
		Creative:    winner.creative,
		Tracking: map[string]string{
			"event_url": "/v1/ads/events",
		},
	})
}

func (h *Handler) recordDecision(id, slotID, campaignID, creativeID, userID, status, reason string, bid int64, creative *store.AdCreative) {
	dec := &store.AdDeliveryDecision{
		DecisionID:  id,
		SlotID:      slotID,
		CampaignID:  campaignID,
		CreativeID:  creativeID,
		UserID:      userID,
		DecidedAt:   time.Now(),
		Status:      status,
		Reason:      reason,
		BidCpmCents: bid,
		Creative:    creative,
	}
	h.Store.AdDecisions[id] = dec
}

// -----------------------------------------------------------------------------
// 2. AD TELEMETRY & BUDGET PACING API (ARTICLE 7.2 Step 12-14)
// -----------------------------------------------------------------------------

type AdEventPayload struct {
	EventID    string `json:"event_id"`
	DecisionID string `json:"decision_id"`
	CampaignID string `json:"campaign_id"`
	EventType  string `json:"event_type"` // impression, start, first_quartile, midpoint, third_quartile, complete, click, skip
}

func (h *Handler) RecordAdEvent(w http.ResponseWriter, r *http.Request) {
	var payload AdEventPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.CampaignID == "" || payload.EventType == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid ad event payload")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	camp, campExists := h.Store.Campaigns[payload.CampaignID]
	if !campExists {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
		return
	}

	// P0-5: billing events must link to a real SERVED decision for this campaign.
	switch payload.EventType {
	case "impression", "click", "complete":
		if payload.DecisionID == "" {
			identity.WriteError(w, r, http.StatusBadRequest, "DECISION_REQUIRED", "Billing events require a decision_id")
			return
		}
		dec, ok := h.Store.AdDecisions[payload.DecisionID]
		if !ok || dec.Status != "SERVED" {
			identity.WriteError(w, r, http.StatusBadRequest, "UNKNOWN_DECISION", "Decision does not exist or was not served")
			return
		}
		if dec.CampaignID == "" || dec.CampaignID != payload.CampaignID {
			identity.WriteError(w, r, http.StatusBadRequest, "DECISION_MISMATCH", "Decision is not bound to this campaign")
			return
		}
	default:
		if payload.DecisionID != "" {
			dec, ok := h.Store.AdDecisions[payload.DecisionID]
			if !ok {
				identity.WriteError(w, r, http.StatusBadRequest, "UNKNOWN_DECISION", "Decision does not exist")
				return
			}
			if dec.CampaignID != "" && dec.CampaignID != payload.CampaignID {
				identity.WriteError(w, r, http.StatusBadRequest, "DECISION_MISMATCH", "Decision is not bound to this campaign")
				return
			}
		}
	}

	// Idempotency check: prevent double charging on retry
	if payload.EventID != "" {
		if _, exists := h.Store.IdempotencyKeys["ad_event_"+payload.EventID]; exists {
			identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "DUPLICATE_IGNORED"})
			return
		}
		h.Store.IdempotencyKeys["ad_event_"+payload.EventID] = payload.CampaignID
	}

	var costCents int64 = 0
	switch payload.EventType {
	case "impression":
		camp.Impressions++
		// CPM billing: cost per impression = CpmBid / 1000
		if camp.CpmBidCents > 0 {
			costCents = camp.CpmBidCents / 1000
			if costCents == 0 {
				costCents = 1
			}
		}
	case "click":
		camp.Clicks++
		if camp.CpcBidCents > 0 {
			costCents = camp.CpcBidCents
		}
	}

	camp.SpentCents += costCents
	if camp.SpentCents >= camp.TotalBudgetCents {
		camp.Status = "exhausted"
	}

	// Record telemetry event log
	eventRecord := &store.AdDeliveryEvent{
		EventID:    payload.EventID,
		DecisionID: payload.DecisionID,
		CampaignID: payload.CampaignID,
		EventType:  payload.EventType,
		CostCents:  costCents,
		OccurredAt: time.Now(),
	}
	h.Store.AdEvents = append(h.Store.AdEvents, eventRecord)

	identity.WriteJSON(w, http.StatusAccepted, map[string]interface{}{
		"status":      "ACCEPTED",
		"campaign_id": camp.ID,
		"event_type":  payload.EventType,
		"spent_cents": camp.SpentCents,
		"exhausted":   camp.Status == "exhausted",
	})
}

// -----------------------------------------------------------------------------
// 3. AD PREVIEW MODE (ARTICLE 7.5 Forced Test Assignment)
// -----------------------------------------------------------------------------

func (h *Handler) PreviewAdDecision(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	var req struct {
		CampaignID string `json:"campaign_id"`
		SlotID     string `json:"slot_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	h.mu.Lock()
	defer h.mu.Unlock()

	// Tenant isolation: preview only creatives of visible campaigns.
	visible := func(c *store.AdCreative) bool {
		camp, ok := h.Store.Campaigns[c.CampaignID]
		return ok && campaignVisible(camp, user)
	}

	// P0-5: forced assignment — preview a specific visible campaign.
	if req.CampaignID != "" {
		camp, ok := h.Store.Campaigns[req.CampaignID]
		if !ok || !campaignVisible(camp, user) {
			identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
			return
		}
		var forced *store.AdCreative
		for _, c := range h.Store.AdCreatives {
			if c.CampaignID == req.CampaignID {
				forced = c
				break
			}
		}
		if forced == nil {
			identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign has no creative asset to preview")
			return
		}
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"ad_available":    true,
			"preview_mode":    true,
			"forced":          true,
			"campaign_id":     camp.ID,
			"campaign_status": camp.Status,
			"creative":        forced,
			"message":         "Preview mode active: no impressions or budget deducted",
		})
		return
	}

	var creative *store.AdCreative
	for _, c := range h.Store.AdCreatives {
		if visible(c) {
			creative = c
			break
		}
	}

	if creative == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "No creative asset available for preview")
		return
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ad_available": true,
		"preview_mode": true,
		"creative":     creative,
		"message":      "Preview mode active: no impressions or budget deducted",
	})
}

// -----------------------------------------------------------------------------
// 4. AD DELIVERY INSPECTOR & DIAGNOSTICS (ARTICLE 7.4)
// -----------------------------------------------------------------------------

func (h *Handler) GetDeliveryDiagnostics(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	totalRequests := int64(len(h.Store.AdDecisions))
	servedDecisions := int64(0)
	unfilledDecisions := int64(0)
	reasonsMap := make(map[string]int)

	var recent []*store.AdDeliveryDecision
	for _, dec := range h.Store.AdDecisions {
		if dec.Status == "SERVED" {
			servedDecisions++
		} else {
			unfilledDecisions++
			reasonsMap[dec.Reason]++
		}
		recent = append(recent, dec)
	}

	// Sort recent decisions descending
	sort.Slice(recent, func(i, j int) bool {
		return recent[i].DecidedAt.After(recent[j].DecidedAt)
	})
	if len(recent) > 50 {
		recent = recent[:50]
	}

	totalImpressions := int64(0)
	totalClicks := int64(0)
	totalSpend := int64(0)
	activeCampaigns := 0

	for _, camp := range h.Store.Campaigns {
		totalImpressions += camp.Impressions
		totalClicks += camp.Clicks
		totalSpend += camp.SpentCents
		if camp.Status == "active" {
			activeCampaigns++
		}
	}

	var fillRate float64 = 0.0
	if totalRequests > 0 {
		fillRate = float64(servedDecisions) / float64(totalRequests) * 100.0
	}
	var ctr float64 = 0.0
	if totalImpressions > 0 {
		ctr = float64(totalClicks) / float64(totalImpressions) * 100.0
	}

	diagnostics := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_requests":             totalRequests,
			"served_decisions":           servedDecisions,
			"unfilled_decisions":         unfilledDecisions,
			"fill_rate_pct":              fillRate,
			"verified_impressions":       totalImpressions,
			"total_clicks":               totalClicks,
			"ctr_pct":                    ctr,
			"total_spend_cents":          totalSpend,
			"active_campaigns_count":     activeCampaigns,
			"total_registered_creatives": len(h.Store.AdCreatives),
		},
		"unfilled_reasons": reasonsMap,
		"recent_decisions": recent,
	}

	identity.WriteJSON(w, http.StatusOK, diagnostics)
}

// -----------------------------------------------------------------------------
// 5. CAMPAIGN CRUD & ADVERTISER PORTAL (ARTICLE 6)
// -----------------------------------------------------------------------------

func (h *Handler) ListCampaigns(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	list := make([]*store.AdCampaign, 0)
	for _, c := range h.Store.Campaigns {
		if campaignVisible(c, user) {
			list = append(list, c)
		}
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	h.mu.Lock()
	defer h.mu.Unlock()

	camp, ok := h.Store.Campaigns[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
		return
	}
	if !campaignVisible(camp, user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot access another advertiser's campaign")
		return
	}
	identity.WriteJSON(w, http.StatusOK, camp)
}

func (h *Handler) CreateCampaign(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Name             string   `json:"name"`
		Format           string   `json:"format"` // preroll, feed_banner, shorts_sponsored
		TotalBudgetCents int64    `json:"totalBudget"`
		DailyBudgetCents int64    `json:"dailyBudget"`
		CpmBidCents      int64    `json:"cpmBidCents"`
		CpcBidCents      int64    `json:"cpcBidCents"`
		Categories       []string `json:"categories"`
		CreativeMediaURL string   `json:"creativeMediaUrl"`
		Headline         string   `json:"headline"`
		DestinationURL   string   `json:"destinationUrl"`
		CallToAction     string   `json:"callToAction"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Name == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid campaign data")
		return
	}
	if payload.TotalBudgetCents < 0 || payload.DailyBudgetCents < 0 || payload.CpmBidCents < 0 || payload.CpcBidCents < 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Budgets and bids must not be negative")
		return
	}
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	campID := "camp-" + uuid.NewString()

	campaign := &store.AdCampaign{
		ID:               campID,
		AdvertiserID:     user.ID,
		Name:             payload.Name,
		Format:           payload.Format,
		Status:           "draft", // ARTICLE 7.1: review state machine, never auto-active
		TotalBudgetCents: payload.TotalBudgetCents,
		DailyBudgetCents: payload.DailyBudgetCents,
		CpmBidCents:      payload.CpmBidCents,
		CpcBidCents:      payload.CpcBidCents,
		StartAt:          now,
		EndAt:            now.Add(30 * 24 * time.Hour),
		TargetCategories: payload.Categories,
		CreatedAt:        now,
	}

	// Create associated creative asset if provided
	if payload.CreativeMediaURL != "" {
		creativeID := "creative-" + uuid.NewString()
		creative := &store.AdCreative{
			ID:                creativeID,
			CampaignID:        campID,
			Format:            payload.Format,
			MediaURL:          payload.CreativeMediaURL,
			ThumbnailURL:      payload.CreativeMediaURL,
			Headline:          payload.Headline,
			Description:       payload.Name,
			CallToAction:      payload.CallToAction,
			DestinationURL:    payload.DestinationURL,
			DurationSeconds:   15,
			SkipOffsetSeconds: 5,
			ReviewStatus:      "pending",
			CreatedAt:         now,
		}
		h.Store.AdCreatives[creativeID] = creative
		campaign.Creatives = append(campaign.Creatives, creativeID)
	}

	h.Store.Campaigns[campaign.ID] = campaign
	identity.WriteJSON(w, http.StatusCreated, campaign)
}

func (h *Handler) UpdateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	id := chi.URLParam(r, "id")
	var payload struct {
		Status string `json:"status"` // ARTICLE 7.1 state machine
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	h.mu.Lock()
	defer h.mu.Unlock()

	camp, ok := h.Store.Campaigns[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
		return
	}
	if !campaignVisible(camp, user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another advertiser's campaign")
		return
	}
	target := strings.ToLower(strings.TrimSpace(payload.Status))
	if target == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Status is required")
		return
	}
	if target == camp.Status {
		identity.WriteJSON(w, http.StatusOK, camp)
		return
	}
	if _, known := adStatusTransitions[target]; !known {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unknown campaign status")
		return
	}
	if !transitionAllowed(camp.Status, target) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_TRANSITION", "Status transition not allowed from "+camp.Status)
		return
	}
	if !isPlatformAdmin(user) && !advertiserTargets[target] {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Only platform admins may set status "+target)
		return
	}
	camp.Status = target
	identity.WriteJSON(w, http.StatusOK, camp)
}

func (h *Handler) GetAdvertiserReports(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	totalImpressions := int64(0)
	totalClicks := int64(0)
	totalSpend := int64(0)
	activeCount := 0
	campaignSummaries := make([]map[string]interface{}, 0)

	for _, camp := range h.Store.Campaigns {
		if !campaignVisible(camp, user) {
			continue
		}
		if camp.Status == "active" {
			activeCount++
		}
		totalImpressions += camp.Impressions
		totalClicks += camp.Clicks
		totalSpend += camp.SpentCents

		ctr := 0.0
		if camp.Impressions > 0 {
			ctr = float64(camp.Clicks) / float64(camp.Impressions) * 100.0
		}
		campaignSummaries = append(campaignSummaries, map[string]interface{}{
			"id":           camp.ID,
			"name":         camp.Name,
			"format":       camp.Format,
			"status":       camp.Status,
			"budget_cents": camp.TotalBudgetCents,
			"spent_cents":  camp.SpentCents,
			"impressions":  camp.Impressions,
			"clicks":       camp.Clicks,
			"ctr_pct":      ctr,
		})
	}

	overallCtr := 0.0
	if totalImpressions > 0 {
		overallCtr = float64(totalClicks) / float64(totalImpressions) * 100.0
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"overview": map[string]interface{}{
			"total_spend_cents": totalSpend,
			"total_impressions": totalImpressions,
			"total_clicks":      totalClicks,
			"overall_ctr_pct":   overallCtr,
			"active_campaigns":  activeCount,
		},
		"campaigns": campaignSummaries,
	})
}

// -----------------------------------------------------------------------------
// 6. CREATIVE REVIEW WORKFLOW (ARTICLE 7.1)
// -----------------------------------------------------------------------------

func (h *Handler) ListCreatives(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var list []*store.AdCreative
	for _, c := range h.Store.AdCreatives {
		list = append(list, c)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) ApproveCreative(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var payload struct {
		TOTPCode string `json:"totpCode"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	// P0-7: creative approvals require MFA step-up.
	if !identity.RequireStepUp(w, r, user, payload.TOTPCode) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	c, ok := h.Store.AdCreatives[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative asset not found")
		return
	}
	c.ReviewStatus = "approved"
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   actorID(user),
		"action":    "APPROVE_AD_CREATIVE",
		"target":    id,
		"ip":        r.RemoteAddr,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	identity.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) RejectCreative(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	id := chi.URLParam(r, "id")
	var payload struct {
		Reason   string `json:"reason"`
		TOTPCode string `json:"totpCode"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)
	// P0-7: creative rejections require MFA step-up.
	if !identity.RequireStepUp(w, r, user, payload.TOTPCode) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()

	c, ok := h.Store.AdCreatives[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative asset not found")
		return
	}
	c.ReviewStatus = "rejected"
	h.Store.AuditLogs = append(h.Store.AuditLogs, map[string]interface{}{
		"id":        "audit-" + uuid.NewString(),
		"actorId":   actorID(user),
		"action":    "REJECT_AD_CREATIVE",
		"target":    id,
		"reason":    payload.Reason,
		"ip":        r.RemoteAddr,
		"timestamp": time.Now().Format(time.RFC3339),
	})
	identity.WriteJSON(w, http.StatusOK, c)
}

func actorID(user *store.User) string {
	if user == nil {
		return "unknown"
	}
	return user.ID
}
