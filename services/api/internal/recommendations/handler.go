package recommendations

import (
	"encoding/json"
	"net/http"
	"strings"

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

	r.Get("/home", h.GetHomeFeed)
	r.Get("/shorts", h.GetShortsFeed)
	r.Get("/subscriptions", h.GetSubscriptionsFeed)
	r.Post("/feedback", h.Feedback)

	return r
}

func (h *Handler) GetHomeFeed(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	var feed []*store.Video

	for _, v := range h.Store.Videos {
		if v.Visibility == "public" && v.Status == "READY" && v.VideoType != "shorts" {
			if category != "" && category != "all" && !strings.EqualFold(v.Category, category) {
				continue
			}
			feed = append(feed, v)
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": feed,
		"total": len(feed),
	})
}

func (h *Handler) GetShortsFeed(w http.ResponseWriter, r *http.Request) {
	var shorts []*store.Video
	for _, v := range h.Store.Videos {
		if v.Visibility == "public" && v.Status == "READY" && v.VideoType == "shorts" {
			shorts = append(shorts, v)
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": shorts,
		"total": len(shorts),
	})
}

func (h *Handler) GetSubscriptionsFeed(w http.ResponseWriter, r *http.Request) {
	var feed []*store.Video
	for _, v := range h.Store.Videos {
		if v.Visibility == "public" && v.Status == "READY" {
			feed = append(feed, v)
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": feed,
		"total": len(feed),
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
	var matchedVideos []*store.Video
	var matchedChannels []*store.Channel

	if query != "" {
		for _, v := range h.Store.Videos {
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

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"videos":   matchedVideos,
		"channels": matchedChannels,
		"total":    len(matchedVideos) + len(matchedChannels),
	})
}

func (h *Handler) SearchSuggestions(w http.ResponseWriter, r *http.Request) {
	q := strings.ToLower(r.URL.Query().Get("q"))
	suggestions := []string{}
	if q != "" {
		for _, v := range h.Store.Videos {
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
