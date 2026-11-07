package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

func newFinanceRouter() (http.Handler, *store.MemoryStore) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	r := h.Routes(auth)
	mux := http.NewServeMux()
	mux.Handle("/finance/summary", auth.Authenticate(
		auth.RequireRole("super_admin", "platform_admin", "finance_admin")(http.HandlerFunc(h.GetFinanceSummary)),
	))
	mux.Handle("/finance/batches/approve", auth.Authenticate(
		auth.RequireRole("super_admin", "platform_admin")(http.HandlerFunc(h.ApproveBatch)),
	))
	_ = r
	return mux, s
}

func TestFinanceSummaryBalanced(t *testing.T) {
	router, _ := newFinanceRouter()
	req := httptest.NewRequest(http.MethodGet, "/finance/summary", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"balanced":true`) {
		t.Fatalf("expected balanced ledger, got %s", rec.Body.String())
	}
}

func TestApproveBatch(t *testing.T) {
	setup := func() (http.Handler, *store.MemoryStore, string) {
		router, s := newFinanceRouter()
		user := s.Users["11111111-1111-1111-1111-111111111111"]
		secret := "JBSWY3DPEHPK3PXP"
		user.MFAEnabled = true
		user.TOTPSecret = secret
		code, err := security.ComputeTOTP(secret, time.Now())
		if err != nil {
			t.Fatalf("compute totp: %v", err)
		}
		return router, s, code
	}
	call := func(router http.Handler, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/finance/batches/approve", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer admin-token")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	t.Run("approves with valid step-up code", func(t *testing.T) {
		router, s, code := setup()
		before := len(s.AuditLogs)
		rec := call(router, `{"batchId":"BATCH-2026-10","totpCode":"`+code+`"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if len(s.AuditLogs) != before+1 {
			t.Fatalf("expected audit log append, got %d -> %d", before, len(s.AuditLogs))
		}
	})

	t.Run("rejects wrong code", func(t *testing.T) {
		router, _, _ := setup()
		rec := call(router, `{"batchId":"BATCH-2026-10","totpCode":"000000"}`)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("requires enrolled TOTP", func(t *testing.T) {
		router, s := newFinanceRouter()
		s.Users["11111111-1111-1111-1111-111111111111"].MFAEnabled = false
		rec := call(router, `{"batchId":"BATCH-2026-10","totpCode":"123456"}`)
		if rec.Code != http.StatusPreconditionFailed {
			t.Fatalf("expected 412, got %d", rec.Code)
		}
	})
}
