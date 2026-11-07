package studio

import (
	"net/http"
	"sort"
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
	r.Use(auth.RequireAuth)

	r.Get("/dashboard", h.GetDashboard)
	r.Get("/videos", h.GetVideos)
	r.Get("/analytics", h.GetAnalytics)
	r.Get("/comments", h.GetComments)
	r.Get("/claims", h.GetClaims)
	r.Post("/videos/{id}/edit-jobs", h.CreateEditJob)
	r.Get("/videos/{id}/edit-jobs/{jobId}", h.GetEditJob)

	return r
}

func (h *Handler) ownedVideos(channelID string) []*store.Video {
	var owned []*store.Video
	for _, v := range h.Store.Videos {
		if v.ChannelID == channelID && v.Status != "DELETED" {
			owned = append(owned, v)
		}
	}
	sort.SliceStable(owned, func(i, j int) bool {
		return owned[i].CreatedAt.After(owned[j].CreatedAt)
	})
	return owned
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	channel := h.Store.ChannelForUser(user.ID)
	if channel == nil {
		identity.WriteError(w, r, http.StatusNotFound, "CHANNEL_NOT_FOUND", "Create a channel before opening the studio dashboard")
		return
	}

	owned := h.ownedVideos(channel.ID)
	var totalViews, totalLikes int64
	for _, v := range owned {
		totalViews += v.ViewCount
		totalLikes += v.LikeCount
	}

	var recentComments []*store.Comment
	for _, v := range owned {
		recentComments = append(recentComments, h.Store.Comments[v.ID]...)
	}
	sort.SliceStable(recentComments, func(i, j int) bool {
		return recentComments[i].CreatedAt.After(recentComments[j].CreatedAt)
	})
	if len(recentComments) > 5 {
		recentComments = recentComments[:5]
	}
	if recentComments == nil {
		recentComments = []*store.Comment{}
	}

	recentVideos := owned
	if len(recentVideos) > 10 {
		recentVideos = recentVideos[:10]
	}
	if recentVideos == nil {
		recentVideos = []*store.Video{}
	}

	estimatedRevenue := totalViews * 4 // KRW per-view blended estimate
	dashboard := map[string]interface{}{
		"channel": channel,
		"analytics": map[string]interface{}{
			"views":            totalViews,
			"watchTimeHours":   totalViews / 120,
			"subscribers":      channel.SubscriberCount,
			"estimatedRevenue": estimatedRevenue,
		},
		"recentVideos":   recentVideos,
		"recentComments": recentComments,
	}

	identity.WriteJSON(w, http.StatusOK, dashboard)
}

func (h *Handler) GetVideos(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	channel := h.Store.ChannelForUser(user.ID)
	if channel == nil {
		identity.WriteError(w, r, http.StatusNotFound, "CHANNEL_NOT_FOUND", "Create a channel before listing studio videos")
		return
	}
	statusFilter := r.URL.Query().Get("status")
	page, limit := identity.ParsePagination(r, 20)

	owned := h.ownedVideos(channel.ID)
	if statusFilter != "" {
		filtered := owned[:0]
		for _, v := range owned {
			if v.Status == statusFilter {
				filtered = append(filtered, v)
			}
		}
		owned = filtered
	}
	window, total := identity.PaginateSlice(owned, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

// rangeDays converts the UI range selector (7d/28d/90d) to trailing day counts.
func rangeDays(raw string) int {
	switch raw {
	case "7d":
		return 7
	case "90d":
		return 90
	default:
		return 28
	}
}

func distributeSeries(total int64, days int) []map[string]interface{} {
	series := make([]map[string]interface{}, 0, days)
	if days <= 0 {
		return series
	}
	// Deterministic ramp: recent days weigh more, weights sum to 1.
	weightSum := float64(days*(days+1)) / 2
	now := time.Now().UTC()
	for i := days; i >= 1; i-- {
		weight := float64(days-i+1) / weightSum
		series = append(series, map[string]interface{}{
			"date":  now.AddDate(0, 0, -(i - 1)).Format("2006-01-02"),
			"value": int64(weight * float64(total)),
		})
	}
	return series
}

func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	days := rangeDays(r.URL.Query().Get("range"))

	var totalViews, totalWatchHours int64
	var subs int64
	if channel := h.Store.ChannelForUser(user.ID); channel != nil {
		for _, v := range h.ownedVideos(channel.ID) {
			totalViews += v.ViewCount
		}
		subs = channel.SubscriberCount
	}
	totalWatchHours = totalViews / 120

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"viewsTimeSeries":      distributeSeries(totalViews, days),
		"watchTimeTimeSeries":  distributeSeries(totalWatchHours, days),
		"subscriberTimeSeries": distributeSeries(subs, days),
		"revenueTimeSeries":    distributeSeries(totalViews*4, days),
	})
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	channel := h.Store.ChannelForUser(user.ID)
	if channel == nil {
		identity.WriteError(w, r, http.StatusNotFound, "CHANNEL_NOT_FOUND", "Create a channel before listing studio comments")
		return
	}
	page, limit := identity.ParsePagination(r, 20)

	var allComments []*store.Comment
	for _, v := range h.ownedVideos(channel.ID) {
		allComments = append(allComments, h.Store.Comments[v.ID]...)
	}
	sort.SliceStable(allComments, func(i, j int) bool {
		return allComments[i].CreatedAt.After(allComments[j].CreatedAt)
	})
	window, total := identity.PaginateSlice(allComments, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

// GetClaims lists copyright claims filed against the caller's own videos.
func (h *Handler) GetClaims(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	channel := h.Store.ChannelForUser(user.ID)
	if channel == nil {
		identity.WriteError(w, r, http.StatusNotFound, "CHANNEL_NOT_FOUND", "Create a channel before listing claims")
		return
	}
	ownedIDs := map[string]bool{}
	for _, v := range h.ownedVideos(channel.ID) {
		ownedIDs[v.ID] = true
	}
	var mine []*store.CopyrightClaim
	for _, c := range h.Store.Claims {
		if ownedIDs[c.ClaimedVideoID] {
			mine = append(mine, c)
		}
	}
	sort.SliceStable(mine, func(i, j int) bool {
		return mine[i].CreatedAt.After(mine[j].CreatedAt)
	})
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": mine,
		"total": len(mine),
	})
}
