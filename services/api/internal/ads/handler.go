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
}

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	// Public / Client Ad Runtime endpoints
	r.Post("/decision", h.MakeAdDecision)
	r.Post("/events", h.RecordAdEvent)
	r.Post("/preview", h.PreviewAdDecision)

	// Campaign Management
	r.Get("/campaigns", h.ListCampaigns)
	r.Post("/campaigns", h.CreateCampaign)
	r.Get("/campaigns/{id}", h.GetCampaign)
	r.Patch("/campaigns/{id}/status", h.UpdateCampaignStatus)

	// Advertiser Portal endpoints
	r.Get("/advertiser/reports", h.GetAdvertiserReports)
	r.Get("/reports", h.GetAdvertiserReports)

	// Admin Governance & Diagnostic endpoints
	r.Get("/admin/diagnostics", h.GetDeliveryDiagnostics)
	r.Get("/admin/creatives", h.ListCreatives)
	r.Post("/admin/creatives/{id}/approve", h.ApproveCreative)
	r.Post("/admin/creatives/{id}/reject", h.RejectCreative)

	return r
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
		h.recordDecision(decisionID, req.SlotID, "", "", req.UserID, "UNFILLED", "NO_ELIGIBLE_CAMPAIGN", 0, nil)
		identity.WriteJSON(w, http.StatusOK, AdDecisionResponse{
			AdAvailable: false,
			Reason:      "NO_ELIGIBLE_CAMPAIGN",
		})
		return
	}

	// Sort by bid score descending (highest CPM bid gets precedence)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].bidScore > candidates[j].bidScore
	})

	winner := candidates[0]
	h.recordDecision(decisionID, req.SlotID, winner.campaign.ID, winner.creative.ID, req.UserID, "SERVED", "MATCHED", winner.campaign.CpmBidCents, winner.creative)

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

	// Idempotency check: prevent double charging on retry
	if payload.EventID != "" {
		if _, exists := h.Store.IdempotencyKeys["ad_event_"+payload.EventID]; exists {
			identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "DUPLICATE_IGNORED"})
			return
		}
		h.Store.IdempotencyKeys["ad_event_"+payload.EventID] = payload.CampaignID
	}

	camp, campExists := h.Store.Campaigns[payload.CampaignID]
	if !campExists {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
		return
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
	var req struct {
		CampaignID string `json:"campaign_id"`
		SlotID     string `json:"slot_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	h.mu.Lock()
	defer h.mu.Unlock()

	var creative *store.AdCreative
	if req.CampaignID != "" {
		for _, c := range h.Store.AdCreatives {
			if c.CampaignID == req.CampaignID {
				creative = c
				break
			}
		}
	}
	if creative == nil {
		for _, c := range h.Store.AdCreatives {
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
	h.mu.Lock()
	defer h.mu.Unlock()

	var list []*store.AdCampaign
	for _, c := range h.Store.Campaigns {
		list = append(list, c)
	}
	identity.WriteJSON(w, http.StatusOK, list)
}

func (h *Handler) GetCampaign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.mu.Lock()
	defer h.mu.Unlock()

	camp, ok := h.Store.Campaigns[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
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

	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	campID := "camp-" + uuid.NewString()

	campaign := &store.AdCampaign{
		ID:               campID,
		AdvertiserID:     "44444444-4444-4444-4444-444444444444",
		Name:             payload.Name,
		Format:           payload.Format,
		Status:           "active", // Activated upon creation in demo/testing
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
			ReviewStatus:      "approved",
			CreatedAt:         now,
		}
		h.Store.AdCreatives[creativeID] = creative
		campaign.Creatives = append(campaign.Creatives, creativeID)
	}

	h.Store.Campaigns[campaign.ID] = campaign
	identity.WriteJSON(w, http.StatusCreated, campaign)
}

func (h *Handler) UpdateCampaignStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var payload struct {
		Status string `json:"status"` // active, paused, archived
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	h.mu.Lock()
	defer h.mu.Unlock()

	camp, ok := h.Store.Campaigns[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
		return
	}
	camp.Status = payload.Status
	identity.WriteJSON(w, http.StatusOK, camp)
}

func (h *Handler) GetAdvertiserReports(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	defer h.mu.Unlock()

	totalImpressions := int64(0)
	totalClicks := int64(0)
	totalSpend := int64(0)
	var campaignSummaries []map[string]interface{}

	for _, camp := range h.Store.Campaigns {
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
			"active_campaigns":  len(h.Store.Campaigns),
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
	id := chi.URLParam(r, "id")
	h.mu.Lock()
	defer h.mu.Unlock()

	c, ok := h.Store.AdCreatives[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative asset not found")
		return
	}
	c.ReviewStatus = "approved"
	identity.WriteJSON(w, http.StatusOK, c)
}

func (h *Handler) RejectCreative(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.mu.Lock()
	defer h.mu.Unlock()

	c, ok := h.Store.AdCreatives[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative asset not found")
		return
	}
	c.ReviewStatus = "rejected"
	identity.WriteJSON(w, http.StatusOK, c)
}
