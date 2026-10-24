package main

import (
	"testing"
	"time"
)

func TestRealtimeChatSlowModeModerationAndSuperChat(t *testing.T) {
	hub := NewChatHub(500*time.Millisecond, []string{"spamword"})

	// 1. Reject unauthenticated
	if err := hub.SendMessage(&ChatMessage{ID: "m0", RoomID: "r1", UserID: "u1", Content: "hi"}, false); err == nil {
		t.Fatal("expected unauthenticated message to be rejected")
	}

	// 2. First valid message succeeds
	if err := hub.SendMessage(&ChatMessage{ID: "m1", RoomID: "r1", UserID: "u1", Content: "hello stream!"}, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 3. Immediate second message triggers slow-mode rate limit
	if err := hub.SendMessage(&ChatMessage{ID: "m2", RoomID: "r1", UserID: "u1", Content: "second fast msg"}, true); err == nil {
		t.Fatal("expected slow-mode rate limit error")
	}

	// 4. Banned word filter
	if err := hub.SendMessage(&ChatMessage{ID: "m3", RoomID: "r1", UserID: "u2", Content: "buy spamword now"}, true); err == nil {
		t.Fatal("expected banned word rejection")
	}

	// 5. SuperChat bypasses slow mode and calculates pin duration
	sc := &ChatMessage{ID: "sc1", RoomID: "r1", UserID: "u1", Content: "Great stream!", IsSuperChat: true, AmountCents: 10000}
	if err := hub.SendMessage(sc, true); err != nil || sc.PinDurationSecs != 300 {
		t.Fatalf("expected SuperChat pinned for 300s, got pin=%d err=%v", sc.PinDurationSecs, err)
	}

	// 6. Moderator ban & delete
	if !hub.ModeratorDeleteMessage("r1", "m1") {
		t.Fatal("expected message m1 to be marked deleted")
	}
	hub.ModeratorBanUser("r1", "u3")
	if err := hub.SendMessage(&ChatMessage{ID: "m4", RoomID: "r1", UserID: "u3", Content: "test"}, true); err == nil {
		t.Fatal("expected banned user message to be rejected")
	}
}
