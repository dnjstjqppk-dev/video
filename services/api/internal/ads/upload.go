package ads

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/security"
	"globalvideoplatform/services/api/internal/store"
)

const (
	maxVideoCreativeBytes = 100 << 20
	maxImageCreativeBytes = 10 << 20
	maxUploadBodyBytes    = 110 << 20
	maxMultipartMemory    = 32 << 20
)

var (
	videoFormats = map[string]bool{"preroll": true, "midroll": true, "shorts_sponsored": true}
	imageFormats = map[string]bool{"feed_banner": true, "banner": true}

	videoExtensions = map[string]bool{".mp4": true, ".webm": true, ".mov": true}
	imageExtensions = map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
)

// creativeDir resolves the local creative-asset root. Production points this
// at a shared volume (or migrates keys to R2); the stored layout already uses
// R2-shaped "<campaignID>/<creativeID><ext>" keys.
func (h *Handler) creativeDir() string {
	if h.CreativeDir != "" {
		return h.CreativeDir
	}
	if dir := os.Getenv("GVP_CREATIVE_DIR"); dir != "" {
		return dir
	}
	return filepath.Join("var", "creatives")
}

func creativeFilePath(dir, campaignID, creativeID, ext string) string {
	return filepath.Join(dir, campaignID, creativeID+ext)
}

// UploadCreative stores a real creative asset for a campaign (P0-5 ARTICLE 7.2
// steps 3-4). The record enters pending review; serving stays gated until approval.
func (h *Handler) UploadCreative(w http.ResponseWriter, r *http.Request) {
	user := identity.GetUserFromContext(r.Context())
	if user == nil {
		identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
		return
	}
	campID := chi.URLParam(r, "id")

	h.mu.Lock()
	camp, ok := h.Store.Campaigns[campID]
	h.mu.Unlock()
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Campaign not found")
		return
	}
	if !campaignVisible(camp, user) {
		identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify another advertiser's campaign")
		return
	}
	wantVideo, ok := videoFormats[camp.Format]
	if !ok {
		if _, ok := imageFormats[camp.Format]; !ok {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Campaign format cannot accept creative uploads")
			return
		}
		wantVideo = false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBodyBytes)
	if err := r.ParseMultipartForm(maxMultipartMemory); err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Valid multipart file upload is required")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Form field 'file' is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if wantVideo && !videoExtensions[ext] {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_FORMAT", "Video creatives require mp4, webm, or mov")
		return
	}
	if !wantVideo && !imageExtensions[ext] {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_FORMAT", "Banner creatives require jpg, png, gif, or webp")
		return
	}
	limit := int64(maxVideoCreativeBytes)
	if !wantVideo {
		limit = int64(maxImageCreativeBytes)
	}
	if header.Size > limit {
		identity.WriteError(w, r, http.StatusBadRequest, "FILE_TOO_LARGE", "Creative file exceeds the size limit")
		return
	}
	content, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Could not read uploaded file")
		return
	}
	if int64(len(content)) > limit {
		identity.WriteError(w, r, http.StatusBadRequest, "FILE_TOO_LARGE", "Creative file exceeds the size limit")
		return
	}
	// MIME sniffing must agree with the declared extension category.
	sniffed := http.DetectContentType(content)
	isVideoBytes := strings.HasPrefix(sniffed, "video/")
	isImageBytes := strings.HasPrefix(sniffed, "image/")
	if wantVideo != isVideoBytes || (!wantVideo && !isImageBytes) {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_FORMAT", "File content does not match its type")
		return
	}

	destination := strings.TrimSpace(r.FormValue("destinationUrl"))
	if destination != "" {
		parsed, err := url.Parse(destination)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.User != nil {
			identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "destinationUrl must be an http(s) URL")
			return
		}
	}

	creativeID := "creative-" + uuid.NewString()
	sum := sha256.Sum256(content)
	storedName := creativeID + ext
	campaignDir := filepath.Join(h.creativeDir(), campID)
	if err := os.MkdirAll(campaignDir, 0o755); err != nil {
		identity.WriteError(w, r, http.StatusInternalServerError, "UNAVAILABLE", "Creative storage is unavailable")
		return
	}
	if err := os.WriteFile(filepath.Join(campaignDir, storedName), content, 0o600); err != nil {
		identity.WriteError(w, r, http.StatusInternalServerError, "UNAVAILABLE", "Creative storage is unavailable")
		return
	}

	h.mu.Lock()
	creative := &store.AdCreative{
		ID:                creativeID,
		CampaignID:        campID,
		Format:            camp.Format,
		MediaURL:          "/v1/ads/creatives/" + creativeID + "/asset",
		ThumbnailURL:      "/v1/ads/creatives/" + creativeID + "/asset",
		Headline:          security.SanitizeUserText(r.FormValue("headline"), 160),
		Description:       fmt.Sprintf("sha256:%s", hex.EncodeToString(sum[:])),
		CallToAction:      security.SanitizeUserText(r.FormValue("callToAction"), 64),
		DestinationURL:    destination,
		DurationSeconds:   0,
		SkipOffsetSeconds: 5,
		ReviewStatus:      "pending",
	}
	h.Store.AdCreatives[creativeID] = creative
	camp.Creatives = append(camp.Creatives, creativeID)
	h.creativeFiles[creativeID] = creativeFilePath(h.creativeDir(), campID, creativeID, ext)
	h.creativeMIME[creativeID] = sniffed
	h.mu.Unlock()

	identity.WriteJSON(w, http.StatusCreated, creative)
}

// ServeCreativeAsset streams a stored creative file. Approved creatives are
// public (players embed them); pending/rejected ones need owner/admin access.
func (h *Handler) ServeCreativeAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.mu.Lock()
	creative, ok := h.Store.AdCreatives[id]
	path := h.creativeFiles[id]
	mime := h.creativeMIME[id]
	h.mu.Unlock()
	if !ok {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative asset not found")
		return
	}
	if creative.ReviewStatus != "approved" {
		user := identity.GetUserFromContext(r.Context())
		if user == nil {
			identity.WriteError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
			return
		}
		h.mu.Lock()
		camp, campOK := h.Store.Campaigns[creative.CampaignID]
		h.mu.Unlock()
		if !campOK || !campaignVisible(camp, user) {
			identity.WriteError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot access another advertiser's creative")
			return
		}
	}
	if path == "" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative file is not stored locally")
		return
	}
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Creative file is not stored locally")
		return
	}
	if mime == "" {
		mime = http.DetectContentType(content)
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
