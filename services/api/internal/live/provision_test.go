package live

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"globalvideoplatform/services/api/internal/store"
)

const streamStubResponse = `{"success":true,"result":{
  "uid":"uid-live-1",
  "rtmps":{"url":"rtmps://live.cloudflarestream.com:443/live","streamKey":"sk_live_abc"},
  "srt":{"url":"srt://live.cloudflarestream.com:778","streamId":"sid","passphrase":"pp"},
  "playback":{"hls":"https://customer-test.cloudflarestream.com/uid-live-1/manifest/video.m3u8","dash":"https://x/dash"},
  "webRTC":{"url":"https://customer-test.cloudflarestream.com/uid-live-1/webrtc"}
}}`

func streamStub(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/stream/live_inputs") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, streamStubResponse)
	}))
}

// P1-15: with Stream credentials, creation provisions a real live input.
func TestCreateStreamProvisionsLiveInput(t *testing.T) {
	stub := streamStub(t)
	defer stub.Close()
	t.Setenv("CF_STREAM_ACCOUNT_ID", "acct_1")
	t.Setenv("CF_STREAM_API_TOKEN", "tok_1")
	t.Setenv("CF_STREAM_BASE_URL", stub.URL)

	srv, _ := newLiveTestServer()
	defer srv.Close()

	req, rec := liveReq(http.MethodPost, "/live-streams", "creator-token", map[string]string{
		"title": "Provisioned Live",
	})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var stream store.LiveStream
	if err := json.Unmarshal(rec.Body.Bytes(), &stream); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stream.StreamKey != "sk_live_abc" || stream.RtmpsIngestURL == "" || stream.PlaybackHlsURL == "" {
		t.Fatalf("provisioned ingest missing: %+v", stream)
	}
}

// P1-15: strangers cannot start, stop, or rotate another creator's stream.
func TestStreamControlOwnership(t *testing.T) {
	srv, _ := newLiveTestServer()
	defer srv.Close()

	// Seed stream owned by the creator channel.
	req, rec := liveReq(http.MethodPost, "/live-streams", "creator-token", map[string]string{"title": "Owned"})
	srv.Config.Handler.ServeHTTP(rec, req)
	var stream store.LiveStream
	if err := json.Unmarshal(rec.Body.Bytes(), &stream); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/live-streams/" + stream.ID + "/start"},
		{http.MethodPost, "/live-streams/" + stream.ID + "/stop"},
		{http.MethodPost, "/live-streams/" + stream.ID + "/rotate-key"},
	} {
		req, rec := liveReq(tc.method, tc.path, "viewer-token", nil)
		srv.Config.Handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("%s %s: expected 403 stranger, got %d", tc.method, tc.path, rec.Code)
		}
	}

	req, rec = liveReq(http.MethodPost, "/live-streams/"+stream.ID+"/start", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 owner start, got %d", rec.Code)
	}
}
