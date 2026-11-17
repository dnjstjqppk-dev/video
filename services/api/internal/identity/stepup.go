package identity

import (
	"net/http"

	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

// StepUpHeader carries the TOTP step-up code for requests without a JSON body.
const StepUpHeader = "X-TOTP-Code"

// StepUpCode extracts the TOTP step-up code: explicit body value first,
// then the X-TOTP-Code header, then the ?totpCode= query parameter.
func StepUpCode(r *http.Request, bodyCode string) string {
	if bodyCode != "" {
		return bodyCode
	}
	if header := r.Header.Get(StepUpHeader); header != "" {
		return header
	}
	return r.URL.Query().Get("totpCode")
}

// RequireStepUp enforces MFA enrollment plus a fresh TOTP code for sensitive
// admin mutations. It writes 412 MFA_REQUIRED when the user has not enrolled,
// 401 INVALID_TOTP on a wrong or expired code, and returns false in both
// cases. Callers proceed only when it returns true.
func RequireStepUp(w http.ResponseWriter, r *http.Request, user *store.User, bodyCode string) bool {
	if user == nil {
		WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return false
	}
	if !user.MFAEnabled || user.TOTPSecret == "" {
		WriteError(w, r, http.StatusPreconditionFailed, "MFA_REQUIRED", "Enroll TOTP via /v1/me/mfa/setup before this sensitive action")
		return false
	}
	if !security.ValidateTOTP(user.TOTPSecret, StepUpCode(r, bodyCode)) {
		WriteError(w, r, http.StatusUnauthorized, "INVALID_TOTP", "Step-up code is invalid or expired")
		return false
	}
	return true
}
