package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newIntakeMux() *http.ServeMux {
	srv := newHealthServer("realtime-chat", NewChatHub(50*time.Millisecond, []string{"scam_link"}))
	mux := http.NewServeMux()
	srv.registerIntakeRoutes(mux)
	return mux
}

func doIntake(t *testing.T, mux *http.ServeMux, method, path, body string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var decoded map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &decoded)
	return rec.Code, decoded
}

func TestIntakePublishListDelete(t *testing.T) {
	mux := newIntakeMux()

	code, sent := doIntake(t, mux, http.MethodPost, "/rooms/room-1/messages",
		`{"userId":"u-1","content":"hello live chat","authenticated":true}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for publish, got %d: %v", code, sent)
	}
	msgID, _ := sent["messageId"].(string)
	if msgID == "" || sent["roomId"] != "room-1" {
		t.Fatalf("invalid publish response: %v", sent)
	}

	code, listed := doIntake(t, mux, http.MethodGet, "/rooms/room-1/messages", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for history, got %d", code)
	}
	msgs, _ := listed["messages"].([]interface{})
	if len(msgs) != 1 {
		t.Fatalf("expected 1 stored message, got %v", listed)
	}

	code, deleted := doIntake(t, mux, http.MethodPost, "/rooms/room-1/messages/"+msgID+"/delete", `{}`)
	if code != http.StatusOK || deleted["deleted"] != true {
		t.Fatalf("expected delete confirmation, got %d: %v", code, deleted)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/rooms/room-1/messages/missing/delete", `{}`)
	if code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown message delete, got %d", code)
	}
}

func TestIntakeModerationEnforcement(t *testing.T) {
	mux := newIntakeMux()

	// Unauthenticated publish is rejected.
	code, _ := doIntake(t, mux, http.MethodPost, "/rooms/room-2/messages",
		`{"userId":"u-2","content":"hi","authenticated":false}`)
	if code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated publish, got %d", code)
	}
	// Banned-word filter rejects the message.
	code, _ = doIntake(t, mux, http.MethodPost, "/rooms/room-2/messages",
		`{"userId":"u-2","content":"click this scam_link now","authenticated":true}`)
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400 for filtered content, got %d", code)
	}
	// Slow mode: immediate second message from the same user is limited.
	code, _ = doIntake(t, mux, http.MethodPost, "/rooms/room-2/messages",
		`{"userId":"u-3","content":"first","authenticated":true}`)
	if code != http.StatusCreated {
		t.Fatalf("expected 201 for first message, got %d", code)
	}
	code, _ = doIntake(t, mux, http.MethodPost, "/rooms/room-2/messages",
		`{"userId":"u-3","content":"second","authenticated":true}`)
	if code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for slow-mode violation, got %d", code)
	}
	// SuperChats bypass slow mode and earn pin seconds.
	code, sc := doIntake(t, mux, http.MethodPost, "/rooms/room-2/messages",
		`{"userId":"u-3","content":"take my support","isSuperChat":true,"amountCents":5000,"authenticated":true}`)
	if code != http.StatusCreated || sc["pinDurationSecs"] != float64(120) {
		t.Fatalf("expected pinned superchat, got %d: %v", code, sc)
	}
	// Moderator ban blocks the user afterwards.
	code, _ = doIntake(t, mux, http.MethodPost, "/rooms/room-2/ban", `{"userId":"u-3"}`)
	if code != http.StatusOK {
		t.Fatalf("expected 200 for ban, got %d", code)
	}
	time.Sleep(60 * time.Millisecond)
	code, _ = doIntake(t, mux, http.MethodPost, "/rooms/room-2/messages",
		`{"userId":"u-3","content":"let me back in","authenticated":true}`)
	if code != http.StatusForbidden {
		t.Fatalf("expected 403 for banned user, got %d", code)
	}
}
