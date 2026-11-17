package comments_test

import (
	"net/http"
	"strings"
	"testing"
)

// P1-16: blocklisted terms are rejected with POLICY_VIOLATION.
func TestCommentBlocklist(t *testing.T) {
	t.Setenv("COMMENT_BLOCKLIST", "spamword,scamlink")
	srv := newCommentsTestServer()
	defer srv.Close()

	req, rec := commentsReq(http.MethodPost, "/videos/00000000-0000-0000-0000-000000000001/comments", "viewer-token", map[string]string{
		"content": "hello world",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 clean comment, got %d: %s", rec.Code, rec.Body.String())
	}

	req, rec = commentsReq(http.MethodPost, "/videos/00000000-0000-0000-0000-000000000001/comments", "viewer-token", map[string]string{
		"content": "buy spamword now cheap",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 blocked comment, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "POLICY_VIOLATION") {
		t.Fatalf("expected POLICY_VIOLATION, got %s", rec.Body.String())
	}
}
