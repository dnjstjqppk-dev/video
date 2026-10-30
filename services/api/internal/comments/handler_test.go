package comments_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/comments"
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

func TestDeleteComment_RequiresAuth(t *testing.T) {
	srv := newCommentsTestServer()
	defer srv.Close()

	req, rec := commentsReq(http.MethodDelete, "/comments/comm-1", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthenticated, got %d", rec.Code)
	}
}

func TestDeleteComment_AuthenticatedCanDeleteAny(t *testing.T) {
	// NOTE (BOLA/IDOR): DeleteComment iterates all comments by ID only and performs
	// no ownership check (comment.UserID vs actor). Any authenticated user can delete
	// any comment. This documents current behavior; add ABAC ownership check to
	// handler and change expectation to 403 for non-owners.
	srv := newCommentsTestServer()
	defer srv.Close()

	// comm-1 is owned by advertiser (4444...), deleted by creator (2222...)
	req, rec := commentsReq(http.MethodDelete, "/comments/comm-1", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 (current no-ownership behavior), got %d: %s", rec.Code, rec.Body.String())
	}
}
