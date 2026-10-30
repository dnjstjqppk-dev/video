package identity

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

func TestDevTokensRequireExplicitOptIn(t *testing.T) {
	for _, tc := range []struct {
		env, flag string
		want      bool
	}{
		{"", "", false}, {"", "true", false}, {"production", "true", false},
		{"staging", "true", false}, {"development", "", false},
		{"development", "true", true}, {"test", "true", true}, {"test", "false", false},
	} {
		t.Run(tc.env+"/"+tc.flag, func(t *testing.T) {
			t.Setenv("GVP_ENV", tc.env)
			t.Setenv("GVP_ALLOW_DEV_TOKENS", tc.flag)
			if got := NewAuthMiddleware(store.NewMemoryStore()).AllowDevTokens; got != tc.want {
				t.Fatalf("AllowDevTokens=%v, want %v", got, tc.want)
			}
		})
	}
}

func TestInactiveUserRejected(t *testing.T) {
	s := store.NewMemoryStore()
	s.Users["22222222-2222-2222-2222-222222222222"].IsActive = false
	auth := &AuthMiddleware{Store: s, AllowDevTokens: true}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer creator-token")
	rec := httptest.NewRecorder()
	auth.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("inactive user reached handler")
	})).ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}
