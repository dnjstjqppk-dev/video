package moderation

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

const stepUpSecret = "JBSWY3DPEHPK3PXP"

func resolveCall(t *testing.T, s *store.MemoryStore, h *Handler, auth *identity.AuthMiddleware, token string, payload map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/mod-case-1/resolve", bytes.NewReader(b))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/", h.Routes(auth))
	r.ServeHTTP(rec, req)
	return rec
}

func enrollAdmin(t *testing.T, s *store.MemoryStore) string {
	t.Helper()
	admin := s.Users["11111111-1111-1111-1111-111111111111"]
	admin.MFAEnabled = true
	admin.TOTPSecret = stepUpSecret
	code, err := security.ComputeTOTP(stepUpSecret, time.Now())
	if err != nil {
		t.Fatalf("compute totp: %v", err)
	}
	return code
}

// P0-7: resolving a moderation case requires MFA step-up and writes an audit record.
func TestResolveCaseRequiresStepUp(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}

	rec := resolveCall(t, s, h, auth, "admin-token", map[string]string{"action": "remove"})
	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 without MFA enrollment, got %d", rec.Code)
	}

	code := enrollAdmin(t, s)
	rec = resolveCall(t, s, h, auth, "admin-token", map[string]string{"action": "remove", "totpCode": "000000"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on wrong code, got %d", rec.Code)
	}

	before := len(s.AuditLogs)
	rec = resolveCall(t, s, h, auth, "admin-token", map[string]string{"action": "remove", "notes": "confirmed", "totpCode": code})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid step-up, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(s.AuditLogs) != before+1 {
		t.Fatalf("resolve must append an audit log, got %d -> %d", before, len(s.AuditLogs))
	}
	if s.Moderation["mod-case-1"].Status != "resolved" {
		t.Fatalf("expected resolved status, got %s", s.Moderation["mod-case-1"].Status)
	}
}

// P0-7: unknown resolve actions are rejected.
func TestResolveCaseValidatesAction(t *testing.T) {
	s := store.NewMemoryStore()
	h := NewHandler(s)
	auth := &identity.AuthMiddleware{Store: s, AllowDevTokens: true}
	code := enrollAdmin(t, s)

	rec := resolveCall(t, s, h, auth, "admin-token", map[string]string{"action": "nuke", "totpCode": code})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unknown action, got %d", rec.Code)
	}
}
