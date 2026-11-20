package admin_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/admin"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func newAdminRouter() http.Handler {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/admin", admin.NewHandler(memStore).Routes(auth))
	return r
}

func adminGet(t *testing.T, router http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestListAIReviews_StaffOnly(t *testing.T) {
	router := newAdminRouter()
	if rec := adminGet(t, router, "/admin/ai-reviews", ""); rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous: expected 401, got %d", rec.Code)
	}
	if rec := adminGet(t, router, "/admin/ai-reviews", "viewer-token"); rec.Code != http.StatusForbidden {
		t.Fatalf("viewer: expected 403, got %d", rec.Code)
	}
	if rec := adminGet(t, router, "/admin/ai-reviews", "creator-token"); rec.Code != http.StatusForbidden {
		t.Fatalf("creator: expected 403, got %d", rec.Code)
	}
}

func TestListAIReviews_ListsPending(t *testing.T) {
	router := newAdminRouter()
	rec := adminGet(t, router, "/admin/ai-reviews", "admin-token")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["total"].(float64) != 1 {
		t.Fatalf("total = %v, want 1 pending seed", body["total"])
	}
	items := body["items"].([]interface{})
	item := items[0].(map[string]interface{})
	video := item["video"].(map[string]interface{})
	if video["id"] != "00000000-0000-0000-0000-0000000000a4" {
		t.Fatalf("wrong queued video: %v", video["id"])
	}
	if item["channel_name"] != "NeoDrama AI" {
		t.Fatalf("missing channel enrichment: %v", item)
	}
}
