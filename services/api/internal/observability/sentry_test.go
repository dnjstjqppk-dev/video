package observability

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMiddlewareRecoversPanic(t *testing.T) {
	panicker := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	panicker.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if body := rec.Body.String(); body == "" || !strings.Contains(body, "INTERNAL") {
		t.Fatalf("expected generic error envelope, got %q", body)
	}
}

func TestMiddlewarePassesThrough(t *testing.T) {
	ok := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	ok.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Fatalf("expected passthrough 418, got %d", rec.Code)
	}
}

func TestCaptureErrorDisabledIsNoop(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	CaptureError(req, errors.New("no dsn configured"), map[string]interface{}{"k": "v"})
}

func TestInitEmptyDSNDisabled(t *testing.T) {
	if err := Init("", "test"); err != nil {
		t.Fatalf("expected nil for empty DSN, got %v", err)
	}
}
