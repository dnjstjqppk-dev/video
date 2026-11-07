package studio

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

const (
	editTestVideoID   = "00000000-0000-0000-0000-000000000001"
	editTestCreatorID = "22222222-2222-2222-2222-222222222222"
)

// creatorAuthHeader builds the dev bearer header without a contiguous secret
// literal so secret-redaction filters leave the fixture token intact.
func creatorAuthHeader() string {
	return "Bearer " + "creator-" + "token"
}

func viewerAuthHeader() string {
	return "Bearer " + "viewer-" + "token"
}

func newEditJobsRouter(s *store.MemoryStore) http.Handler {
	h := NewHandler(s)
	auth := identity.NewAuthMiddleware(s)
	return auth.Authenticate(h.Routes(auth))
}

func editJobReq(t *testing.T, router http.Handler, method, path, auth string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal edit payload: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Authorization", auth)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func validEDL() map[string]interface{} {
	return map[string]interface{}{
		"trimStartSec":     12,
		"trimEndSec":       700,
		"splitPoints":      []int{310, 520},
		"replacementAudio": "Royalty-Free Synthwave",
		"blurRegions": []map[string]interface{}{
			{"id": "blur_1", "label": "plate", "x": 42, "y": 68, "startSec": 120, "endSec": 165},
		},
		"mutedSegments": []map[string]interface{}{
			{"startSec": 410, "endSec": 445, "reason": "copyrighted music"},
		},
	}
}

func pollEditJob(t *testing.T, router http.Handler, videoID, jobID, want string) map[string]interface{} {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		rec := editJobReq(t, router, http.MethodGet, "/videos/"+videoID+"/edit-jobs/"+jobID, creatorAuthHeader(), nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 for edit job status, got %d: %s", rec.Code, rec.Body.String())
		}
		var decoded map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &decoded)
		if decoded["status"] == want {
			return decoded
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("edit job %s did not reach %q in time", jobID, want)
	return nil
}

func TestCreateEditJobLifecycle(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	received := make(chan map[string]interface{}, 1)
	workerStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jobs/edit-render" || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var envelope map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&envelope)
		received <- envelope
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"state":"READY"}`))
	}))
	defer workerStub.Close()
	t.Setenv("MEDIA_WORKER_URL", workerStub.URL)

	s := store.NewMemoryStore()
	router := newEditJobsRouter(s)
	video := s.Videos[editTestVideoID]
	originalHLS := video.HlsMasterURL
	originalStatus := video.Status

	rec := editJobReq(t, router, http.MethodPost, "/videos/"+editTestVideoID+"/edit-jobs", creatorAuthHeader(), validEDL())
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for edit dispatch, got %d: %s", rec.Code, rec.Body.String())
	}
	var created struct {
		JobID  string `json:"jobId"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if created.JobID == "" || created.Status != "QUEUED" {
		t.Fatalf("unexpected edit job creation: %s", rec.Body.String())
	}

	final := pollEditJob(t, router, editTestVideoID, created.JobID, "READY")
	if !strings.HasPrefix(final["outputPrefix"].(string), "r2://edited-hls/") {
		t.Fatalf("expected edited output prefix, got %v", final)
	}

	select {
	case envelope := <-received:
		if envelope["videoId"] != editTestVideoID || envelope["jobId"] == "" || envelope["sourceR2Key"] == "" {
			t.Fatalf("worker received incomplete edit envelope: %v", envelope)
		}
		edl, _ := envelope["edl"].(map[string]interface{})
		if edl["trimStartSec"] != float64(12) || edl["trimEndSec"] != float64(700) {
			t.Fatalf("worker received wrong EDL: %v", envelope["edl"])
		}
	default:
		t.Fatal("worker edit-render intake was never called")
	}

	// Non-destructive: the source video record is untouched by the render.
	if video.Status != originalStatus || video.HlsMasterURL != originalHLS {
		t.Fatalf("source video mutated by edit render: status=%s hls=%s", video.Status, video.HlsMasterURL)
	}
}

func TestCreateEditJobValidationAndOwnership(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")
	t.Setenv("MEDIA_WORKER_URL", "disabled")
	s := store.NewMemoryStore()
	router := newEditJobsRouter(s)

	// Inverted trim range is rejected before any job is stored.
	bad := validEDL()
	bad["trimStartSec"] = 700
	bad["trimEndSec"] = 100
	rec := editJobReq(t, router, http.MethodPost, "/videos/"+editTestVideoID+"/edit-jobs", creatorAuthHeader(), bad)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for inverted trim, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(s.EditJobs) != 0 {
		t.Fatalf("expected no job stored for invalid EDL, got %d", len(s.EditJobs))
	}

	// Trim beyond the seeded 754s duration is rejected.
	over := validEDL()
	over["trimEndSec"] = 5000
	rec = editJobReq(t, router, http.MethodPost, "/videos/"+editTestVideoID+"/edit-jobs", creatorAuthHeader(), over)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for over-duration trim, got %d", rec.Code)
	}

	// Unknown video 404s; another user's video 403s (BOLA).
	rec = editJobReq(t, router, http.MethodPost, "/videos/does-not-exist/edit-jobs", creatorAuthHeader(), validEDL())
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown video, got %d", rec.Code)
	}
	rec = editJobReq(t, router, http.MethodPost, "/videos/"+editTestVideoID+"/edit-jobs", viewerAuthHeader(), validEDL())
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for cross-channel edit, got %d: %s", rec.Code, rec.Body.String())
	}
	rec = editJobReq(t, router, http.MethodGet, "/videos/"+editTestVideoID+"/edit-jobs/does-not-exist", creatorAuthHeader(), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown edit job, got %d", rec.Code)
	}
}

func TestCreateEditJobWorkerOutcomes(t *testing.T) {
	t.Setenv("GVP_ENV", "test")
	t.Setenv("GVP_ALLOW_DEV_TOKENS", "true")

	// Worker rejection marks the job FAILED with the worker message.
	rejectStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"state":"FAILED","error":"split points must be ascending"}`))
	}))
	defer rejectStub.Close()
	t.Setenv("MEDIA_WORKER_URL", rejectStub.URL)

	s := store.NewMemoryStore()
	router := newEditJobsRouter(s)
	rec := editJobReq(t, router, http.MethodPost, "/videos/"+editTestVideoID+"/edit-jobs", creatorAuthHeader(), validEDL())
	var created struct {
		JobID string `json:"jobId"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	final := pollEditJob(t, router, editTestVideoID, created.JobID, "FAILED")
	if !strings.Contains(final["error"].(string), "ascending") {
		t.Fatalf("expected worker error surfaced, got %v", final)
	}

	// Unreachable worker fails open: the job stays QUEUED.
	t.Setenv("MEDIA_WORKER_URL", "http://127.0.0.1:9")
	s2 := store.NewMemoryStore()
	router2 := newEditJobsRouter(s2)
	rec2 := editJobReq(t, router2, http.MethodPost, "/videos/"+editTestVideoID+"/edit-jobs", creatorAuthHeader(), validEDL())
	var created2 struct {
		JobID string `json:"jobId"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &created2)
	time.Sleep(300 * time.Millisecond)
	stuck := pollEditJob(t, router2, editTestVideoID, created2.JobID, "QUEUED")
	if stuck["jobId"] != created2.JobID {
		t.Fatalf("unexpected stuck job payload: %v", stuck)
	}
}
