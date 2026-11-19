package videos

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/store"
	"globalvideoplatform/services/api/internal/uploads"
)

const maxThumbnailBytes = 5 << 20

// framePosterKey is the worker-emitted poster published for every READY video.
const framePosterKey = "poster.jpg"

func (h *Handler) loadOwnedVideo(w http.ResponseWriter, r *http.Request) (*store.Video, bool) {
	id := chi.URLParam(r, "id")
	video, ok := h.Store.Videos[id]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return nil, false
	}
	if !h.canManageVideo(r, video) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify video of another user")
		return nil, false
	}
	return video, true
}

// UploadThumbnail stores an owner-supplied custom thumbnail in the public
// HLS bucket and makes it the video's active thumbnail.
func (h *Handler) UploadThumbnail(w http.ResponseWriter, r *http.Request) {
	video, ok := h.loadOwnedVideo(w, r)
	if !ok {
		return
	}
	s3cfg := uploads.S3ConfigFromEnv()
	if !s3cfg.Configured() {
		identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Object storage is unreachable")
		return
	}
	if err := r.ParseMultipartForm(maxThumbnailBytes + (1 << 20)); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Thumbnail multipart body is required")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Form field 'file' is required")
		return
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maxThumbnailBytes+1))
	if err != nil || len(body) == 0 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Thumbnail bytes are required")
		return
	}
	if len(body) > maxThumbnailBytes {
		identity.WriteError(w, r, http.StatusBadRequest, "QUOTA_EXCEEDED", "Thumbnail exceeds the 5 MiB limit")
		return
	}
	contentType := http.DetectContentType(body)
	var ext string
	switch contentType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	case "image/webp":
		ext = "webp"
	default:
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Thumbnail must be a JPEG, PNG, or WebP image")
		return
	}
	cfg, _, err := image.DecodeConfig(strings.NewReader(string(body)))
	if err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Thumbnail bytes are not a decodable image")
		return
	}
	if cfg.Width < 320 || cfg.Height < 180 {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Thumbnail must be at least 320x180")
		return
	}
	key := "hls/" + video.ID + "/custom." + ext
	ctx := r.Context()
	if err := s3cfg.EnsureBucketNamed(ctx, s3cfg.HLSBucket); err != nil {
		identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Object storage is unreachable")
		return
	}
	if err := s3cfg.PutObject(ctx, s3cfg.HLSBucket, key, contentType, body); err != nil {
		identity.WriteError(w, r, http.StatusBadGateway, "R2_UPLOAD_FAILED", "Object storage rejected the thumbnail")
		return
	}
	customURL := uploads.PublicBase(s3cfg) + "/" + key
	video.CustomThumbnailURL = customURL
	video.ThumbnailURL = customURL
	h.Store.PersistVideo(video)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"video_id":      video.ID,
		"thumbnail_url": customURL,
	})
}

// frameOption describes one selectable auto-extracted still.
type frameOption struct {
	Source string `json:"source"`
	URL    string `json:"url"`
}

// GetFrames lists the selectable thumbnails: worker poster, auto-extracted
// frames recorded at publish, and the owner-uploaded custom image.
func (h *Handler) GetFrames(w http.ResponseWriter, r *http.Request) {
	video, ok := h.loadOwnedVideo(w, r)
	if !ok {
		return
	}
	frames := []frameOption{}
	if video.Status == "READY" {
		s3cfg := uploads.S3ConfigFromEnv()
		if s3cfg.Configured() {
			frames = append(frames, frameOption{
				Source: "poster",
				URL:    uploads.PublicBase(s3cfg) + "/hls/" + video.ID + "/" + framePosterKey,
			})
		}
		for i, url := range video.AutoFrameURLs {
			if url == "" {
				continue
			}
			frames = append(frames, frameOption{Source: fmt.Sprintf("auto_%d", i+1), URL: url})
		}
	}
	if video.CustomThumbnailURL != "" {
		frames = append(frames, frameOption{Source: "custom", URL: video.CustomThumbnailURL})
	}
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"video_id": video.ID,
		"current":  video.ThumbnailURL,
		"frames":   frames,
	})
}

// SelectThumbnail switches the active thumbnail to one of the known frames.
func (h *Handler) SelectThumbnail(w http.ResponseWriter, r *http.Request) {
	video, ok := h.loadOwnedVideo(w, r)
	if !ok {
		return
	}
	var payload struct {
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Source == "" {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Field 'source' is required")
		return
	}
	var next string
	switch payload.Source {
	case "poster":
		if video.Status != "READY" {
			identity.WriteError(w, r, http.StatusConflict, "INVALID_STATE", "Poster is available after processing completes")
			return
		}
		s3cfg := uploads.S3ConfigFromEnv()
		if !s3cfg.Configured() {
			identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Object storage is unreachable")
			return
		}
		next = uploads.PublicBase(s3cfg) + "/hls/" + video.ID + "/" + framePosterKey
	case "custom":
		if video.CustomThumbnailURL == "" {
			identity.WriteError(w, r, http.StatusConflict, "INVALID_STATE", "No custom thumbnail uploaded")
			return
		}
		next = video.CustomThumbnailURL
	default:
		var idx int
		if _, err := fmt.Sscanf(payload.Source, "auto_%d", &idx); err != nil || idx < 1 || idx > len(video.AutoFrameURLs) {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unknown frame source")
			return
		}
		next = video.AutoFrameURLs[idx-1]
		if next == "" {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Unknown frame source")
			return
		}
	}
	video.ThumbnailURL = next
	h.Store.PersistVideo(video)
	identity.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"video_id":      video.ID,
		"thumbnail_url": next,
	})
}
