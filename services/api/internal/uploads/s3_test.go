package uploads_test

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/uploads"
)

func testS3Config(endpoint string) uploads.S3Config {
	return uploads.S3Config{
		Endpoint:       endpoint,
		PublicEndpoint: endpoint,
		Bucket:         "gvp-raw-uploads",
		AccessKey:      "test-access-key",
		SecretKey:      "test-secret-key",
		Region:         "auto",
	}
}

func TestPresignPartUploadStructure(t *testing.T) {
	cfg := testS3Config("http://localhost:9002")
	fixed := time.Date(2026, 11, 6, 7, 0, 0, 0, time.UTC)

	signed, err := cfg.PresignPartUpload("raw/user1/vid.mp4", "s3-upload-id-1", 3, time.Hour, fixed)
	if err != nil {
		t.Fatalf("presign failed: %v", err)
	}
	parsed, err := url.Parse(signed)
	if err != nil {
		t.Fatalf("presigned URL does not parse: %v", err)
	}
	if parsed.Path != "/gvp-raw-uploads/raw/user1/vid.mp4" {
		t.Fatalf("unexpected presigned path: %s", parsed.Path)
	}
	q := parsed.Query()
	if q.Get("partNumber") != "3" || q.Get("uploadId") != "s3-upload-id-1" {
		t.Fatalf("part parameters missing: %s", parsed.RawQuery)
	}
	if q.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		t.Fatalf("missing SigV4 algorithm: %s", parsed.RawQuery)
	}
	if q.Get("X-Amz-Expires") != "3600" {
		t.Fatalf("unexpected expiry: %s", q.Get("X-Amz-Expires"))
	}
	sig := q.Get("X-Amz-Signature")
	if len(sig) != 64 {
		t.Fatalf("signature must be 64 hex chars, got %q", sig)
	}
	for _, c := range sig {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Fatalf("signature must be lowercase hex, got %q", sig)
		}
	}
}

func TestPresignIsDeterministicAndKeyBound(t *testing.T) {
	cfg := testS3Config("http://localhost:9002")
	fixed := time.Date(2026, 11, 6, 7, 0, 0, 0, time.UTC)

	first, err := cfg.PresignPartUpload("raw/u/v.mp4", "up-1", 1, time.Hour, fixed)
	if err != nil {
		t.Fatalf("presign failed: %v", err)
	}
	second, err := cfg.PresignPartUpload("raw/u/v.mp4", "up-1", 1, time.Hour, fixed)
	if err != nil || first != second {
		t.Fatalf("identical inputs must produce identical URLs: %q vs %q (err=%v)", first, second, err)
	}

	other := cfg
	other.SecretKey = "different-secret"
	third, err := other.PresignPartUpload("raw/u/v.mp4", "up-1", 1, time.Hour, fixed)
	if err != nil {
		t.Fatalf("presign failed: %v", err)
	}
	if first == third {
		t.Fatalf("different secret keys must produce different signatures")
	}

	if _, err := cfg.PresignPartUpload("raw/u/v.mp4", "up-1", 0, time.Hour, fixed); err == nil {
		t.Fatalf("part number 0 must be rejected")
	}
	if _, err := cfg.PresignPartUpload("", "up-1", 1, time.Hour, fixed); err == nil {
		t.Fatalf("empty key must be rejected")
	}
}

func TestS3MultipartFlowAgainstStub(t *testing.T) {
	var sawAuth bool
	var sawCompleteXML string
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 ") {
			sawAuth = true
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Query().Has("uploads"):
			w.Header().Set("Content-Type", "application/xml")
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><InitiateMultipartUploadResult><UploadId>stub-upload-9</UploadId></InitiateMultipartUploadResult>`))
		case r.Method == http.MethodPost && r.URL.Query().Get("uploadId") != "":
			body, _ := io.ReadAll(r.Body)
			sawCompleteXML = string(body)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer stub.Close()

	cfg := testS3Config(stub.URL)
	ctx := context.Background()

	if err := cfg.EnsureBucket(ctx); err != nil {
		t.Fatalf("ensure bucket failed: %v", err)
	}
	uploadID, err := cfg.CreateMultipartUpload(ctx, "raw/u/v.mp4", "video/mp4")
	if err != nil {
		t.Fatalf("create multipart failed: %v", err)
	}
	if uploadID != "stub-upload-9" {
		t.Fatalf("unexpected upload id %q", uploadID)
	}
	if !sawAuth {
		t.Fatalf("stub never observed a SigV4 Authorization header")
	}

	err = cfg.CompleteMultipartUpload(ctx, "raw/u/v.mp4", uploadID, []uploads.PartETag{
		{PartNumber: 1, ETag: "aaa"},
		{PartNumber: 2, ETag: `"bbb"`},
	})
	if err != nil {
		t.Fatalf("complete multipart failed: %v", err)
	}
	var parsed struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []struct {
			Number int    `xml:"PartNumber"`
			ETag   string `xml:"ETag"`
		} `xml:"Part"`
	}
	if err := xml.Unmarshal([]byte(sawCompleteXML), &parsed); err != nil {
		t.Fatalf("complete body is not valid XML: %v (%s)", err, sawCompleteXML)
	}
	if len(parsed.Parts) != 2 || parsed.Parts[0].Number != 1 || parsed.Parts[1].ETag != `"bbb"` {
		t.Fatalf("unexpected complete XML: %s", sawCompleteXML)
	}
}
