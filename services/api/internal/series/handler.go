package series

import (
	"encoding/json"
	"net/http"
	"sort"
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

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.ListSeries)
	r.Get("/{id}", h.GetSeries)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/", h.CreateSeries)
		pr.Patch("/{id}", h.UpdateSeries)
		pr.Delete("/{id}", h.DeleteSeries)
		pr.Post("/{id}/episodes", h.AttachEpisode)
		pr.Delete("/{id}/episodes/{videoId}", h.DetachEpisode)
	})

	return r
}

// ListSeries returns series, optionally filtered by ?channel_id=. Public.
func (h *Handler) ListSeries(w http.ResponseWriter, r *http.Request) {
	channelID := r.URL.Query().Get("channel_id")
	page, limit := identity.ParsePagination(r, 24)
	var out []*store.Series
	for _, s := range h.Store.Series {
		if channelID != "" && s.ChannelID != channelID {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	window, total := identity.PaginateSlice(out, page, limit)
	identity.WriteJSON(w, http.StatusOK, identity.PageEnvelope(window, total, page, limit))
}

// seriesDetail is the public series payload with its ordered episodes.
func (h *Handler) seriesDetail(s *store.Series) map[string]interface{} {
	eps := h.Store.EpisodesForSeries(s.ID)
	entries := make([]store.SeriesEpisode, 0, len(eps))
	for _, v := range eps {
		entries = append(entries, store.ToSeriesEpisode(v))
	}
	return map[string]interface{}{
		"series":        s,
		"episodes":      entries,
		"episode_count": len(entries),
	}
}

// GetSeries returns one series with its ordered episodes. Public.
func (h *Handler) GetSeries(w http.ResponseWriter, r *http.Request) {
	s, ok := h.Store.Series[chi.URLParam(r, "id")]
	if !ok || s == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Series not found")
		return
	}
	identity.WriteJSON(w, http.StatusOK, h.seriesDetail(s))
}

func (h *Handler) canManageSeries(r *http.Request, s *store.Series) bool {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		return false
	}
	if user.Role == "super_admin" {
		return true
	}
	channel, ok := h.Store.Channels[s.ChannelID]
	return ok && channel != nil && channel.OwnerID == user.ID
}

// CreateSeries makes a series on a channel the caller owns.
func (h *Handler) CreateSeries(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		ChannelID   string `json:"channel_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		CoverURL    string `json:"cover_url"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	channel, ok := h.Store.Channels[payload.ChannelID]
	if !ok || channel == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Channel not found")
		return
	}
	user := identity.GetUserFromContext(r.Context())
	if user == nil || (user.Role != "super_admin" && channel.OwnerID != user.ID) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot create a series on another user's channel")
		return
	}
	title := security.SanitizeUserText(payload.Title, 255)
	if title == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "title is required")
		return
	}
	status := payload.Status
	if status == "" {
		status = "ongoing"
	}
	if status != "ongoing" && status != "completed" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "status must be ongoing or completed")
		return
	}
	now := time.Now()
	s := &store.Series{
		ID:          uuid.NewString(),
		ChannelID:   channel.ID,
		Title:       title,
		Description: security.SanitizeUserText(payload.Description, 4000),
		CoverURL:    security.SanitizeUserText(payload.CoverURL, 2048),
		Status:      status,
		AiOriginal:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	h.Store.Series[s.ID] = s
	h.Store.PersistSeries(s)
	identity.WriteJSON(w, http.StatusCreated, h.seriesDetail(s))
}

// UpdateSeries edits title, description, cover, or status (owner only).
func (h *Handler) UpdateSeries(w http.ResponseWriter, r *http.Request) {
	s, ok := h.Store.Series[chi.URLParam(r, "id")]
	if !ok || s == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Series not found")
		return
	}
	if !h.canManageSeries(r, s) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another user's series")
		return
	}
	var payload struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		CoverURL    *string `json:"cover_url"`
		Status      *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	if payload.Title != nil {
		title := security.SanitizeUserText(*payload.Title, 255)
		if title == "" {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "title must not be empty")
			return
		}
		s.Title = title
	}
	if payload.Description != nil {
		s.Description = security.SanitizeUserText(*payload.Description, 4000)
	}
	if payload.CoverURL != nil {
		s.CoverURL = security.SanitizeUserText(*payload.CoverURL, 2048)
	}
	if payload.Status != nil {
		if *payload.Status != "ongoing" && *payload.Status != "completed" {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "status must be ongoing or completed")
			return
		}
		s.Status = *payload.Status
	}
	s.UpdatedAt = time.Now()
	h.Store.PersistSeries(s)
	identity.WriteJSON(w, http.StatusOK, h.seriesDetail(s))
}

// DeleteSeries removes a series and detaches its episodes (owner only).
func (h *Handler) DeleteSeries(w http.ResponseWriter, r *http.Request) {
	s, ok := h.Store.Series[chi.URLParam(r, "id")]
	if !ok || s == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Series not found")
		return
	}
	if !h.canManageSeries(r, s) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot delete another user's series")
		return
	}
	for _, v := range h.Store.EpisodesForSeries(s.ID) {
		v.SeriesID = ""
		v.SeasonNo = 0
		v.EpisodeNo = 0
		h.Store.PersistVideo(v)
	}
	delete(h.Store.Series, s.ID)
	h.Store.DeletePersistedSeries(s.ID)
	identity.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// AttachEpisode links a same-channel video into the series at the given
// season/episode slot. Re-attaching moves the video; a video lives in at
// most one series.
func (h *Handler) AttachEpisode(w http.ResponseWriter, r *http.Request) {
	s, ok := h.Store.Series[chi.URLParam(r, "id")]
	if !ok || s == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Series not found")
		return
	}
	if !h.canManageSeries(r, s) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another user's series")
		return
	}
	var payload struct {
		VideoID   string `json:"video_id"`
		SeasonNo  int    `json:"season_no"`
		EpisodeNo int    `json:"episode_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	video, ok := h.Store.Videos[payload.VideoID]
	if !ok || video == nil || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	if video.ChannelID != s.ChannelID {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Episode must belong to the series channel")
		return
	}
	if payload.SeasonNo < 1 || payload.EpisodeNo < 1 || payload.SeasonNo > 99 || payload.EpisodeNo > 999 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "season_no (1-99) and episode_no (1-999) are required")
		return
	}
	video.SeriesID = s.ID
	video.SeasonNo = payload.SeasonNo
	video.EpisodeNo = payload.EpisodeNo
	h.Store.PersistVideo(video)
	s.UpdatedAt = time.Now()
	h.Store.PersistSeries(s)
	identity.WriteJSON(w, http.StatusOK, h.seriesDetail(s))
}

// DetachEpisode unlinks a video from the series (owner only).
func (h *Handler) DetachEpisode(w http.ResponseWriter, r *http.Request) {
	s, ok := h.Store.Series[chi.URLParam(r, "id")]
	if !ok || s == nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Series not found")
		return
	}
	if !h.canManageSeries(r, s) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another user's series")
		return
	}
	video, ok := h.Store.Videos[chi.URLParam(r, "videoId")]
	if !ok || video == nil || video.SeriesID != s.ID {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Episode not found in this series")
		return
	}
	video.SeriesID = ""
	video.SeasonNo = 0
	video.EpisodeNo = 0
	h.Store.PersistVideo(video)
	s.UpdatedAt = time.Now()
	h.Store.PersistSeries(s)
	identity.WriteJSON(w, http.StatusOK, h.seriesDetail(s))
}
