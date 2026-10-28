package studio

import (
	"net/http"

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

	return r
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	channel := h.Store.Channels["aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"]
	var recentVideos []*store.Video
	for _, v := range h.Store.Videos {
		recentVideos = append(recentVideos, v)
	}

	dashboard := map[string]interface{}{
		"channel": channel,
		"analytics": map[string]interface{}{
			"views":            482000,
			"watchTimeHours":   14200,
			"subscribers":      channel.SubscriberCount,
			"estimatedRevenue": 3120000,
		},
		"recentVideos": recentVideos,
	}

	identity.WriteJSON(w, http.StatusOK, dashboard)
}

func (h *Handler) GetVideos(w http.ResponseWriter, r *http.Request) {
	var list []*store.Video
	for _, v := range h.Store.Videos {
		list = append(list, v)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": list,
		"total": len(list),
	})
}

func (h *Handler) GetAnalytics(w http.ResponseWriter, r *http.Request) {
	timeSeries := []map[string]interface{}{
		{"date": "2026-10-14", "value": 14200},
		{"date": "2026-10-15", "value": 16800},
		{"date": "2026-10-16", "value": 19400},
		{"date": "2026-10-17", "value": 24100},
		{"date": "2026-10-18", "value": 28900},
		{"date": "2026-10-19", "value": 31400},
		{"date": "2026-10-20", "value": 35200},
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"viewsTimeSeries":      timeSeries,
		"watchTimeTimeSeries":  timeSeries,
		"subscriberTimeSeries": timeSeries,
		"revenueTimeSeries":    timeSeries,
	})
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	var allComments []*store.Comment
	for _, comments := range h.Store.Comments {
		allComments = append(allComments, comments...)
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": allComments,
		"total": len(allComments),
	})
}
