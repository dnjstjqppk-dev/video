package uploads

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/policies"
	"globalvideoplatform/services/api/internal/store"
)

const (
	defaultPartSize     = 10 * 1024 * 1024       // 10MB chunks
	maxLocalUploadBytes = 2 * 1024 * 1024 * 1024 // 2GB cap for local relay mode
	maxS3UploadBytes    = 50 * 1024 * 1024 * 1024
	chunkSlackBytes     = 5 * 1024 * 1024
	partURLTTL          = time.Hour
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
	r.Put("/{id}/chunks/{partNumber}", h.PutChunk)
	r.Post("/{id}/complete", h.CompleteUpload)
	r.Get("/{id}/status", h.GetUploadStatus)

	return r
}

func (h *Handler) loadSession(w http.ResponseWriter, r *http.Request) (*store.UploadSession, bool) {
	uploadID := chi.URLParam(r, "id")
	if uploadID == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Upload ID required")
		return nil, false
	}
	session, ok := h.Store.UploadSessions[uploadID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "UPLOAD_NOT_FOUND", "Upload session not found or expired")
		return nil, false
	}
	user := identity.GetUserFromContext(r.Context())
	if user == nil || (session.UserID != user.ID && user.Role != "super_admin") {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot access another user's upload session")
		return nil, false
	}
	return session, true
}

func (h *Handler) CreateUploadSession(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if !policies.FlagEnabled(h.Store, policies.FlagUploadIngest) {
		identity.WriteError(w, r, http.StatusForbidden, "POLICY_DISABLED", "Video upload ingest is paused by platform policy")
		return
	}
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
	if payload.FileSizeBytes <= 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "fileSizeBytes must be positive")
		return
	}

	s3cfg := S3ConfigFromEnv()
	mode := "local"
	if s3cfg.Configured() {
		mode = "s3"
	}
	maxBytes := int64(maxLocalUploadBytes)
	if mode == "s3" {
		maxBytes = maxS3UploadBytes
	}
	if policyCap := h.Store.Policy.MaxUploadSizeGB * 1024 * 1024 * 1024; policyCap > 0 && policyCap < maxBytes {
		maxBytes = policyCap
	}
	if payload.FileSizeBytes > maxBytes {
		identity.WriteError(w, r, http.StatusBadRequest, "QUOTA_EXCEEDED", "File exceeds the maximum upload size for this environment")
		return
	}

	mimeType := payload.MimeType
	if mimeType == "" {
		mimeType = "video/mp4"
	}

	partSize := int64(defaultPartSize)
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

	channel := h.Store.EnsureUserChannel(user)

	var s3UploadID string
	if mode == "s3" {
		ctx := r.Context()
		if err := s3cfg.EnsureBucket(ctx); err != nil {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Object storage is unreachable")
			return
		}
		s3ID, err := s3cfg.CreateMultipartUpload(ctx, r2Key, mimeType)
		if err != nil {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Could not initiate multipart upload")
			return
		}
		s3UploadID = s3ID
	}

	now := time.Now()
	video := &store.Video{
		ID:              videoID,
		ChannelID:       channel.ID,
		ChannelName:     channel.Name,
		ChannelHandle:   channel.Handle,
		ChannelAvatar:   channel.AvatarURL,
		Title:           payload.Title,
		Description:     payload.Description,
		VideoType:       videoType,
		Visibility:      "private",
		Status:          "CREATED",
		DurationSeconds: 0,
		Tags:            []string{"upload"},
		Category:        "General",
		CreatedAt:       now,
	}
	h.Store.Videos[videoID] = video
	h.Store.PersistVideo(video)
	channel.VideoCount++

	h.Store.UploadSessions[uploadID] = &store.UploadSession{
		ID:            uploadID,
		VideoID:       videoID,
		UserID:        user.ID,
		R2Key:         r2Key,
		Mode:          mode,
		S3UploadID:    s3UploadID,
		TotalParts:    totalParts,
		PartSize:      partSize,
		FileSizeBytes: payload.FileSizeBytes,
		MimeType:      mimeType,
		Chunks:        make(map[int][]byte),
		CreatedAt:     now,
	}

	identity.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"uploadId":   uploadID,
		"videoId":    videoID,
		"partSize":   partSize,
		"totalParts": totalParts,
		"key":        r2Key,
		"mode":       mode,
	})
}

