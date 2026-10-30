package admin_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/admin"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func newAdminTestServer() *httptest.Server {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	handler := admin.NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/admin", handler.Routes(auth))

	return httptest.NewServer(r)
}

func adminReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
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

func TestAdminDashboard_RequiresSuperAdmin(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	// creator (2222...) accessing super_admin-only admin dashboard -> 403
	req, rec := adminReq(http.MethodGet, "/admin/dashboard", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 FORBIDDEN for non-super_admin, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminDashboard_SuperAdminCanAccess(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	req, rec := adminReq(http.MethodGet, "/admin/dashboard", "admin-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListUsers_RequiresSuperAdmin(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	req, rec := adminReq(http.MethodGet, "/admin/users", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 FORBIDDEN, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListUsers_SuperAdminCanAccess(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	req, rec := adminReq(http.MethodGet, "/admin/users", "admin-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateUserStatus_SuperAdminCanUpdate(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	userID := "22222222-2222-2222-2222-222222222222"
	req, rec := adminReq(http.MethodPatch, "/admin/users/"+userID+"/status", "admin-token", map[string]string{"status": "suspended", "reason": "Policy violation"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateUserStatus_RequiresSuperAdmin(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	userID := "22222222-2222-2222-2222-222222222222"
	req, rec := adminReq(http.MethodPatch, "/admin/users/"+userID+"/status", "creator-token", map[string]string{"status": "suspended"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 FORBIDDEN for non-super_admin, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetAuditLogs_SuperAdminCanAccess(t *testing.T) {
	srv := newAdminTestServer()
	defer srv.Close()

	req, rec := adminReq(http.MethodGet, "/admin/audit-logs", "admin-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
