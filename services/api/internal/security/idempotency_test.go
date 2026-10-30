package security

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestIdempotencyIsolationAndConflict(t *testing.T) {
	manager := NewIdempotencyManager(time.Minute)
	var calls int
	handler := manager.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, calls)
	}))
	request := func(token, method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", token)
		req.Header.Set("X-Idempotency-Key", "same-key")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		return rec
	}
	if request("Bearer a", "POST", "/a", "one").Body.String() != "1" {
		t.Fatal("first request")
	}
	if request("Bearer a", "POST", "/a", "one").Body.String() != "1" || calls != 1 {
		t.Fatal("replay executed")
	}
	if request("Bearer a", "POST", "/a", "two").Code != http.StatusConflict {
		t.Fatal("body conflict accepted")
	}
	for _, args := range [][3]string{{"Bearer b", "POST", "/a"}, {"Bearer a", "POST", "/b"}, {"Bearer a", "PATCH", "/a"}} {
		if request(args[0], args[1], args[2], "one").Header().Get("X-Cache-Lookup") != "" {
			t.Fatal("cross-scope replay")
		}
	}
	if calls != 4 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestIdempotencyConcurrentRequestsExecuteOnce(t *testing.T) {
	manager := NewIdempotencyManager(time.Minute)
	var calls atomic.Int32
	handler := manager.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, "created")
	}))
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("POST", "/a", strings.NewReader("body"))
			req.Header.Set("Authorization", "Bearer a")
			req.Header.Set("X-Idempotency-Key", "same")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != 200 || rec.Body.String() != "created" {
				t.Errorf("unexpected response %d %s", rec.Code, rec.Body.String())
			}
		}()
	}
	wg.Wait()
	if calls.Load() != 1 {
		t.Fatalf("executed %d times", calls.Load())
	}
}

func TestIdempotencyRetriesFailures(t *testing.T) {
	manager := NewIdempotencyManager(time.Minute)
	calls := 0
	handler := manager.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}))
	for _, want := range []int{503, 201, 201} {
		req := httptest.NewRequest("POST", "/a", strings.NewReader("body"))
		req.Header.Set("Authorization", "Bearer a")
		req.Header.Set("X-Idempotency-Key", "same")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("got %d want %d", rec.Code, want)
		}
	}
	if calls != 2 {
		t.Fatalf("calls=%d", calls)
	}
}
