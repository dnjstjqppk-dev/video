package recommendations_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/recommendations"
	"globalvideoplatform/services/api/internal/store"
)

func newRecomRouter() (http.Handler, *store.MemoryStore) {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/feed", recommendations.NewHandler(memStore).Routes(auth))
	h := recommendations.NewHandler(memStore)
	r.Get("/ai-originals", h.GetOriginals)
	return r, memStore
}

func getJSON(t *testing.T, router http.Handler, path string) map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: expected 200, got %d: %s", path, rec.Code, rec.Body.String())
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func itemIDs(t *testing.T, envelope map[string]interface{}, key string) []string {
	t.Helper()
	items, ok := envelope[key].([]interface{})
	if !ok {
		t.Fatalf("missing %q items in %v", key, envelope)
	}
	var ids []string
	for _, it := range items {
		m := it.(map[string]interface{})
		id, _ := m["id"].(string)
		if id == "" {
			id, _ = m["video_id"].(string)
		}
		ids = append(ids, id)
	}
	return ids
}

func contains(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func TestGetOriginals_ApprovedOnly(t *testing.T) {
	router, _ := newRecomRouter()
	body := getJSON(t, router, "/ai-originals")
	videos := body["videos"].(map[string]interface{})
	series := body["series"].(map[string]interface{})

	if videos["total"].(float64) != 3 {
		t.Fatalf("videos total = %v, want 3 approved episodes", videos["total"])
	}
	ids := itemIDs(t, videos, "items")
	for _, want := range []string{
		"00000000-0000-0000-0000-0000000000a1",
		"00000000-0000-0000-0000-0000000000a2",
		"00000000-0000-0000-0000-0000000000a3",
	} {
		if !contains(ids, want) {
			t.Fatalf("missing approved video %s in %v", want, ids)
		}
	}
	if contains(ids, "00000000-0000-0000-0000-0000000000a4") {
		t.Fatal("pending AI video must not appear on the Originals shelf")
	}
	if contains(ids, "00000000-0000-0000-0000-000000000001") {
		t.Fatal("human-made video must not appear on the Originals shelf")
	}
	if series["total"].(float64) != 1 {
		t.Fatalf("series total = %v, want 1", series["total"])
	}
}

func TestGetOriginals_TypeFilter(t *testing.T) {
	router, _ := newRecomRouter()
	body := getJSON(t, router, "/ai-originals?type=series")
	if body["videos"].(map[string]interface{})["total"].(float64) != 0 {
		t.Fatal("type=series must exclude videos")
	}
	if body["series"].(map[string]interface{})["total"].(float64) != 1 {
		t.Fatal("type=series must keep the series rail")
	}
	body = getJSON(t, router, "/ai-originals?type=videos&video_type=shorts")
	if body["videos"].(map[string]interface{})["total"].(float64) != 0 {
		t.Fatal("no approved AI shorts exist in seeds; want empty")
	}
}

func TestHomeFeed_ExcludesAIVideos(t *testing.T) {
	router, _ := newRecomRouter()
	body := getJSON(t, router, "/feed/home")
	ids := itemIDs(t, body, "items")
	if !contains(ids, "00000000-0000-0000-0000-000000000001") {
		t.Fatalf("human video missing from home feed: %v", ids)
	}
	for _, ai := range []string{
		"00000000-0000-0000-0000-0000000000a1",
		"00000000-0000-0000-0000-0000000000a2",
		"00000000-0000-0000-0000-0000000000a3",
	} {
		if contains(ids, ai) {
			t.Fatalf("AI video %s must not appear in the home feed", ai)
		}
	}
}

func TestShortsFeed_ExcludesAIShorts(t *testing.T) {
	router, _ := newRecomRouter()
	body := getJSON(t, router, "/feed/shorts")
	ids := itemIDs(t, body, "items")
	if contains(ids, "00000000-0000-0000-0000-0000000000a4") {
		t.Fatal("pending AI short must not appear in the shorts feed")
	}
}

// Shorts shelf serves every public non-AI short.
func TestShortsFeedServesPublicShorts(t *testing.T) {
	router, _ := newRecomRouter()
	body := getJSON(t, router, "/feed/shorts")
	ids := itemIDs(t, body, "items")
	for _, want := range []string{
		"00000000-0000-0000-0000-000000000004",
		"00000000-0000-0000-0000-000000000006",
		"00000000-0000-0000-0000-000000000007",
	} {
		if !contains(ids, want) {
			t.Fatalf("shorts feed missing %s (got %v)", want, ids)
		}
	}
}
