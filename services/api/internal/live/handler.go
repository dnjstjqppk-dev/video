package live

import (
	"encoding/json"
	"fmt"
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
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()

	r.Get("/{id}", h.GetStream)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/", h.CreateStream)
		pr.Post("/{id}/start", h.StartStream)
		pr.Post("/{id}/stop", h.StopStream)
		pr.Post("/{id}/rotate-key", h.RotateKey)
	})

	return r
}

func (h *Handler) CreateStream(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(r.Body).Decode(&payload)

	streamID := uuid.NewString()
	streamKey := fmt.Sprintf("live_%s_%d", uuid.NewString()[:8], time.Now().Unix())

	stream := &store.LiveStream{
		ID:             streamID,
		ChannelID:      "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		Title:          payload.Title,
		Description:    payload.Description,
		ThumbnailURL:   "https://images.unsplash.com/photo-1511671782779-c97d3d27a1d4?w=800",
		Status:         "idle",
		StreamKey:      streamKey,
		RtmpsIngestURL: "rtmps://live.cloudflare.com:443/live/",
		SrtIngestURL:   "srt://live.cloudflare.com:10000",
		PlaybackHlsURL: "https://customer-test.cloudflarestream.com/live-hls/manifest/video.m3u8",
		ViewerCount:    0,
		CreatedAt:      time.Now(),
	}

	h.Store.LiveStreams[streamID] = stream
	identity.WriteJSON(w, http.StatusCreated, stream)
}

func (h *Handler) GetStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stream, ok := h.Store.LiveStreams[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}
	identity.WriteJSON(w, http.StatusOK, stream)
}

func (h *Handler) StartStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stream, ok := h.Store.LiveStreams[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}
	stream.Status = "live"
	identity.WriteJSON(w, http.StatusOK, stream)
}

func (h *Handler) StopStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stream, ok := h.Store.LiveStreams[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}
	stream.Status = "ended"
	identity.WriteJSON(w, http.StatusOK, stream)
}

func (h *Handler) RotateKey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	stream, ok := h.Store.LiveStreams[id]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}
	stream.StreamKey = fmt.Sprintf("live_%s_%d", uuid.NewString()[:8], time.Now().Unix())
	identity.WriteJSON(w, http.StatusOK, map[string]string{"streamKey": stream.StreamKey})
}
