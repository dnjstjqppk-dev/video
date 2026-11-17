package videos

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// PlaybackTokenTTL is the signed playback token lifetime (SECURITY_AND_COMPLIANCE §2).
const PlaybackTokenTTL = time.Hour

// signingSecret resolves the HMAC key for playback tokens. Production refuses
// to operate without an explicitly configured secret (fail-closed).
func signingSecret() ([]byte, error) {
	if secret := os.Getenv("PLAYBACK_SIGNING_SECRET"); secret != "" {
		return []byte(secret), nil
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return []byte(secret), nil
	}
	if strings.ToLower(os.Getenv("GVP_ENV")) == "production" {
		return nil, fmt.Errorf("playback signing secret is not configured")
	}
	return []byte("dev-only-playback-signing-secret"), nil
}

// SignPlaybackToken issues a "v1.<b64VideoID>.<expUnix>.<sigHex>" token bound
// to the video path prefix /hls/{videoID}/* with an expiry timestamp.
func SignPlaybackToken(videoID string, ttl time.Duration) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("invalid video id for playback token")
	}
	secret, err := signingSecret()
	if err != nil {
		return "", err
	}
	exp := time.Now().Add(ttl).Unix()
	unsigned := "v1." + base64.RawURLEncoding.EncodeToString([]byte(videoID)) + "." + strconv.FormatInt(exp, 10)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsigned))
	return unsigned + "." + hex.EncodeToString(mac.Sum(nil)), nil
}

// VerifyPlaybackToken checks signature, expiry, and video binding.
func VerifyPlaybackToken(token, videoID string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] == "" {
		return fmt.Errorf("malformed playback token")
	}
	rawID, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("malformed playback token")
	}
	if subtle.ConstantTimeCompare(rawID, []byte(videoID)) != 1 {
		return fmt.Errorf("playback token is not bound to this video")
	}
	exp, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return fmt.Errorf("playback token is expired")
	}
	secret, err := signingSecret()
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(strings.Join(parts[:3], ".")))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[3])) != 1 {
		return fmt.Errorf("invalid playback token signature")
	}
	return nil
}
