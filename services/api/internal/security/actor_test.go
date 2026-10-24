package security

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestActorBoundToMethodAndPath(t *testing.T) {
	key := []byte(strings.Repeat("k", 40))
	r := httptest.NewRequest("POST", "https://api.test/v1/videos", nil)
	a := Actor{UserID: "u1", Audience: "catalog", Method: "POST", Path: "/v1/videos", Expires: time.Now().Add(time.Minute).Unix()}
	token := Sign(key, a)
	if _, err := Verify(key, token, r); err != nil {
		t.Fatal(err)
	}
	r = httptest.NewRequest("DELETE", "https://api.test/v1/videos", nil)
	if _, err := Verify(key, token, r); err == nil {
		t.Fatal("method-changing replay accepted")
	}
	r = httptest.NewRequest("POST", "https://api.test/v1/channels", nil)
	if _, err := Verify(key, token, r); err == nil {
		t.Fatal("cross-resource replay accepted")
	}
	if _, err := Verify([]byte(strings.Repeat("x", 40)), token, httptest.NewRequest("POST", "https://api.test/v1/videos", nil)); err == nil {
		t.Fatal("forged token accepted")
	}
}
func TestActorExpiry(t *testing.T) {
	key := []byte(strings.Repeat("a", 40))
	r := httptest.NewRequest("POST", "/v1/videos", nil)
	a := Actor{UserID: "u1", Audience: "catalog", Method: "POST", Path: "/v1/videos", Expires: time.Now().Add(-time.Minute).Unix()}
	if _, err := Verify(key, Sign(key, a), r); err == nil {
		t.Fatal("expired assertion accepted")
	}
}
