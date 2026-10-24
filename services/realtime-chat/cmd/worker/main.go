package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ChatMessage struct {
	ID              string
	RoomID          string
	UserID          string
	Content         string
	IsSuperChat     bool
	AmountCents     int64
	PinDurationSecs int
	Deleted         bool
	CreatedAt       time.Time
}

type ChatHub struct {
	mu              sync.RWMutex
	Messages        map[string][]*ChatMessage
	BannedUsers     map[string]map[string]bool // roomID -> userID -> banned
	LastMessageTime map[string]map[string]time.Time
	SlowModeDelay   time.Duration
	BannedWords     []string
}

func NewChatHub(slowModeDelay time.Duration, bannedWords []string) *ChatHub {
	return &ChatHub{
		Messages:        make(map[string][]*ChatMessage),
		BannedUsers:     make(map[string]map[string]bool),
		LastMessageTime: make(map[string]map[string]time.Time),
		SlowModeDelay:   slowModeDelay,
		BannedWords:     bannedWords,
	}
}

func (h *ChatHub) SendMessage(msg *ChatMessage, isAuthenticated bool) error {
	if !isAuthenticated || msg.UserID == "" {
		return errors.New("unauthenticated user cannot send chat messages")
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.BannedUsers[msg.RoomID] != nil && h.BannedUsers[msg.RoomID][msg.UserID] {
		return errors.New("user is banned from this live chat room")
	}

	lower := strings.ToLower(msg.Content)
	for _, bw := range h.BannedWords {
		if strings.Contains(lower, strings.ToLower(bw)) {
			return errors.New("message rejected by automated word filter")
		}
	}

	now := time.Now()
	if h.LastMessageTime[msg.RoomID] == nil {
		h.LastMessageTime[msg.RoomID] = make(map[string]time.Time)
	}
	if last, exists := h.LastMessageTime[msg.RoomID][msg.UserID]; exists && !msg.IsSuperChat {
		if now.Sub(last) < h.SlowModeDelay {
			return errors.New("slow-mode rate limit exceeded")
		}
	}
	h.LastMessageTime[msg.RoomID][msg.UserID] = now

	if msg.IsSuperChat {
		msg.PinDurationSecs = SuperChatPinSeconds(msg.AmountCents)
	}

	msg.CreatedAt = now
	h.Messages[msg.RoomID] = append(h.Messages[msg.RoomID], msg)
	return nil
}

func SuperChatPinSeconds(amountCents int64) int {
	switch {
	case amountCents >= 10000:
		return 300
	case amountCents >= 5000:
		return 120
	case amountCents >= 1000:
		return 60
	default:
		return 30
	}
}

func (h *ChatHub) ModeratorBanUser(roomID, targetUserID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.BannedUsers[roomID] == nil {
		h.BannedUsers[roomID] = make(map[string]bool)
	}
	h.BannedUsers[roomID][targetUserID] = true
}

func (h *ChatHub) ModeratorDeleteMessage(roomID, messageID string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, m := range h.Messages[roomID] {
		if m.ID == messageID {
			m.Deleted = true
			return true
		}
	}
	return false
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Info("Starting Realtime Live Chat WebSocket Hub")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hub := NewChatHub(2*time.Second, []string{"scam_link", "phishing"})
	_ = hub

	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
	}
}
