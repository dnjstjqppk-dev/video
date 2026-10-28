package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	TOTPPeriod = 30 // seconds
	TOTPDigits = 6
)

// GenerateTOTPSecret generates a secure Base32-encoded random secret for RFC 6238 TOTP (TASK-008)
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20) // 160 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// FormatTOTPURI formats the secret into standard otpauth URI for Google Authenticator / Authy
func FormatTOTPURI(secret, accountName, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		issuer, accountName, secret, issuer, TOTPDigits, TOTPPeriod)
}

// ComputeTOTP calculates the current 6-digit TOTP code for a secret and time
func ComputeTOTP(secret string, t time.Time) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid base32 secret: %w", err)
	}

	counter := uint64(t.Unix() / TOTPPeriod)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	// Dynamic truncation
	offset := h[len(h)-1] & 0x0f
	code := binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff
	code = code % uint32(math.Pow10(TOTPDigits))

	return fmt.Sprintf("%06d", code), nil
}

// ValidateTOTP validates the provided TOTP code allowing ±1 period clock drift (TASK-008)
func ValidateTOTP(secret, presentedCode string) bool {
	presentedCode = strings.TrimSpace(presentedCode)
	if len(presentedCode) != TOTPDigits {
		return false
	}

	now := time.Now().UTC()
	// Test current, previous (-30s), and next (+30s) windows
	for _, offset := range []int{0, -1, 1} {
		t := now.Add(time.Duration(offset*TOTPPeriod) * time.Second)
		expected, err := ComputeTOTP(secret, t)
		if err == nil && hmac.Equal([]byte(expected), []byte(presentedCode)) {
			return true
		}
	}
	return false
}

// GenerateBackupCodes produces N single-use 8-character recovery codes and their SHA-256 hashes (TASK-008)
func GenerateBackupCodes(count int) (plainCodes []string, hashedCodes []string, err error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // omit confusing 0, O, 1, I
	plainCodes = make([]string, count)
	hashedCodes = make([]string, count)

	for i := 0; i < count; i++ {
		bytes := make([]byte, 8)
		if _, err := rand.Read(bytes); err != nil {
			return nil, nil, err
		}
		var codeBuilder strings.Builder
		for _, b := range bytes {
			codeBuilder.WriteByte(charset[int(b)%len(charset)])
		}
		plain := codeBuilder.String()
		plainCodes[i] = plain[:4] + "-" + plain[4:]

		h := sha256.Sum256([]byte(plainCodes[i]))
		hashedCodes[i] = hex.EncodeToString(h[:])
	}
	return plainCodes, hashedCodes, nil
}

// ValidateAndBurnBackupCode verifies a recovery code against stored hashes, returning new hash list if valid
func ValidateAndBurnBackupCode(plainCode string, storedHashes []string) (bool, []string) {
	plainCode = strings.ToUpper(strings.TrimSpace(plainCode))
	h := sha256.Sum256([]byte(plainCode))
	hashHex := hex.EncodeToString(h[:])

	for i, stored := range storedHashes {
		if hmac.Equal([]byte(stored), []byte(hashHex)) {
			// Burn this code by removing from list
			remaining := append(storedHashes[:i], storedHashes[i+1:]...)
			return true, remaining
		}
	}
	return false, storedHashes
}
