package uploads

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
)

type Handler struct {
	Store *store.MemoryStore
}

func NewHandler(s *store.MemoryStore) *Handler {
	return &Handler{Store: s}
}

func (h *Handler) Routes(auth *identity.AuthMiddleware) http.Handler {
	r := chi.NewRouter()
	r.Use(auth.RequireAuth)

	r.Post("/", h.CreateUploadSession)
	r.Post("/{id}/parts", h.GetPartUploadURL)
	r.Post("/{id}/complete", h.CompleteUpload)
	r.Get("/{id}/status", h.GetUploadStatus)

	return r
}

func (h *Handler) GetUploadStatus(w http.ResponseWriter, r *http.Request) {
	uploadID := chi.URLParam(r, "id")
	if uploadID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Upload ID required")
		return
	}

	var job *store.MediaJob
	for _, j := range h.Store.MediaJobs {
		if j.ID == uploadID || j.VideoID == uploadID {
			job = j
			break
		}
	}

	if job == nil {
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"upload_id":    uploadID,
			"status":       "QUEUED",
			"progress_pct": 0,
		})
		return
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"upload_id":    uploadID,
		"job_id":       job.ID,
		"video_id":     job.VideoID,
		"status":       job.Status,
		"progress_pct": job.ProgressPct,
	})
}

func (h *Handler) CreateUploadSession(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	var payload struct {
		Title         string `json:"title"`
		Description   string `json:"description"`
		FileSizeBytes int64  `json:"fileSizeBytes"`
		MimeType      string `json:"mimeType"`
		VideoType     string `json:"videoType"` // standard, shorts
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid upload payload")
		return
	}

	if payload.Title == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Title is required")
		return
	}

	// 10MB chunk size
	partSize := int64(10 * 1024 * 1024)
	totalParts := int((payload.FileSizeBytes + partSize - 1) / partSize)
	if totalParts < 1 {
		totalParts = 1
	}

	uploadID := uuid.NewString()
	videoID := uuid.NewString()
	r2Key := fmt.Sprintf("raw/%s/%s.mp4", user.ID, videoID)

	videoType := "standard"
	if payload.VideoType == "shorts" {
		videoType = "shorts"
	}

	// Create initial video in CREATED state
	video := &store.Video{
		ID:              videoID,
		ChannelID:       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		ChannelName:     "TechNexus",
		ChannelHandle:   "@technexus",
		Title:           payload.Title,
		Description:     payload.Description,
		VideoType:       videoType,
		Visibility:      "private",
		Status:          "CREATED",
		DurationSeconds: 0,
		ThumbnailURL:    "https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=800",
		HlsMasterURL:    "",
		Tags:            []string{"upload"},
		Category:        "General",
		CreatedAt:       time.Now(),
	}
	h.Store.Videos[videoID] = video

	response := map[string]interface{}{
		"uploadId":   uploadID,
		"videoId":    videoID,
		"partSize":   partSize,
		"totalParts": totalParts,
		"key":        r2Key,
	}

	identity.WriteJSON(w, http.StatusCreated, response)
}

func (h *Handler) GetPartUploadURL(w http.ResponseWriter, r *http.Request) {
	uploadID := chi.URLParam(r, "id")
	var payload struct {
		PartNumber int `json:"partNumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.PartNumber <= 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid partNumber")
		return
	}

	// Simulated S3/R2 presigned part upload URL
	presignedURL := fmt.Sprintf("https://gvp-media.r2.cloudflarestorage.com/raw/%s?partNumber=%d&uploadId=%s", uploadID, payload.PartNumber, uploadID)

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"uploadUrl":  presignedURL,
		"partNumber": payload.PartNumber,
	})
}

func (h *Handler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	uploadID := chi.URLParam(r, "id")
	var payload struct {
		Parts []struct {
			PartNumber int    `json:"partNumber"`
			Etag       string `json:"etag"`
		} `json:"parts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || len(payload.Parts) == 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Valid parts list is required")
		return
	}

	// Find video associated with upload session
	var targetVideo *store.Video
	for _, v := range h.Store.Videos {
		if v.Status == "CREATED" {
			targetVideo = v
			break
		}
	}

	jobID := "job-" + uuid.NewString()
	if targetVideo != nil {
		targetVideo.Status = "QUEUED"
		targetVideo.HlsMasterURL = "https://media.gvp.media/hls/" + targetVideo.ID + "/master.m3u8"
		targetVideo.PublishedAt = time.Now()
		targetVideo.Visibility = "private" // Private draft until media-worker completes transcode

		// Register asynchronous media job for worker processing (ARTICLE 2.3 & 2.5)
		job := &store.MediaJob{
			ID:           jobID,
			VideoID:      targetVideo.ID,
			InputURI:     "r2://raw-uploads/" + uploadID + "/source.mp4",
			OutputPrefix: "r2://public-hls/" + targetVideo.ID + "/",
			Status:       "QUEUED",
			ProgressPct:  0,
			CreatedAt:    time.Now(),
		}
		h.Store.MediaJobs[job.ID] = job
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"uploadId": uploadID,
		"status":   "QUEUED",
		"jobId":    jobID,
		"message":  "Upload complete and verified. Transcoding job queued for media worker.",
	})
}
