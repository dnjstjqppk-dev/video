package uploads_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	pollUploadStatus(t, srv.Config.Handler, uploadID, "READY", 5*time.Second)

	select {
	case envelope := <-received:
		if envelope["videoId"] != videoID || envelope["jobId"] == "" || envelope["sourceR2Key"] == "" {
			t.Fatalf("worker received incomplete envelope: %v", envelope)
		}
	default:
		t.Fatal("worker intake was never called")
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
