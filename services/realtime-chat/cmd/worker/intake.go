package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// intake.go exposes the realtime-chat execution API: message publish with
// slow-mode/word-filter/ban enforcement, moderator ban/delete and room
// history reads. The API owns its inline live-chat WebSocket path; this
// service is the scale-out executor behind CHAT_WORKER_URL.

func writeIntakeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeIntakeError(w http.ResponseWriter, status int, code, message string) {
	writeIntakeJSON(w, status, map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

func newMessageID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "msg_" + hex.EncodeToString(b)
}

// ListMessages returns a copy of the stored room history oldest-first.
func (h *ChatHub) ListMessages(roomID string) []*ChatMessage {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]*ChatMessage, len(h.Messages[roomID]))
	copy(out, h.Messages[roomID])
	return out
}

type publishMessageRequest struct {
	MessageID     string `json:"messageId"`
	UserID        string `json:"userId"`
	Content       string `json:"content"`
	IsSuperChat   bool   `json:"isSuperChat"`
	AmountCents   int64  `json:"amountCents"`
	Authenticated bool   `json:"authenticated"`
}

type banUserRequest struct {
	UserID string `json:"userId"`
}

func chatMessageResponse(m *ChatMessage) map[string]interface{} {
	return map[string]interface{}{
		"messageId":       m.ID,
		"roomId":          m.RoomID,
		"userId":          m.UserID,
		"content":         m.Content,
		"isSuperChat":     m.IsSuperChat,
		"amountCents":     m.AmountCents,
		"pinDurationSecs": m.PinDurationSecs,
		"deleted":         m.Deleted,
		"createdAt":       m.CreatedAt.Format(time.RFC3339),
	}
}

func sendErrorStatus(err error) (int, string) {
	switch {
	case strings.Contains(err.Error(), "unauthenticated"):
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	case strings.Contains(err.Error(), "banned"):
		return http.StatusForbidden, "BANNED"
	case strings.Contains(err.Error(), "slow-mode"):
		return http.StatusTooManyRequests, "SLOW_MODE"
	default:
		return http.StatusBadRequest, "REJECTED"
	}
}

func (h *healthServer) handlePublishMessage(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomId")
	var req publishMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	if req.UserID == "" || req.Content == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "userId and content are required")
		return
	}
	msgID := req.MessageID
	if msgID == "" {
		msgID = newMessageID()
	}
	msg := &ChatMessage{
		ID:          msgID,
		RoomID:      roomID,
		UserID:      req.UserID,
		Content:     req.Content,
		IsSuperChat: req.IsSuperChat,
		AmountCents: req.AmountCents,
	}
	if err := h.hub.SendMessage(msg, req.Authenticated); err != nil {
		status, code := sendErrorStatus(err)
		writeIntakeError(w, status, code, err.Error())
		return
	}
	writeIntakeJSON(w, http.StatusCreated, chatMessageResponse(msg))
}

func (h *healthServer) handleListMessages(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomId")
	msgs := h.hub.ListMessages(roomID)
	out := make([]map[string]interface{}, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, chatMessageResponse(m))
	}
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"roomId": roomID, "messages": out})
}

func (h *healthServer) handleBanUser(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomId")
	var req banUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "userId is required")
		return
	}
	h.hub.ModeratorBanUser(roomID, req.UserID)
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"roomId": roomID, "userId": req.UserID, "banned": true})
}

func (h *healthServer) handleDeleteMessage(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("roomId")
	messageID := r.PathValue("messageId")
	if !h.hub.ModeratorDeleteMessage(roomID, messageID) {
		writeIntakeError(w, http.StatusNotFound, "NOT_FOUND", "Chat message not found")
		return
	}
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"roomId": roomID, "messageId": messageID, "deleted": true})
}

func (h *healthServer) registerIntakeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /rooms/{roomId}/messages", h.handlePublishMessage)
	mux.HandleFunc("GET /rooms/{roomId}/messages", h.handleListMessages)
	mux.HandleFunc("POST /rooms/{roomId}/ban", h.handleBanUser)
	mux.HandleFunc("POST /rooms/{roomId}/messages/{messageId}/delete", h.handleDeleteMessage)
}
