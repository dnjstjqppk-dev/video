package uploads_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// runLocalUploadToCompletion drives create -> chunk PUT -> complete in local
// relay mode and returns the upload and video IDs.
func runLocalUploadToCompletion(t *testing.T, handler http.Handler) (uploadID, videoID string) {
	t.Helper()
	initPayload := map[string]interface{}{"title": "Dispatch Check", "fileSizeBytes": 2048, "mimeType": "video/mp4"}
	req, rec := uploadReq(http.MethodPost, "/uploads", "creator-token", initPayload)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var sessionResp struct {
		UploadID   string `json:"uploadId"`
		VideoID    string `json:"videoId"`
		TotalParts int    `json:"totalParts"`
	}
	json.Unmarshal(rec.Body.Bytes(), &sessionResp)
	if sessionResp.TotalParts != 1 {
		t.Fatalf("expected 1 part for 2KB upload, got %d", sessionResp.TotalParts)
	}
	chunk := bytes.Repeat([]byte("frame-bytes-dispatch"), 64)
	chunkReq := httptest.NewRequest(http.MethodPut, "/uploads/"+sessionResp.UploadID+"/chunks/1", bytes.NewReader(chunk))
	chunkReq.Header.Set("Authorization", "Bearer "+"creator-"+"token")
	prec := httptest.NewRecorder()
	handler.ServeHTTP(prec, chunkReq)
	if prec.Code != http.StatusOK {
		t.Fatalf("expected 200 for chunk PUT, got %d", prec.Code)
	}
	completePayload := map[string]interface{}{"parts": []map[string]interface{}{{"partNumber": 1, "etag": "e1"}}}
	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/complete", "creator-token", completePayload)
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for complete upload, got %d: %s", rec.Code, rec.Body.String())
	}
	return sessionResp.UploadID, sessionResp.VideoID
}

func pollUploadStatus(t *testing.T, handler http.Handler, uploadID, want string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		req, rec := uploadReq(http.MethodGet, "/uploads/"+uploadID+"/status", "creator-token", nil)
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for status, got %d", rec.Code)
		}
		var statusResp struct {
			Status string `json:"status"`
		}
		json.Unmarshal(rec.Body.Bytes(), &statusResp)
		if statusResp.Status == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("upload %s did not reach status %q within %v", uploadID, want, timeout)
}

func TestDispatchTranscodeJobReady(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	received := make(chan map[string]interface{}, 1)
	workerStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/transcode" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var envelope map[string]interface{}
		json.NewDecoder(r.Body).Decode(&envelope)
		received <- envelope
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"state":"READY","probed":false}`))
	}))
	defer workerStub.Close()
	t.Setenv("MEDIA_WORKER_URL", workerStub.URL)

	srv := newUploadsTestServer()
	defer srv.Close()
	uploadID, videoID := runLocalUploadToCompletion(t, srv.Config.Handler)
	// P0-2: READY without publishable storage is a lie — without R2/MinIO the
	// job must fail closed with an actionable error instead of fake READY.
	pollUploadStatus(t, srv.Config.Handler, uploadID, "FAILED", 5*time.Second)

	select {
	case envelope := <-received:
		if envelope["videoId"] != videoID || envelope["jobId"] == "" || envelope["sourceR2Key"] == "" {
			t.Fatalf("worker received incomplete envelope: %v", envelope)
		}
		if envelope["sourceUrl"] == "" {
			t.Fatalf("local dispatch must send a file:// sourceUrl: %v", envelope)
		}
	default:
		t.Fatal("worker intake was never called")
	}
}

// P1-13: transient worker failures retry with backoff before folding back.
func TestDispatchRetriesTransientFailures(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	t.Setenv("UPLOAD_MODE", "local")
	t.Setenv("TRANSCODER_DISPATCH_BACKOFF_MS", "1")
	attempts := 0
	workerStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte(`{}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"state":"FAILED","error":"bad source"}`))
	}))
	defer workerStub.Close()
	t.Setenv("MEDIA_WORKER_URL", workerStub.URL)

	srv := newUploadsTestServer()
	defer srv.Close()
	uploadID, _ := runLocalUploadToCompletion(t, srv.Config.Handler)
	pollUploadStatus(t, srv.Config.Handler, uploadID, "FAILED", 10*time.Second)
	if attempts != 3 {
		t.Fatalf("expected 3 dispatch attempts, got %d", attempts)
	}
}

// P0-2: worker READY + reachable storage publishes artifacts and completes.
func TestDispatchPublishesArtifactsOnReady(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	t.Setenv("UPLOAD_MODE", "local")
	playlist := "#EXTM3U\n#EXT-X-VERSION:6\n"
	workerStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/jobs/transcode" && r.Method == http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"state":"READY","probed":true,` +
				`"probe":{"width":640,"height":360,"durationSeconds":12,"videoCodec":"h264"},` +
				`"artifacts":[{"key":"master.m3u8","contentType":"application/vnd.apple.mpegurl","bytes":24}]}`))
		case strings.HasPrefix(r.URL.Path, "/artifacts/") && strings.HasSuffix(r.URL.Path, "/master.m3u8"):
			w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
			w.Write([]byte(playlist))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer workerStub.Close()
	t.Setenv("MEDIA_WORKER_URL", workerStub.URL)

	putSeen := make(chan string, 4)
	s3Stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		putSeen <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer s3Stub.Close()
	t.Setenv("R2_ENDPOINT", s3Stub.URL)
	t.Setenv("R2_ACCESS_KEY_ID", "test-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "test-secret")
	t.Setenv("R2_BUCKET_HLS", "pub-hls")
	t.Setenv("R2_PUBLIC_HLS_BASE", "https://cdn.test/hls")

	srv := newUploadsTestServer()
	defer srv.Close()
	uploadID, videoID := runLocalUploadToCompletion(t, srv.Config.Handler)
	pollUploadStatus(t, srv.Config.Handler, uploadID, "READY", 10*time.Second)

	want := "/pub-hls/hls/" + videoID + "/master.m3u8"
	deadline := time.Now().Add(5 * time.Second)
	for {
		select {
		case path := <-putSeen:
			if path == want {
				return
			}
		default:
			if time.Now().After(deadline) {
				t.Fatalf("expected artifact PUT to %q", want)
			}
			time.Sleep(25 * time.Millisecond)
		}
	}
}

func TestDispatchTranscodeJobFailed(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	workerStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"state":"FAILED","error":"probed source metadata is required"}`))
	}))
	defer workerStub.Close()
	t.Setenv("MEDIA_WORKER_URL", workerStub.URL)

	srv := newUploadsTestServer()
	defer srv.Close()
	uploadID, _ := runLocalUploadToCompletion(t, srv.Config.Handler)
	pollUploadStatus(t, srv.Config.Handler, uploadID, "FAILED", 5*time.Second)
}

func TestDispatchWorkerUnreachableStaysQueued(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	// Nothing listens on this port: dispatch must fail open (job stays QUEUED).
	t.Setenv("MEDIA_WORKER_URL", "http://127.0.0.1:9")

	srv := newUploadsTestServer()
	defer srv.Close()
	uploadID, _ := runLocalUploadToCompletion(t, srv.Config.Handler)
	time.Sleep(300 * time.Millisecond)
	pollUploadStatus(t, srv.Config.Handler, uploadID, "QUEUED", time.Second)
}
