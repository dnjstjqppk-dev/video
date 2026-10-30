package live

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all web/mobile origins for live streaming
	},
}

type WSMessage struct {
	Type      string      `json:"type"`      // chat_message, super_chat, poll_update, viewer_count, stream_status
	RoomID    string      `json:"room_id"`   // streamID
	Payload   interface{} `json:"payload"`   // dynamic event data
	Timestamp string      `json:"timestamp"` // ISO 8601
}

type Client struct {
	Hub         *LiveChatHub
	Conn        *websocket.Conn
	Send        chan []byte
	RoomID      string
	UserID      string
	DisplayName string
	IsModerator bool
	IsCreator   bool
}

type LiveChatHub struct {
	mu          sync.RWMutex
	Rooms       map[string]map[*Client]bool
	Broadcast   chan *WSMessage
	Register    chan *Client
	Unregister  chan *Client
	ViewerCount map[string]int
}

func NewLiveChatHub() *LiveChatHub {
	hub := &LiveChatHub{
		Rooms:       make(map[string]map[*Client]bool),
		Broadcast:   make(chan *WSMessage, 256),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		ViewerCount: make(map[string]int),
	}
	go hub.Run()
	return hub
}

func (h *LiveChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Rooms[client.RoomID] == nil {
				h.Rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.Rooms[client.RoomID][client] = true
			h.ViewerCount[client.RoomID] = len(h.Rooms[client.RoomID])
			count := h.ViewerCount[client.RoomID]
			h.mu.Unlock()

			// Broadcast updated viewer count to room
			h.BroadcastToRoom(client.RoomID, "viewer_count", map[string]interface{}{
				"stream_id":    client.RoomID,
				"viewer_count": count,
			})

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Rooms[client.RoomID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					h.ViewerCount[client.RoomID] = len(clients)
					if len(clients) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}
			count := h.ViewerCount[client.RoomID]
			h.mu.Unlock()

			h.BroadcastToRoom(client.RoomID, "viewer_count", map[string]interface{}{
				"stream_id":    client.RoomID,
				"viewer_count": count,
			})

		case message := <-h.Broadcast:
			data, err := json.Marshal(message)
			if err != nil {
				continue
			}
			h.mu.RLock()
			clients := h.Rooms[message.RoomID]
			for client := range clients {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *LiveChatHub) BroadcastToRoom(roomID string, msgType string, payload interface{}) {
	msg := &WSMessage{
		Type:      msgType,
		RoomID:    roomID,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	select {
	case h.Broadcast <- msg:
	default:
		slog.Warn("WebSocket broadcast channel full, message dropped", "room", roomID, "type", msgType)
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(4096)
	_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming struct {
			Action  string          `json:"action"` // send_chat, send_superchat, vote_poll
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}

		switch incoming.Action {
		case "send_chat":
			var chat struct {
				Message        string  `json:"message"`
				DonationAmount float64 `json:"donation_amount"`
				Currency       string  `json:"currency"`
			}
			if err := json.Unmarshal(incoming.Payload, &chat); err == nil && chat.Message != "" {
				msgType := "chat_message"
				if chat.DonationAmount > 0 {
					msgType = "super_chat"
				}
				if chat.Currency == "" {
					chat.Currency = "KRW"
				}
				c.Hub.BroadcastToRoom(c.RoomID, msgType, map[string]interface{}{
					"id":                uuid.NewString(),
					"stream_id":         c.RoomID,
					"user_id":           c.UserID,
					"user_display_name": c.DisplayName,
					"is_moderator":      c.IsModerator,
					"is_creator":        c.IsCreator,
					"message":           chat.Message,
					"donation_amount":   chat.DonationAmount,
					"currency":          chat.Currency,
					"timestamp":         time.Now().UTC().Format(time.RFC3339),
				})
			}
		case "ping":
			_ = c.Conn.WriteMessage(websocket.PongMessage, nil)
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWebSocket upgrades an HTTP connection to a full-duplex live streaming chat WebSocket.
func (h *LiveChatHub) ServeWebSocket(w http.ResponseWriter, r *http.Request, roomID string, userID string, displayName string, isMod bool, isCreator bool) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade WebSocket", "error", err)
		return
	}

	if userID == "" {
		userID = "anon_" + uuid.NewString()[:8]
	}
	if displayName == "" {
		displayName = "Viewer " + userID[min(5, len(userID)):]
	}

	client := &Client{
		Hub:         h,
		Conn:        conn,
		Send:        make(chan []byte, 256),
		RoomID:      roomID,
		UserID:      userID,
		DisplayName: displayName,
		IsModerator: isMod,
		IsCreator:   isCreator,
	}

	client.Hub.Register <- client

	go client.writePump()
	go client.readPump()
}
