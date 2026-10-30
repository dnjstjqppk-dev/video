package live

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"globalvideoplatform/services/api/internal/store"
)

func TestWebSocketLiveChatFullDuplex(t *testing.T) {
	memStore := store.NewMemoryStore()
	handler := NewHandler(memStore)

	stream := &store.LiveStream{
		ID:        "stream_ws_test",
		ChannelID: "ch_1",
		Title:     "Live WS Test",
		Status:    "live",
	}
	memStore.LiveStreams["stream_ws_test"] = stream

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.HandleWebSocketChat(w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "?id=stream_ws_test"

	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}
	defer ws.Close()

	// 1. Verify initial viewer count broadcast on connect
	_ = ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msgBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to receive initial viewer count broadcast: %v", err)
	}

	var initialMsg WSMessage
	if err := json.Unmarshal(msgBytes, &initialMsg); err != nil {
		t.Fatalf("failed to parse message: %v", err)
	}
	if initialMsg.Type != "viewer_count" {
		t.Fatalf("expected viewer_count event, got %s", initialMsg.Type)
	}

	// 2. Send chat message through WebSocket
	chatPayload := map[string]interface{}{
		"action": "send_chat",
		"payload": map[string]string{
			"message": "Hello from automated test!",
		},
	}
	payloadBytes, _ := json.Marshal(chatPayload)
	if err := ws.WriteMessage(websocket.TextMessage, payloadBytes); err != nil {
		t.Fatalf("failed to write websocket message: %v", err)
	}

	// 3. Receive broadcasted chat message
	_ = ws.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, responseBytes, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read broadcasted chat message: %v", err)
	}

	var chatMsg WSMessage
	if err := json.Unmarshal(responseBytes, &chatMsg); err != nil {
		t.Fatalf("failed to parse chat message: %v", err)
	}
	if chatMsg.Type != "chat_message" {
		t.Fatalf("expected chat_message event, got %s", chatMsg.Type)
	}
}