func (h *Handler) GetPartUploadURL(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}
	var payload struct {
		PartNumber int `json:"partNumber"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.PartNumber <= 0 || payload.PartNumber > session.TotalParts {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid partNumber")
		return
	}

	if session.Mode == "s3" {
		s3cfg := S3ConfigFromEnv()
		signed, err := s3cfg.PresignPartUpload(session.R2Key, session.S3UploadID, payload.PartNumber, partURLTTL, time.Now())
		if err != nil {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Could not sign part URL")
			return
		}
		identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"uploadUrl":  signed,
			"partNumber": payload.PartNumber,
		})
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if forwarded := r.Header.Get("X-Forwarded-Proto"); forwarded != "" {
		scheme = forwarded
	}
	relayURL := fmt.Sprintf("%s://%s/v1/uploads/%s/chunks/%d?partNumber=%d", scheme, r.Host, session.ID, payload.PartNumber, payload.PartNumber)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"uploadUrl":  relayURL,
		"partNumber": payload.PartNumber,
	})
}

// PutChunk receives one chunk's raw bytes in local relay mode.
func (h *Handler) PutChunk(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}
	if session.Mode != "local" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_STATE", "Chunk relay is only available for local-mode sessions; use the presigned URL")
		return
	}
	partNumber, err := strconv.Atoi(chi.URLParam(r, "partNumber"))
	if err != nil || partNumber <= 0 || partNumber > session.TotalParts {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid part number")
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, session.PartSize+chunkSlackBytes))
	if err != nil || len(body) == 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Chunk body is required")
		return
	}

	session.Chunks[partNumber] = body
	sum := md5.Sum(body)
	etag := hex.EncodeToString(sum[:])
	w.Header().Set("ETag", `"`+etag+`"`)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"partNumber": partNumber,
		"etag":       etag,
		"size":       len(body),
	})
}

func localMediaDir() string {
	if dir := os.Getenv("GVP_LOCAL_MEDIA_DIR"); dir != "" {
		return dir
	}
	return "./var/media"
}

func (h *Handler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}
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

	video, ok := h.Store.Videos[session.VideoID]
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Upload video record is missing")
		return
	}

	if session.Mode == "s3" {
		s3cfg := S3ConfigFromEnv()
		s3Parts := make([]PartETag, 0, len(payload.Parts))
		for _, p := range payload.Parts {
			if p.PartNumber <= 0 || p.PartNumber > session.TotalParts || p.Etag == "" {
				identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Each part needs a valid partNumber and etag")
				return
			}
			s3Parts = append(s3Parts, PartETag{PartNumber: p.PartNumber, ETag: p.Etag})
		}
		if err := s3cfg.CompleteMultipartUpload(r.Context(), session.R2Key, session.S3UploadID, s3Parts); err != nil {
			identity.WriteError(w, r, http.StatusBadGateway, "R2_COMPLETE_FAILED", "Object storage rejected multipart completion")
			return
		}
	} else {
		missing := []int{}
		for part := 1; part <= session.TotalParts; part++ {
			if _, present := session.Chunks[part]; !present {
				missing = append(missing, part)
			}
		}
		if len(missing) > 0 {
			identity.WriteError(w, r, http.StatusBadRequest, "INCOMPLETE_UPLOAD", fmt.Sprintf("Missing chunk bytes for parts %v", missing))
			return
		}
		dir := filepath.Join(localMediaDir(), session.VideoID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			identity.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not stage upload")
			return
		}
		stored := filepath.Join(dir, "source.mp4")
		out, err := os.Create(stored)
		if err != nil {
			identity.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not stage upload")
			return
		}
		var assembled int64
		for part := 1; part <= session.TotalParts; part++ {
			n, err := out.Write(session.Chunks[part])
			assembled += int64(n)
			if err != nil {
				out.Close()
				identity.WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not assemble upload")
				return
			}
		}
		out.Close()
		session.StoredPath = stored
		session.Chunks = make(map[int][]byte)
		_ = assembled
	}

	now := time.Now()
	session.Completed = true
	session.CompletedAt = &now

	video.Status = "QUEUED"
	video.PublishedAt = now
	video.Visibility = "private" // Private draft until the media worker completes transcoding

	inputURI := "r2://" + session.R2Key
	if session.Mode == "local" {
		inputURI = "file://" + session.StoredPath
	}
	job := &store.MediaJob{
		ID:           "job-" + uuid.NewString(),
		VideoID:      video.ID,
		InputURI:     inputURI,
		OutputPrefix: "r2://public-hls/" + video.ID + "/",
		Status:       "QUEUED",
		ProgressPct:  0,
		CreatedAt:    now,
	}
	h.Store.MediaJobs[job.ID] = job

	// Dispatch asynchronously so the completion acknowledgement stays fast;
	// the worker folds READY/FAILED back into the job and video records.
	go h.dispatchTranscodeJob(job, video, session.R2Key)

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"uploadId": session.ID,
		"videoId":  video.ID,
		"status":   "QUEUED",
		"jobId":    job.ID,
		"message":  "Upload complete and verified. Transcoding job queued for media worker.",
	})
}

func (h *Handler) GetUploadStatus(w http.ResponseWriter, r *http.Request) {
	session, ok := h.loadSession(w, r)
	if !ok {
		return
	}

	status := "CREATED"
	progress := 0
	jobID := ""
	if session.Completed {
		status = "QUEUED"
	}
	for _, job := range h.Store.MediaJobs {
		if job.VideoID == session.VideoID {
			jobID = job.ID
			status = job.Status
			progress = job.ProgressPct
			break
		}
	}

	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"upload_id":    session.ID,
		"job_id":       jobID,
		"video_id":     session.VideoID,
		"status":       status,
		"progress_pct": progress,
	})
}
