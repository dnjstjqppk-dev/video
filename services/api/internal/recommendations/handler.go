package recommendations

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/videos"
)

type Handler struct {
	Store *store.MemoryStore
}

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/home", h.GetHomeFeed)
	r.Get("/shorts", h.GetShortsFeed)
	r.Get("/subscriptions", h.GetSubscriptionsFeed)
	r.Post("/feedback", h.Feedback)

	return r
}

func (h *Handler) GetHomeFeed(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	page, limit := identity.ParsePagination(r, 24)
	var feed []*store.Video

	for _, v := range h.Store.Videos {
		// Shelf split: AI-made videos live in the AI Originals rail, not
		// the general home feed, so the main shelf stays human-made.
		if v.Visibility == "public" && v.Status == "READY" && v.VideoType != "shorts" && !v.AiGenerated {
			if category != "" && category != "all" && !strings.EqualFold(v.Category, category) {
				continue
			}
			feed = append(feed, v)
		}
	}

	window, total := identity.PaginateSlice(feed, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

func (h *Handler) GetShortsFeed(w http.ResponseWriter, r *http.Request) {
	page, limit := identity.ParsePagination(r, 10)
	var shorts []*store.Video
	for _, v := range h.Store.Videos {
		// Shelf split: AI shorts surface in AI Originals, not the
		// general shorts feed.
		if v.Visibility == "public" && v.Status == "READY" && v.VideoType == "shorts" && !v.AiGenerated {
			shorts = append(shorts, v)
		}
	}

	window, total := identity.PaginateSlice(shorts, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

func (h *Handler) GetSubscriptionsFeed(w http.ResponseWriter, r *http.Request) {
	page, limit := identity.ParsePagination(r, 24)
	user := identity.GetUserFromContext(r.Context())
	var feed []*store.Video
	for _, v := range h.Store.Videos {
		if v.Visibility != "public" || v.Status != "READY" {
			continue
		}
		if user != nil {
			if subs, ok := h.Store.Subscriptions[user.ID]; !ok || !subs[v.ChannelID] {
				continue
			}
		}
		feed = append(feed, v)
	}

	window, total := identity.PaginateSlice(feed, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

// GetOriginals serves the AI Originals shelf: approved AI videos plus
// AI Original series. Public. ?type=videos|series|all (default all).
func (h *Handler) GetOriginals(w http.ResponseWriter, r *http.Request) {
	which := r.URL.Query().Get("type")
	if which == "" {
		which = "all"
	}
	videoType := r.URL.Query().Get("video_type")
	page, limit := identity.ParsePagination(r, 24)

	var approved []*store.Video
	if which == "all" || which == "videos" {
		for _, v := range h.Store.Videos {
			if !v.AiGenerated || v.AiReviewStatus != "approved" {
				continue
			}
			if v.Visibility != "public" || v.Status != "READY" {
				continue
			}
			if videoType != "" && v.VideoType != videoType {
				continue
			}
			approved = append(approved, v)
		}
		sort.Slice(approved, func(i, j int) bool { return approved[i].PublishedAt.After(approved[j].PublishedAt) })
	}
	var originals []*store.Series
	if which == "all" || which == "series" {
		for _, s := range h.Store.Series {
			if s.AiOriginal {
				originals = append(originals, s)
			}
		}
		sort.Slice(originals, func(i, j int) bool { return originals[i].UpdatedAt.After(originals[j].UpdatedAt) })
	}

	videoWindow, videoTotal := identity.PaginateSlice(approved, page, limit)
	seriesWindow, seriesTotal := identity.PaginateSlice(originals, page, limit)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"videos": identity.PageEnvelope(videoWindow, videoTotal, page, limit),
		"series": identity.PageEnvelope(seriesWindow, seriesTotal, page, limit),
	})
}

func (h *Handler) Feedback(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		VideoID string `json:"videoId"`
		Action  string `json:"action"` // not_interested, dont_recommend_channel
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	page, limit := identity.ParsePagination(r, 20)
	user := identity.GetUserFromContext(r.Context())
	var matchedVideos []*store.Video
	var matchedChannels []*store.Channel

	if query != "" {
		for _, v := range h.Store.Videos {
			if !videos.SearchVisible(user, v, h.Store.Channels) {
				continue
			}
			if strings.Contains(strings.ToLower(v.Title), query) || strings.Contains(strings.ToLower(v.Description), query) {
				matchedVideos = append(matchedVideos, v)
			}
		}

		for _, c := range h.Store.Channels {
			if strings.Contains(strings.ToLower(c.Name), query) || strings.Contains(strings.ToLower(c.Handle), query) {
				matchedChannels = append(matchedChannels, c)
			}
		}
	}

	videoWindow, _ := identity.PaginateSlice(matchedVideos, page, limit)
	channelWindow, _ := identity.PaginateSlice(matchedChannels, page, limit)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"videos":   videoWindow,
		"channels": channelWindow,
		"total":    len(matchedVideos) + len(matchedChannels),
		"page":     page,
		"limit":    limit,
	})
}

func (h *Handler) SearchSuggestions(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	user := identity.GetUserFromContext(r.Context())
	suggestions := []string{}
	if q != "" {
		for _, v := range h.Store.Videos {
			if !videos.SearchVisible(user, v, h.Store.Channels) {
				continue
			}
			if strings.HasPrefix(strings.ToLower(v.Title), q) {
				suggestions = append(suggestions, v.Title)
			}
		}
	}
	if len(suggestions) == 0 {
		suggestions = []string{"Go 1.27 tutorial", "Vue 3 Nuxt 4", "HLS adaptive streaming", "Shorts viral"}
	}
	identity.WriteJSON(w, http.StatusOK, suggestions)
}
