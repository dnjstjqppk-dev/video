package recommendations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func seedSearchStore() *store.MemoryStore {
	s := store.NewMemoryStore()
	s.Channels["ch-search-test"] = &store.Channel{
		ID: "ch-search-test", OwnerID: "22222222-2222-2222-2222-222222222222",
		Handle: "@search-test", Name: "Search Test",
	}
	mk := func(id, visibility, title string) {
		s.Videos[id] = &store.Video{
			ID: id, ChannelID: "ch-search-test", Title: title,
			Description: "searchable fixture", Visibility: visibility, Status: "READY",
		}
	}
	mk("vid-s-pub", "public", "Zebra migration documentary")
	mk("vid-s-unlisted", "unlisted", "Zebra unlisted cut")
	mk("vid-s-private", "private", "Zebra private draft")
	mk("vid-s-members", "members_only", "Zebra members premiere")
	return s
}

func searchRouter(s *store.MemoryStore) http.Handler {
	auth := identity.NewAuthMiddleware(s)
	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Get("/search", NewHandler(s).Search)
	r.Get("/search/suggestions", NewHandler(s).SearchSuggestions)
	return r
}

func searchIDs(t *testing.T, router http.Handler, path, token string) []string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for %s, got %d", path, rec.Code)
	}
	var body struct {
		Videos []struct {
			ID string `json:"id"`
		} `json:"videos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode search: %v", err)
	}
	ids := make([]string, 0, len(body.Videos))
	for _, v := range body.Videos {
		ids = append(ids, v.ID)
	}
	return ids
}

func containsID(ids []string, id string) bool {
	for _, v := range ids {
		if v == id {
			return true
		}
	}
	return false
}

// P0-4: public search never surfaces private or unlisted videos.
func TestSearchExcludesNonPublic(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := searchRouter(seedSearchStore())

	for _, token := range []string{"", "viewer-token", "creator-token", "admin-token"} {
		ids := searchIDs(t, router, "/search?q=Zebra", token)
		if !containsID(ids, "vid-s-pub") {
			t.Fatalf("token %q: public match must be returned, got %v", token, ids)
		}
		for _, hidden := range []string{"vid-s-unlisted", "vid-s-private"} {
			if containsID(ids, hidden) {
				t.Fatalf("token %q: search leaked %s: %v", token, hidden, ids)
			}
		}
	}
}

// P0-4: members-only videos appear only for entitled viewers.
func TestSearchMembersOnlyEntitlement(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := searchRouter(seedSearchStore())

	ids := searchIDs(t, router, "/search?q=Zebra", "")
	if containsID(ids, "vid-s-members") {
		t.Fatalf("anonymous search must not surface members-only: %v", ids)
	}
	ids = searchIDs(t, router, "/search?q=Zebra", "viewer-token")
	if containsID(ids, "vid-s-members") {
		t.Fatalf("free-tier search must not surface members-only: %v", ids)
	}
	ids = searchIDs(t, router, "/search?q=Zebra", "creator-token")
	if !containsID(ids, "vid-s-members") {
		t.Fatalf("entitled search must surface members-only, got %v", ids)
	}
}

// P0-4: autocomplete never suggests non-public titles.
func TestSearchSuggestionsExcludeNonPublic(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	router := searchRouter(seedSearchStore())

	req := httptest.NewRequest(http.MethodGet, "/search/suggestions?q=Zebra", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for suggestions, got %d", rec.Code)
	}
	var suggestions []string
	if err := json.Unmarshal(rec.Body.Bytes(), &suggestions); err != nil {
		t.Fatalf("failed to decode suggestions: %v", err)
	}
	for _, title := range suggestions {
		switch title {
		case "Zebra unlisted cut", "Zebra private draft", "Zebra members premiere":
			t.Fatalf("suggestions leaked non-public title %q", title)
		}
	}
}
