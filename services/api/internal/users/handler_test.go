package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/users"
)

func newUsersTestServer() *httptest.Server {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	handler := users.NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/users", handler.Routes(auth))

	return httptest.NewServer(r)
}

func usersReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
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

func TestGetMe_RequiresAuth(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	req, rec := usersReq(http.MethodGet, "/users/me", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestGetMe_Authenticated(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	req, rec := usersReq(http.MethodGet, "/users/me", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateMe_Authenticated(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	req, rec := usersReq(http.MethodPatch, "/users/me", "creator-token", map[string]string{"display_name": "Updated Display Name"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSessions_Authenticated(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	req, rec := usersReq(http.MethodGet, "/users/me/sessions", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestClearHistory_Authenticated(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	req, rec := usersReq(http.MethodDelete, "/users/me/history", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestMFAWorkflow_SetupVerifyDisable(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	// 1. Initial MFA status
	req, rec := usersReq(http.MethodGet, "/users/me/mfa", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for MFA status, got %d", rec.Code)
	}

	// 2. Setup MFA -> returns secret & backup codes
	req, rec = usersReq(http.MethodPost, "/users/me/mfa/setup", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for MFA setup, got %d: %s", rec.Code, rec.Body.String())
	}
	var setupResp struct {
		Secret      string   `json:"secret"`
		OtpauthURI  string   `json:"otpauth_uri"`
		BackupCodes []string `json:"backup_codes"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &setupResp); err != nil || setupResp.Secret == "" || len(setupResp.BackupCodes) == 0 {
		t.Fatalf("failed to decode setup resp: %v", err)
	}

	// 3. Verify with bad code -> fails 401
	req, rec = usersReq(http.MethodPost, "/users/me/mfa/verify", "creator-token", map[string]string{"code": "000000"})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid code, got %d", rec.Code)
	}

	// 4. Disable with valid backup code
	req, rec = usersReq(http.MethodPost, "/users/me/mfa/disable", "creator-token", map[string]string{"code": setupResp.BackupCodes[0]})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for disable MFA with backup code, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBlockedUsersWorkflow(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	// 1. Block a target user
	targetID := "33333333-3333-3333-3333-333333333333"
	req, rec := usersReq(http.MethodPost, "/users/me/blocked-users", "creator-token", map[string]string{"target_user_id": targetID})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for blocking user, got %d: %s", rec.Code, rec.Body.String())
	}

	// 2. Query blocked users
	req, rec = usersReq(http.MethodGet, "/users/me/blocked-users", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for listing blocked users, got %d", rec.Code)
	}
	var listResp struct {
		BlockedUsers []string `json:"blocked_users"`
		Count        int      `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &listResp)
	if listResp.Count != 1 || listResp.BlockedUsers[0] != targetID {
		t.Fatalf("expected 1 blocked user, got %d", listResp.Count)
	}

	// 3. Unblock the user
	req, rec = usersReq(http.MethodDelete, "/users/me/blocked-users/"+targetID, "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for unblock user, got %d", rec.Code)
	}
}

func TestStrikesAndGDPRDeletion(t *testing.T) {
	srv := newUsersTestServer()
	defer srv.Close()

	// 1. Get strikes
	req, rec := usersReq(http.MethodGet, "/users/me/strikes", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for strikes, got %d", rec.Code)
	}

	// 2. Request account deletion
	req, rec = usersReq(http.MethodPost, "/users/me/delete-request", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for delete request, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Cancel account deletion
	req, rec = usersReq(http.MethodDelete, "/users/me/delete-request", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for cancelling delete request, got %d", rec.Code)
	}
}
