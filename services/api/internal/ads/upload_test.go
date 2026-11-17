package ads

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

func uploadRouter(s *store.MemoryStore, dir string) http.Handler {
	auth := identity.NewAuthMiddleware(s)
	h := NewHandler(s)
	h.CreativeDir = dir
	r := auth.Authenticate(h.Routes(auth))
	return r
}

// Minimal file magic so MIME sniffing classifies correctly.
var (
	// 32-byte ftyp box: size(4) + "ftyp" + major(4) + minor(4) + compatibles(16).
	fakeMP4 = append([]byte{0x00, 0x00, 0x00, 0x20}, []byte("ftypmp42")...)
	fakePNG = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D}
)

func init() {
	fakeMP4 = append(fakeMP4, make([]byte, 20)...)
}

func multipartUpload(t *testing.T, router http.Handler, token, campaignID, filename string, content []byte, fields map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/campaigns/"+campaignID+"/creatives", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func makeDraftCampaign(t *testing.T, router http.Handler, format string) string {
	t.Helper()
	rec := postAds(t, router, "/campaigns", "advertiser-token", map[string]interface{}{
		"name": "upload-pipeline", "format": format, "totalBudget": 100000, "dailyBudget": 10000,
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create campaign: %d", rec.Code)
	}
	var camp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &camp); err != nil {
		t.Fatalf("decode campaign: %v", err)
	}
	return camp.ID
}

// P0-5: advertisers upload real creative bytes; the record enters pending review.
func TestCreativeUploadPipeline(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	dir := t.TempDir()
	router := uploadRouter(s, dir)
	campID := makeDraftCampaign(t, router, "preroll")

	rec := multipartUpload(t, router, "advertiser-token", campID, "spot.mp4", fakeMP4, map[string]string{
		"headline": "Pipe Test", "destinationUrl": "https://advertiser.test/lp", "callToAction": "Watch",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 upload, got %d: %s", rec.Code, rec.Body.String())
	}
	var creative struct {
		ID           string `json:"id"`
		CampaignID   string `json:"campaign_id"`
		MediaURL     string `json:"media_url"`
		ReviewStatus string `json:"review_status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &creative); err != nil {
		t.Fatalf("decode creative: %v", err)
	}
	if creative.CampaignID != campID || creative.ReviewStatus != "pending" || creative.MediaURL == "" {
		t.Fatalf("bad creative record: %+v", creative)
	}
	if entries, err := os.ReadDir(filepath.Join(dir, campID)); err != nil || len(entries) != 1 {
		t.Fatalf("expected 1 stored file, err=%v entries=%v", err, entries)
	}

	// Served bytes match the upload.
	req := httptest.NewRequest(http.MethodGet, "/creatives/"+creative.ID+"/asset", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req)
	// Pending creatives are not public; owner fetch works.
	if rec2.Code != http.StatusUnauthorized && rec2.Code != http.StatusForbidden {
		t.Fatalf("expected auth gate on pending asset, got %d", rec2.Code)
	}
	reqAuth := httptest.NewRequest(http.MethodGet, "/creatives/"+creative.ID+"/asset", nil)
	reqAuth.Header.Set("Authorization", "Bearer advertiser-token")
	rec3 := httptest.NewRecorder()
	router.ServeHTTP(rec3, reqAuth)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 owner asset fetch, got %d", rec3.Code)
	}
	served, _ := io.ReadAll(rec3.Body)
	if !bytes.Equal(served, fakeMP4) {
		t.Fatal("served bytes must match the upload")
	}
}

// P0-5: upload validation rejects wrong types, mismatched formats, strangers.
func TestCreativeUploadValidation(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	s := store.NewMemoryStore()
	s.Users["55555555-5555-5555-5555-555555555555"] = &store.User{
		ID: "55555555-5555-5555-5555-555555555555", Email: "adv2@test.local",
		Role: "advertiser", IsActive: true,
	}
	router := uploadRouter(s, t.TempDir())
	campID := makeDraftCampaign(t, router, "preroll")

	// Anonymous refused.
	rec := multipartUpload(t, router, "", campID, "spot.mp4", fakeMP4, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 anonymous, got %d", rec.Code)
	}
	// Other advertiser refused.
	rec = multipartUpload(t, router, "55555555-5555-5555-5555-555555555555", campID, "spot.mp4", fakeMP4, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 cross-tenant, got %d", rec.Code)
	}
	// Executable refused.
	rec = multipartUpload(t, router, "advertiser-token", campID, "evil.exe", []byte("MZ fake binary payload..........."), nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 executable, got %d", rec.Code)
	}
	// Image bytes for a preroll (video) campaign refused.
	rec = multipartUpload(t, router, "advertiser-token", campID, "banner.png", fakePNG, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 format mismatch, got %d", rec.Code)
	}
}
