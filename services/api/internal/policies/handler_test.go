package policies

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func newPoliciesTestServer() (*httptest.Server, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	h := NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/policies", h.Routes(auth))

	return httptest.NewServer(r), memStore
}

func policiesReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, httptest.NewRecorder()
}

func TestPoliciesLifecycle(t *testing.T) {
	srv, memStore := newPoliciesTestServer()
	defer srv.Close()

	// 1. Admin reads config + flags
	req, rec := policiesReq(http.MethodGet, "/policies", "admin-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for policies, got %d", rec.Code)
	}
	var listing struct {
		Config struct {
			Version         string `json:"version"`
			MaxUploadSizeGB int64  `json:"max_upload_size_gb"`
		} `json:"config"`
		Flags []struct {
			Key     string `json:"key"`
			Enabled bool   `json:"enabled"`
		} `json:"flags"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listing); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(listing.Flags) != 4 || listing.Config.MaxUploadSizeGB != 128 {
		t.Fatalf("unexpected policies payload: %+v", listing)
	}

	// 2. Creator is forbidden from reading policies
	req, rec = policiesReq(http.MethodGet, "/policies", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for creator policies read, got %d", rec.Code)
	}

	// 3. Admin updates config with validation
	req, rec = policiesReq(http.MethodPatch, "/policies", "admin-token", map[string]interface{}{
		"max_upload_size_gb": 64,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for policy update, got %d: %s", rec.Code, rec.Body.String())
	}

	req, rec = policiesReq(http.MethodPatch, "/policies", "admin-token", map[string]interface{}{
		"creator_revenue_share_pct": 150,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid revenue share, got %d", rec.Code)
	}

	// 4. Admin toggles a kill-switch off and on
	req, rec = policiesReq(http.MethodPatch, "/policies/flags/"+FlagUploadIngest, "admin-token", map[string]interface{}{
		"enabled": false,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for flag toggle, got %d", rec.Code)
	}
	if FlagEnabled(memStore, FlagUploadIngest) {
		t.Fatalf("kill-switch must report disabled after toggle")
	}

	req, rec = policiesReq(http.MethodPatch, "/policies/flags/"+FlagUploadIngest, "admin-token", map[string]interface{}{
		"enabled": true,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for flag re-enable, got %d", rec.Code)
	}

	// 5. Unknown flag key 404s
	req, rec = policiesReq(http.MethodPatch, "/policies/flags/NOPE", "admin-token", map[string]interface{}{
		"enabled": true,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown flag, got %d", rec.Code)
	}
}

func TestFlagEnabledDefaults(t *testing.T) {
	s := store.NewMemoryStore()
	if !FlagEnabled(s, FlagUploadIngest) {
		t.Fatalf("seeded flags must default to enabled")
	}
	if !FlagEnabled(s, "UNKNOWN_KEY") {
		t.Fatalf("unknown keys must default to enabled")
	}
	s.FeatureFlags[FlagLiveIngest].Enabled = false
	if FlagEnabled(s, FlagLiveIngest) {
		t.Fatalf("disabled flag must report disabled")
	}
}
