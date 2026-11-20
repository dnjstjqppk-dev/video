package admin

import (
	"net/http"
	"sort"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

// ListAIReviews returns the Originals review queue: AI-made videos with a
// pending review, oldest first so nothing waits silently. Staff only.
func (h *Handler) ListAIReviews(w http.ResponseWriter, r *http.Request) {
	page, limit := identity.ParsePagination(r, 24)
	var pending []*store.Video
	for _, v := range h.Store.Videos {
		if v == nil || !v.AiGenerated || v.AiReviewStatus != "pending" || v.Status == "DELETED" {
			continue
		}
		pending = append(pending, v)
	}
	sort.Slice(pending, func(i, j int) bool {
		if pending[i].CreatedAt.Equal(pending[j].CreatedAt) {
			return pending[i].ID < pending[j].ID
		}
		return pending[i].CreatedAt.Before(pending[j].CreatedAt)
	})
	queue := make([]map[string]interface{}, 0, len(pending))
	for _, v := range pending {
		item := map[string]interface{}{"video": v}
		if ch := h.Store.Channels[v.ChannelID]; ch != nil {
			item["channel_name"] = ch.Name
			item["channel_handle"] = ch.Handle
		}
		if v.SeriesID != "" {
			if s := h.Store.Series[v.SeriesID]; s != nil {
				item["series_id"] = s.ID
				item["series_title"] = s.Title
			}
		}
		queue = append(queue, item)
	}
	window, total := identity.PaginateSlice(queue, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}
