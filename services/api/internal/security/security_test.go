package security

import (
	"strings"
	"testing"
	"time"
)

func TestArgon2HashAndVerify(t *testing.T) {
	password := "SuperSecretPassword123!"
	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id hash prefix, got %s", hash)
	}

	ok, err := VerifyPassword(password, hash)
	if err != nil || !ok {
		t.Fatalf("VerifyPassword should succeed: ok=%v err=%v", ok, err)
	}

	ok, err = VerifyPassword("WrongPassword", hash)
	if err != nil {
		t.Fatalf("VerifyPassword error: %v", err)
	}
	if ok {
		t.Fatal("VerifyPassword should fail for wrong password")
	}
}

func TestArgon2InvalidHash(t *testing.T) {
	_, err := VerifyPassword("password", "invalid-hash")
	if err == nil {
		t.Fatal("expected error for invalid hash")
	}
}

func TestTOTPGenerateAndValidate(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("GenerateTOTPSecret failed: %v", err)
	}
	if secret == "" {
		t.Fatal("secret should not be empty")
	}

	code, err := ComputeTOTP(secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("ComputeTOTP failed: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-digit code, got %s", code)
	}

	if !ValidateTOTP(secret, code) {
		t.Fatal("ValidateTOTP should accept current code")
	}
}

func TestTOTPBackupCodes(t *testing.T) {
	plain, hashed, err := GenerateBackupCodes(5)
	if err != nil {
		t.Fatalf("GenerateBackupCodes failed: %v", err)
	}
	if len(plain) != 5 || len(hashed) != 5 {
		t.Fatalf("expected 5 codes, got %d/%d", len(plain), len(hashed))
	}

	ok, remaining := ValidateAndBurnBackupCode(plain[0], hashed)
	if !ok {
		t.Fatal("ValidateAndBurnBackupCode should accept valid code")
	}
	if len(remaining) != 4 {
		t.Fatalf("expected 4 remaining codes, got %d", len(remaining))
	}

	ok2, _ := ValidateAndBurnBackupCode(plain[0], remaining)
	if ok2 {
		t.Fatal("burned code should not be accepted")
	}
}

func TestMaskEmail(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"alexander@bookingnetwork.com", "a***r@bookingnetwork.com"},
		{"ab@test.com", "a***@test.com"},
	}
	for _, c := range cases {
		got := MaskEmail(c.input)
		if got != c.expected {
			t.Errorf("MaskEmail(%s) = %s, want %s", c.input, got, c.expected)
		}
	}
	if MaskEmail("invalid") != "***" {
		t.Errorf("MaskEmail(invalid) should return ***")
	}
}

func TestMaskPhone(t *testing.T) {
	got := MaskPhone("+82-10-1234-5678")
	if got != "+82-10-****-5678" {
		t.Errorf("MaskPhone = %s", got)
	}
	if MaskPhone("short") != "****" {
		t.Errorf("MaskPhone(short) should return ****")
	}
}

func TestMaskTaxID(t *testing.T) {
	got := MaskTaxID("1234567890")
	if got != "123-**-*****" {
		t.Errorf("MaskTaxID = %s", got)
	}
	if MaskTaxID("123") != "***-**-*****" {
		t.Errorf("MaskTaxID(short) wrong")
	}
}

func TestApplyCOPPARules(t *testing.T) {
	kids := ApplyCOPPARules(true)
	if kids.CommentsAllowed || kids.PersonalizedAdsAllowed || kids.LiveChatAllowed {
		t.Fatal("COPPA should disable comments, personalized ads, live chat for kids")
	}
	normal := ApplyCOPPARules(false)
	if !normal.CommentsAllowed || !normal.PersonalizedAdsAllowed {
		t.Fatal("normal content should allow comments and personalized ads")
	}
}

func TestNewAccountDeletionRequest(t *testing.T) {
	req := NewAccountDeletionRequest("user-123")
	if req.UserID != "user-123" {
		t.Fatalf("UserID = %s", req.UserID)
	}
	if req.IsCancelled {
		t.Fatal("should not be cancelled")
	}
	if req.PermanentDeleteAt.Before(req.RequestedAt) {
		t.Fatal("PermanentDeleteAt should be after RequestedAt")
	}
}

func TestSessionCreateAndRotate(t *testing.T) {
	sm := NewSessionManager(15*time.Minute, 30*24*time.Hour)
	sess, token, err := sm.CreateSession("user-1", "fp-1", "1.2.3.4", "KR", 37.5, 127.0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if sess == nil || token == "" {
		t.Fatal("session and token should not be nil/empty")
	}
	if sess.UserID != "user-1" {
		t.Fatalf("UserID = %s", sess.UserID)
	}

	rotated, newToken, err := sm.RotateRefreshToken(sess.ID, token, "fp-1", "1.2.3.4", "KR", 37.5, 127.0)
	if err != nil {
		t.Fatalf("RotateRefreshToken failed: %v", err)
	}
	if rotated == nil || newToken == "" || newToken == token {
		t.Fatal("rotation should issue new token")
	}

	_, _, err = sm.RotateRefreshToken(sess.ID, token, "fp-1", "1.2.3.4", "KR", 37.5, 127.0)
	if err != ErrTokenReused {
		t.Fatalf("expected ErrTokenReused, got %v", err)
	}
}

func TestSessionRevoke(t *testing.T) {
	sm := NewSessionManager(15*time.Minute, 30*24*time.Hour)
	sess, _, err := sm.CreateSession("user-2", "fp-2", "5.6.7.8", "US", 40.0, -74.0)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	err = sm.RevokeSession(sess.ID)
	if err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}
	err = sm.RevokeSession(sess.ID)
	if err != ErrSessionNotFound {
		t.Fatalf("expected ErrSessionNotFound after revoke, got %v", err)
	}
}

func TestRateLimiterAllow(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		allowed, _, _ := rl.Allow("ip-1")
		if !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	allowed, remaining, retryAfter := rl.Allow("ip-1")
	if allowed {
		t.Fatal("4th request should be denied")
	}
	if remaining != 0 {
		t.Errorf("remaining = %d", remaining)
	}
	if retryAfter <= 0 {
		t.Error("retryAfter should be positive")
	}
}

func TestAdminGuardSudo(t *testing.T) {
	ag := NewAdminGuard()
	ag.ElevateSudo("admin-1")
	if !ag.IsSudoActive("admin-1") {
		t.Fatal("sudo should be active")
	}
	if ag.IsSudoActive("admin-2") {
		t.Fatal("sudo should not be active for unknown user")
	}
}

func TestVerifyWebhookHMAC(t *testing.T) {
	payload := []byte("test payload")
	if VerifyWebhookHMAC(payload, "", "sig") {
		t.Fatal("should fail with empty secret")
	}
	if VerifyWebhookHMAC(payload, "secret", "") {
		t.Fatal("should fail with empty signature")
	}
}
