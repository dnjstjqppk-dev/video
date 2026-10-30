package comments_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/comments"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func newCommentsTestServer() *httptest.Server {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	handler := comments.NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Get("/videos/{id}/comments", handler.GetComments)
	r.With(auth.RequireAuth).Post("/videos/{id}/comments", handler.AddComment)
	r.With(auth.RequireAuth).Delete("/comments/{id}", handler.DeleteComment)
	r.With(auth.RequireAuth).Post("/comments/{id}/like", handler.LikeComment)
	r.With(auth.RequireAuth).Post("/comments/{id}/heart", handler.HeartComment)
	r.With(auth.RequireAuth).Post("/comments/{id}/pin", handler.PinComment)

	return httptest.NewServer(r)
}

func commentsReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var rbody *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rbody = bytes.NewReader(b)
	} else {
		rbody = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, rbody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, httptest.NewRecorder()
}

func TestGetComments_ReturnsList(t *testing.T) {
	srv := newCommentsTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := commentsReq(http.MethodGet, "/videos/"+videoID+"/comments", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddComment_RequiresAuth(t *testing.T) {
	srv := newCommentsTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := commentsReq(http.MethodPost, "/videos/"+videoID+"/comments", "", map[string]string{"content": "spam"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestAddComment_Authenticated(t *testing.T) {
	srv := newCommentsTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"
	req, rec := commentsReq(http.MethodPost, "/videos/"+videoID+"/comments", "creator-token", map[string]string{"content": "Great video!"})
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNestedRepliesAndInteractions(t *testing.T) {
	srv := newCommentsTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"

	// 1. Add root comment (viewer)
	req, rec := commentsReq(http.MethodPost, "/videos/"+videoID+"/comments", "viewer-token", map[string]string{
		"content": "This is an insightful architectural walkthrough!",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for root comment, got %d: %s", rec.Code, rec.Body.String())
	}
	var rootComment store.Comment
	json.Unmarshal(rec.Body.Bytes(), &rootComment)

	// 2. Add reply comment with parent_id (creator)
	req, rec = commentsReq(http.MethodPost, "/videos/"+videoID+"/comments", "creator-token", map[string]string{
		"content":   "Thank you! Glad you found the HLS section useful.",
		"parent_id": rootComment.ID,
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for reply comment, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. Like comment
	req, rec = commentsReq(http.MethodPost, "/comments/"+rootComment.ID+"/like", "viewer-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for like comment, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Creator hearts comment
	req, rec = commentsReq(http.MethodPost, "/comments/"+rootComment.ID+"/heart", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for creator heart, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. Creator pins comment
	req, rec = commentsReq(http.MethodPost, "/comments/"+rootComment.ID+"/pin", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for creator pin, got %d: %s", rec.Code, rec.Body.String())
	}

	// 6. Verify GetComments returns root comment with nested reply and pinned flag
	req, rec = commentsReq(http.MethodGet, "/videos/"+videoID+"/comments?sort=top", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for get comments, got %d", rec.Code)
	}
	var listResp struct {
		Items []*store.Comment `json:"items"`
		Total int              `json:"total"`
	}
	json.Unmarshal(rec.Body.Bytes(), &listResp)

	var foundRoot *store.Comment
	for _, c := range listResp.Items {
		if c.ID == rootComment.ID {
			foundRoot = c
			break
		}
	}
	if foundRoot == nil {
		t.Fatal("root comment not found in comments list")
	}
	if !foundRoot.IsPinned || !foundRoot.IsHearted || foundRoot.LikeCount < 1 {
		t.Fatalf("root comment properties mismatch: pinned=%v hearted=%v likes=%d", foundRoot.IsPinned, foundRoot.IsHearted, foundRoot.LikeCount)
	}
	if len(foundRoot.Replies) != 1 {
		t.Fatalf("expected 1 nested reply in root comment, got %d", len(foundRoot.Replies))
	}
}

func TestDeleteComment_Ownership(t *testing.T) {
	srv := newCommentsTestServer()
	defer srv.Close()

	videoID := "00000000-0000-0000-0000-000000000001"

	// 1. Post comment as viewer
	req, rec := commentsReq(http.MethodPost, "/videos/"+videoID+"/comments", "viewer-token", map[string]string{
		"content": "To be deleted soon",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	var comm store.Comment
	json.Unmarshal(rec.Body.Bytes(), &comm)

	// 2. Another user (advertiser) tries to delete -> forbidden 403
	req, rec = commentsReq(http.MethodDelete, "/comments/"+comm.ID, "advertiser-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden for unauthorized delete, got %d", rec.Code)
	}

	// 3. Author (viewer) deletes -> ok 200
	req, rec = commentsReq(http.MethodDelete, "/comments/"+comm.ID, "viewer-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for author delete, got %d: %s", rec.Code, rec.Body.String())
	}
}
