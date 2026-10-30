package subscriptions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func TestSubscriptionsLifecycle(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	// 1. Subscribe
	reqSub := httptest.NewRequest(http.MethodPost, "/"+channelID, nil)
	reqSub.Header.Set("Authorization", "Bearer viewer-token")
	recSub := httptest.NewRecorder()
	router.ServeHTTP(recSub, reqSub)
	if recSub.Code != http.StatusOK {
		t.Fatalf("expected 200 OK subscribing, got %d", recSub.Code)
	}

	// 2. List subscriptions
	reqList := httptest.NewRequest(http.MethodGet, "/", nil)
	reqList.Header.Set("Authorization", "Bearer viewer-token")
	recList := httptest.NewRecorder()
	router.ServeHTTP(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("expected 200 OK listing subscriptions, got %d", recList.Code)
	}

	// 3. Unsubscribe
	reqUnsubscribe := httptest.NewRequest(http.MethodDelete, "/"+channelID, nil)
	reqUnsubscribe.Header.Set("Authorization", "Bearer viewer-token")
	recUnsubscribe := httptest.NewRecorder()
	router.ServeHTTP(recUnsubscribe, reqUnsubscribe)
	if recUnsubscribe.Code != http.StatusOK {
		t.Fatalf("expected 200 OK unsubscribing, got %d", recUnsubscribe.Code)
	}
}
