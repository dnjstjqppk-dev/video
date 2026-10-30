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
	r.Get("/{id}/health", h.GetStreamHealth)
	r.Get("/{id}/superchats", h.GetSuperChats)
	r.Get("/{id}/chat", h.GetLiveChatMessages)
	r.Get("/{id}/polls/active", h.GetActiveLivePoll)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.RequireAuth)
		pr.Post("/", h.CreateStream)
		pr.Post("/{id}/start", h.StartStream)
		pr.Post("/{id}/stop", h.StopStream)
		pr.Post("/{id}/rotate-key", h.RotateKey)
		pr.Post("/{id}/superchat", h.CreateSuperChat)
		pr.Post("/{id}/chat", h.SendLiveChatMessage)
		pr.Post("/{id}/polls", h.CreateLivePoll)
		pr.Post("/{id}/polls/vote", h.VoteLivePoll)
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

func (h *Handler) CreateSuperChat(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	_, ok := h.Store.LiveStreams[streamID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}

	var payload struct {
		AmountCents int64  `json:"amount_cents"`
		Currency    string `json:"currency"`
		Message     string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.AmountCents <= 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_INPUT", "Valid amount_cents required")
		return
	}

	if payload.Currency == "" {
		payload.Currency = "USD"
	}

	tierColor, pinDurationSecs := calculateSuperChatTier(payload.AmountCents)
	now := time.Now()

	sc := &store.SuperChat{
		ID:              uuid.NewString(),
		StreamID:        streamID,
		UserID:          user.ID,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		AmountCents:     payload.AmountCents,
		Currency:        payload.Currency,
		Message:         payload.Message,
		TierColor:       tierColor,
		PinDurationSecs: pinDurationSecs,
		ExpiresAt:       now.Add(time.Duration(pinDurationSecs) * time.Second),
		CreatedAt:       now,
	}

	h.Store.SuperChats[streamID] = append(h.Store.SuperChats[streamID], sc)

	// Also insert into chat stream
	chatItem := &store.LiveChatMessageItem{
		ID:          uuid.NewString(),
		StreamID:    streamID,
		UserID:      user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Message:     payload.Message,
		SuperChat:   sc,
		CreatedAt:   now,
	}
	h.Store.LiveChat[streamID] = append(h.Store.LiveChat[streamID], chatItem)

	identity.WriteJSON(w, http.StatusCreated, sc)
}

func (h *Handler) GetSuperChats(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	all := h.Store.SuperChats[streamID]
	active := []*store.SuperChat{}
	now := time.Now()

	for _, sc := range all {
		if sc.ExpiresAt.After(now) {
			active = append(active, sc)
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": active,
		"total": len(active),
	})
}

func (h *Handler) SendLiveChatMessage(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	stream, ok := h.Store.LiveStreams[streamID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}

	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}

	var payload struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Message == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_INPUT", "Message text required")
		return
	}

	isCreator := false
	if ch, ok := h.Store.Channels[stream.ChannelID]; ok && ch.OwnerID == user.ID {
		isCreator = true
	}

	chatItem := &store.LiveChatMessageItem{
		ID:          uuid.NewString(),
		StreamID:    streamID,
		UserID:      user.ID,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Message:     payload.Message,
		IsModerator: user.Role == "admin",
		IsCreator:   isCreator,
		CreatedAt:   time.Now(),
	}

	h.Store.LiveChat[streamID] = append(h.Store.LiveChat[streamID], chatItem)
	identity.WriteJSON(w, http.StatusCreated, chatItem)
}

func (h *Handler) GetLiveChatMessages(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	messages := h.Store.LiveChat[streamID]
	if messages == nil {
		messages = []*store.LiveChatMessageItem{}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"items": messages,
		"total": len(messages),
	})
}

func (h *Handler) GetStreamHealth(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	stream, ok := h.Store.LiveStreams[streamID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Live stream not found")
		return
	}

	status := "EXCELLENT"
	if stream.Status != "live" {
		status = "OFFLINE"
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"stream_id":                     streamID,
		"status":                        status,
		"ingest_protocol":               "RTMPS",
		"bitrate_kbps":                  6800,
		"fps":                           60,
		"keyframe_interval_seconds":     2.0,
		"audio_sample_rate_hz":          48000,
		"packet_loss_pct":               0.0,
		"last_keyframe_received_ms_ago": 85,
	})
}

func (h *Handler) CreateLivePoll(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	var payload struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Question == "" || len(payload.Options) < 2 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_INPUT", "Question and at least 2 options required")
		return
	}

	opts := make([]store.LivePollOption, len(payload.Options))
	for i, opt := range payload.Options {
		opts[i] = store.LivePollOption{
			Index: i,
			Text:  opt,
			Votes: 0,
		}
	}

	poll := &store.LivePoll{
		ID:         uuid.NewString(),
		StreamID:   streamID,
		Question:   payload.Question,
		Options:    opts,
		IsActive:   true,
		TotalVotes: 0,
		CreatedAt:  time.Now(),
	}

	h.Store.LivePolls[streamID] = poll
	identity.WriteJSON(w, http.StatusCreated, poll)
}

func (h *Handler) VoteLivePoll(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	poll := h.Store.LivePolls[streamID]
	if poll == nil || !poll.IsActive {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "No active live poll for stream")
		return
	}

	var payload struct {
		OptionIndex int `json:"option_index"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.OptionIndex < 0 || payload.OptionIndex >= len(poll.Options) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_INPUT", "Invalid option_index")
		return
	}

	poll.Options[payload.OptionIndex].Votes++
	poll.TotalVotes++

	identity.WriteJSON(w, http.StatusOK, poll)
}

func (h *Handler) GetActiveLivePoll(w http.ResponseWriter, r *http.Request) {
	streamID := chi.URLParam(r, "id")
	poll := h.Store.LivePolls[streamID]
	if poll == nil {
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{"poll": nil})
		return
	}
	identity.WriteJSON(w, http.StatusOK, poll)
}

func calculateSuperChatTier(amountCents int64) (tierColor string, pinDurationSecs int) {
	switch {
	case amountCents >= 10000: // $100+
		return "red", 3600
	case amountCents >= 5000: // $50+
		return "orange", 1800
	case amountCents >= 2000: // $20+
		return "yellow", 600
	case amountCents >= 1000: // $10+
		return "green", 300
	case amountCents >= 500: // $5+
		return "cyan", 120
	case amountCents >= 200: // $2+
		return "blue", 60
	default:
		return "blue", 30
	}
}
