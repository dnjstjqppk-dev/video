package partners

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func TestPartnerEndpoints(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	router := auth.Authenticate(h.Routes(auth))

	for _, path := range []string{"/me", "/channels", "/contracts"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer advertiser-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK for %s, got %d", path, rec.Code)
		}
	}
}

func TestRevokeContract(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	newRouter := func() (http.Handler, *store.MemoryStore) {
		s := store.NewMemoryStore()
		h := NewHandler(s)
		auth := identity.NewAuthMiddleware(s)
		return auth.Authenticate(h.Routes(auth)), s
	}
	call := func(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer advertiser-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	t.Run("revokes active contract", func(t *testing.T) {
		router, s := newRouter()
		rec := call(router, http.MethodPatch, "/contracts/ct-2026-001", `{"reason":"creator request"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		if s.Contracts["ct-2026-001"].Status != "revoked" {
			t.Fatalf("expected status revoked, got %s", s.Contracts["ct-2026-001"].Status)
		}
	})

	t.Run("rejects empty reason", func(t *testing.T) {
		router, _ := newRouter()
		rec := call(router, http.MethodPatch, "/contracts/ct-2026-001", `{"reason":"  "}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("rejects unknown contract", func(t *testing.T) {
		router, _ := newRouter()
		rec := call(router, http.MethodPatch, "/contracts/nope", `{"reason":"x"}`)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})

	t.Run("rejects double revoke", func(t *testing.T) {
		router, _ := newRouter()
		if rec := call(router, http.MethodPatch, "/contracts/ct-2026-001", `{"reason":"first"}`); rec.Code != http.StatusOK {
			t.Fatalf("expected 200 on first revoke, got %d", rec.Code)
		}
		if rec := call(router, http.MethodPatch, "/contracts/ct-2026-001", `{"reason":"second"}`); rec.Code != http.StatusConflict {
			t.Fatalf("expected 409 on second revoke, got %d", rec.Code)
		}
	})
}
