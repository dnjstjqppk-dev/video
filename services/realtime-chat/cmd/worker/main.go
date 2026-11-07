package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
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

	health := newHealthServer("realtime-chat", hub)
	go func() {
		if err := health.listenAndServe(ctx); err != nil {
			slog.Error("Health server failure", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("Shutting down Realtime Chat gracefully")
}

// healthServer exposes liveness, readiness and Prometheus metrics so the
// worker is observable behind Docker/Kubernetes health checks.
type healthServer struct {
	name  string
	start time.Time
	hub   *ChatHub
}

func newHealthServer(name string, hub *ChatHub) *healthServer {
	return &healthServer{name: name, start: time.Now(), hub: hub}
}

func (h *healthServer) listenAndServe(ctx context.Context) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"UP","worker":%q,"uptime_seconds":%d}`, h.name, int64(time.Since(h.start).Seconds()))
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"READY","worker":%q}`, h.name)
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		fmt.Fprintf(w, "# HELP gvp_worker_uptime_seconds Worker uptime in seconds\n# TYPE gvp_worker_uptime_seconds gauge\ngvp_worker_uptime_seconds{worker=%q} %d\n", h.name, int64(time.Since(h.start).Seconds()))
	})
	h.registerIntakeRoutes(mux)
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	slog.Info("Worker health server listening", "worker", h.name, "port", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
