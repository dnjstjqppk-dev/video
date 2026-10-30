package channels_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/channels"
)

func newChannelsTestServer() *httptest.Server {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	handler := channels.NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/channels", handler.Routes(auth))

	return httptest.NewServer(r)
}

func channelsReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var rbody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rbody = bytes.NewReader(b)
	} else {
		rbody = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rbody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, httptest.NewRecorder()
}

func TestGetChannel_ReturnsChannel(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, rec := channelsReq(http.MethodGet, "/channels/"+channelID, "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateChannel_RequiresAuth(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	req, rec := channelsReq(http.MethodPost, "/channels/", "", map[string]string{"name": "New Channel", "handle": "@newchannel"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestCreateChannel_Authenticated(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	req, rec := channelsReq(http.MethodPost, "/channels/", "creator-token", map[string]string{"name": "My Channel", "handle": "@mychannel"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateChannel_OwnerCanUpdate(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	// channel "aaaaaaaa-..." is owned by creator (2222...) per seedDefaults
	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, rec := channelsReq(http.MethodPatch, "/channels/"+channelID, "creator-token", map[string]string{"name": "Updated Channel Name"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateChannel_BOLA_OtherUserCannotUpdate(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	// advertiser (4444...) attempts to modify creator-owned channel -> 403
	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, rec := channelsReq(http.MethodPatch, "/channels/"+channelID, "advertiser-token", map[string]string{"name": "Hacked Channel"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 FORBIDDEN for cross-owner channel update, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateChannel_SuperAdminCanUpdate(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, rec := channelsReq(http.MethodPatch, "/channels/"+channelID, "admin-token", map[string]string{"name": "Admin Updated Channel"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for super_admin, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateChannel_NotFound(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()

	req, rec := channelsReq(http.MethodPatch, "/channels/nonexistent-id", "creator-token", map[string]string{"name": "Nope"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
