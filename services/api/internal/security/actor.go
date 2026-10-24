package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ActorHeader = "X-Internal-Actor"

type Actor struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email,omitempty"`
	Audience string `json:"aud"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Expires  int64  `json:"exp"`
}

func Sign(key []byte, a Actor) string {
	body, _ := json.Marshal(a)
	encoded := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func Verify(key []byte, raw string, r *http.Request) (Actor, error) {
	var a Actor
	parts := strings.Split(raw, ".")
	if len(parts) != 2 || len(raw) > 8192 {
		return a, errors.New("invalid actor assertion")
	}
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(parts[0]))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(want, got) {
		return a, errors.New("invalid signature")
	}
	b, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || json.Unmarshal(b, &a) != nil {
		return a, errors.New("invalid actor payload")
	}
	now := time.Now().Unix()
	if a.UserID == "" || a.Audience != "catalog" || a.Method != r.Method || a.Path != r.URL.EscapedPath() || a.Expires < now || a.Expires > now+90 {
		return a, fmt.Errorf("actor assertion mismatch")
	}
	return a, nil
}
