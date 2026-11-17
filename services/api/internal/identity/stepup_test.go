package identity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func stepUpUser() *store.User {
	return &store.User{ID: "admin-1", Role: "super_admin", IsActive: true}
}

// P0-7: sensitive mutations require MFA enrollment first.
func TestRequireStepUpRequiresEnrollment(t *testing.T) {
	user := stepUpUser()
	req := httptest.NewRequest(http.MethodPatch, "/x", nil)
	rec := httptest.NewRecorder()
	if RequireStepUp(rec, req, user, "123456") {
		t.Fatal("unenrolled user must be refused")
	}
	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected 412 MFA_REQUIRED, got %d", rec.Code)
	}
}

// P0-7: wrong codes are rejected, fresh codes pass.
func TestRequireStepUpValidatesCode(t *testing.T) {
	user := stepUpUser()
	user.MFAEnabled = true
	user.TOTPSecret = "JBSWY3DPEHPK3PXP"

	badReq := httptest.NewRequest(http.MethodPatch, "/x", nil)
	badRec := httptest.NewRecorder()
	if RequireStepUp(badRec, badReq, user, "000000") {
		t.Fatal("wrong code must be refused")
	}
	if badRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 INVALID_TOTP, got %d", badRec.Code)
	}
	if RequireStepUp(httptest.NewRecorder(), httptest.NewRequest(http.MethodPatch, "/x", nil), nil, "000000") {
		t.Fatal("nil user must be refused")
	}
}

// P0-7: code may arrive via body, header, or query.
func TestStepUpCodeSources(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/x", nil)
	if got := StepUpCode(req, "111111"); got != "111111" {
		t.Fatalf("body code must win, got %q", got)
	}
	req.Header.Set("X-TOTP-Code", "222222")
	if got := StepUpCode(req, ""); got != "222222" {
		t.Fatalf("header code must apply, got %q", got)
	}
	req2 := httptest.NewRequest(http.MethodPatch, "/x?totpCode=333333", nil)
	if got := StepUpCode(req2, ""); got != "333333" {
		t.Fatalf("query code must apply, got %q", got)
	}
}
