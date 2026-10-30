package uploads_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	srv := newUploadsTestServer()
	defer srv.Close()

	// 1. Initiate upload session
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
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &sessionResp); err != nil || sessionResp.UploadID == "" {
		t.Fatalf("failed to decode upload response: %v", err)
	}
	if sessionResp.TotalParts < 1 {
		t.Fatalf("expected at least 1 part, got %d", sessionResp.TotalParts)
	}

	// 2. Request presigned URL for Part 1
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

	// 3. Complete Multipart Upload
	completePayload := map[string]interface{}{
		"parts": []map[string]interface{}{
			{"partNumber": 1, "etag": "etag-part-1-md5"},
		},
	}
	req, rec = uploadReq(http.MethodPost, "/uploads/"+sessionResp.UploadID+"/complete", "creator-token", completePayload)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for complete upload, got %d: %s", rec.Code, rec.Body.String())
	}

	// 4. Query status
	req, rec = uploadReq(http.MethodGet, "/uploads/"+sessionResp.UploadID+"/status", "creator-token", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for upload status, got %d", rec.Code)
	}
}
