package store

import (
	"net/http/httptest"
	"testing"
	"time"
)

// P1-14: one viewer counts once per video per window.
func TestCountViewDedups(t *testing.T) {
	s := NewMemoryStore()
	if !s.CountView("user:a", "vid-1") {
		t.Fatal("first view must count")
	}
	if s.CountView("user:a", "vid-1") {
		t.Fatal("repeat view inside the window must not count")
	}
	if !s.CountView("user:a", "vid-2") {
		t.Fatal("other video must count")
	}
	if !s.CountView("user:b", "vid-1") {
		t.Fatal("other viewer must count")
	}
	if s.CountView("", "vid-1") || s.CountView("user:a", "") {
		t.Fatal("empty identity must not count")
	}
}

// P1-14: the window expires and counting resumes.
func TestCountViewWindowExpiry(t *testing.T) {
	old := ViewDedupWindow
	ViewDedupWindow = 5 * time.Millisecond
	defer func() { ViewDedupWindow = old }()

	s := NewMemoryStore()
	if !s.CountView("user:a", "vid-1") {
		t.Fatal("first view must count")
	}
	time.Sleep(10 * time.Millisecond)
	if !s.CountView("user:a", "vid-1") {
		t.Fatal("view after window expiry must count")
	}
}

// P1-14: anonymous fingerprints are stable per client, distinct per client.
func TestViewerKey(t *testing.T) {
	mk := func(remote, ua string) string {
		r := httptest.NewRequest("GET", "/x", nil)
		r.RemoteAddr = remote
		r.Header.Set("User-Agent", ua)
		return ViewerKey(r, "")
	}
	if got := ViewerKey(httptest.NewRequest("GET", "/x", nil), "u1"); got != "user:u1" {
		t.Fatalf("authed key wrong: %q", got)
	}
	if mk("1.1.1.1:1", "A") != mk("1.1.1.1:1", "A") {
		t.Fatal("same client must map to the same key")
	}
	if mk("1.1.1.1:1", "A") == mk("2.2.2.2:2", "A") {
		t.Fatal("different clients must map apart")
	}
}
