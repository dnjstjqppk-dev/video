package uploads_test

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/uploads"
)

func newUploadsTestServer() *httptest.Server {
	memStore := store.NewMemoryStore()
	auth := &identity.AuthMiddleware{Store: memStore, AllowDevTokens: true}
	handler := uploads.NewHandler(memStore)

	r := chi.NewRouter()
	r.Use(auth.Authenticate)
	r.Mount("/uploads", handler.Routes(auth))

	return httptest.NewServer(r)
}

func uploadReq(method, path, token string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
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

func TestUploadLifecycle(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	srv := newUploadsTestServer()
	defer srv.Close()

	// 1. Initiate upload session (local relay mode)
	initPayload := map[string]interface{}{
		"title":         "4K Tech Walkthrough",
		"description":   "HLS transcoding pipeline test video",
		"fileSizeBytes": 52428800, // 50MB
		"mimeType":      "video/mp4",
		"videoType":     "standard",
	}
	req, rec := uploadReq(http.MethodPost, "/uploads", "creator-token", initPayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for upload creation, got %d: %s", rec.Code, rec.Body.String())
	}

	var sessionResp struct {
		UploadID   string `json:"uploadId"`
		VideoID    string `json:"videoId"`
		PartSize   int64  `json:"partSize"`
		TotalParts int    `json:"totalParts"`
		Key        string `json:"key"`
		Mode       string `json:"mode"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sessionResp); err != nil || sessionResp.UploadID == "" {
		t.Fatalf("failed to decode upload response: %v", err)
	}
	if sessionResp.TotalParts < 1 {
		t.Fatalf("expected at least 1 part, got %d", sessionResp.TotalParts)
	}
	if sessionResp.Mode != "local" {
		t.Fatalf("expected local relay mode without R2 credentials, got %q", sessionResp.Mode)
	}

	// 2. Request relay URL for Part 1 and PUT real bytes
	partPayload := map[string]int{"partNumber": 1}
	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/parts", "creator-token", partPayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for part URL, got %d: %s", rec.Code, rec.Body.String())
	}
	var partResp struct {
		UploadURL  string `json:"uploadUrl"`
		PartNumber int    `json:"partNumber"`
	}
	json.Unmarshal(rec.Body.Bytes(), &partResp)
	if partResp.PartNumber != 1 || partResp.UploadURL == "" {
		t.Fatalf("invalid part presigned response: %+v", partResp)
	}
	relayPath, err := url.Parse(partResp.UploadURL)
	if err != nil || !strings.HasSuffix(relayPath.Path, "/chunks/1") {
		t.Fatalf("expected local relay chunk URL, got %q", partResp.UploadURL)
	}

	chunk := bytes.Repeat([]byte("0123456789abcdef"), 64*1024) // 1MB of real bytes
	putReq := httptest.NewRequest(http.MethodPut, strings.TrimPrefix(relayPath.Path, "/v1"), bytes.NewReader(chunk))
	putReq.Header.Set("Authorization", "Bearer creator-token")
	putReq.Header.Set("Content-Type", "video/mp4")
	putRec := httptest.NewRecorder()
	srv.Config.Handler.ServeHTTP(putRec, putReq)
	if putRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for chunk PUT, got %d: %s", putRec.Code, putRec.Body.String())
	}
	var chunkResp struct {
		ETag string `json:"etag"`
		Size int    `json:"size"`
	}
	json.Unmarshal(putRec.Body.Bytes(), &chunkResp)
	sum := md5.Sum(chunk)
	if chunkResp.ETag != hex.EncodeToString(sum[:]) || chunkResp.Size != len(chunk) {
		t.Fatalf("chunk ETag/size mismatch: %+v", chunkResp)
	}

	// Remaining parts must also be uploaded before completion
	for part := 2; part <= sessionResp.TotalParts; part++ {
		extra := bytes.Repeat([]byte{byte(part)}, 1024)
		preq := httptest.NewRequest(http.MethodPut, "/uploads/"+sessionResp.UploadID+"/chunks/"+strconv.Itoa(part), bytes.NewReader(extra))
		preq.Header.Set("Authorization", "Bearer creator-token")
		prec := httptest.NewRecorder()
		srv.Config.Handler.ServeHTTP(prec, preq)
		if prec.Code != http.StatusOK {
			t.Fatalf("expected 200 for chunk %d PUT, got %d", part, prec.Code)
		}
	}

	// 3. Complete Multipart Upload
	parts := make([]map[string]interface{}, 0, sessionResp.TotalParts)
	for part := 1; part <= sessionResp.TotalParts; part++ {
		parts = append(parts, map[string]interface{}{"partNumber": part, "etag": "etag-part"})
	}
	completePayload := map[string]interface{}{"parts": parts}
	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/complete", "creator-token", completePayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for complete upload, got %d: %s", rec.Code, rec.Body.String())
	}
	var completeResp struct {
		VideoID string `json:"videoId"`
		Status  string `json:"status"`
		JobID   string `json:"jobId"`
	}
	json.Unmarshal(rec.Body.Bytes(), &completeResp)
	if completeResp.VideoID != sessionResp.VideoID || completeResp.Status != "QUEUED" || completeResp.JobID == "" {
		t.Fatalf("unexpected complete response: %+v", completeResp)
	}

	// Assembled bytes must exist on disk
	staged := filepath.Join(os.Getenv("GVP_LOCAL_MEDIA_DIR"), sessionResp.VideoID, "source.mp4")
	info, err := os.Stat(staged)
	if err != nil || info.Size() == 0 {
		t.Fatalf("expected assembled source file at %s (err=%v)", staged, err)
	}

	// 4. Query status
	req, rec = uploadReq(http.MethodGet, "/uploads/"+sessionResp.UploadID+"/status", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for upload status, got %d", rec.Code)
	}
}

func TestUploadSessionIsolation(t *testing.T) {
	t.Setenv("GVP_LOCAL_MEDIA_DIR", t.TempDir())
	srv := newUploadsTestServer()
	defer srv.Close()

	initPayload := map[string]interface{}{"title": "Isolation Check", "fileSizeBytes": 1024, "mimeType": "video/mp4"}
	req, rec := uploadReq(http.MethodPost, "/uploads", "creator-token", initPayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var sessionResp struct {
		UploadID string `json:"uploadId"`
	}
	json.Unmarshal(rec.Body.Bytes(), &sessionResp)

	// Another user must be rejected (BOLA)
	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/parts", "viewer-token", map[string]int{"partNumber": 1})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-user part URL, got %d", rec.Code)
	}

	// Unknown sessions must 404 instead of returning fake progress
	req, rec = uploadReq(http.MethodGet, "/uploads/does-not-exist/status", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown session status, got %d", rec.Code)
	}

	// Completion without chunk bytes must fail honestly
	completePayload := map[string]interface{}{"parts": []map[string]interface{}{{"partNumber": 1, "etag": "x"}}}
	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/complete", "creator-token", completePayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for incomplete upload, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestS3ModeSessionUsesStub(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Query().Has("uploads") {
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(`<InitiateMultipartUploadResult><UploadId>stub-up-1</UploadId></InitiateMultipartUploadResult>`))
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer stub.Close()

	t.Setenv("R2_ENDPOINT", stub.URL)
	t.Setenv("R2_PUBLIC_ENDPOINT", stub.URL)
	t.Setenv("R2_ACCESS_KEY_ID", "stub-key")
	t.Setenv("R2_SECRET_ACCESS_KEY", "stub-secret")
	t.Setenv("R2_BUCKET_RAW", "gvp-raw-uploads")

	srv := newUploadsTestServer()
	defer srv.Close()

	initPayload := map[string]interface{}{"title": "S3 Mode", "fileSizeBytes": 20971520, "mimeType": "video/mp4"}
	req, rec := uploadReq(http.MethodPost, "/uploads", "creator-token", initPayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var sessionResp struct {
		UploadID string `json:"uploadId"`
		Mode     string `json:"mode"`
	}
	json.Unmarshal(rec.Body.Bytes(), &sessionResp)
	if sessionResp.Mode != "s3" {
		t.Fatalf("expected s3 mode with credentials configured, got %q", sessionResp.Mode)
	}

	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/parts", "creator-token", map[string]int{"partNumber": 1})
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for presigned part URL, got %d: %s", rec.Code, rec.Body.String())
	}
	var partResp struct {
		UploadURL string `json:"uploadUrl"`
	}
	json.Unmarshal(rec.Body.Bytes(), &partResp)
	parsed, err := url.Parse(partResp.UploadURL)
	if err != nil {
		t.Fatalf("presigned URL does not parse: %v", err)
	}
	q := parsed.Query()
	if q.Get("X-Amz-Signature") == "" || q.Get("uploadId") != "stub-up-1" || q.Get("partNumber") != "1" {
		t.Fatalf("presigned URL missing SigV4 multipart params: %s", partResp.UploadURL)
	}
}
