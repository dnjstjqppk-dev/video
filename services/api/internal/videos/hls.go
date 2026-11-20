package videos

import (
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"globalvideoplatform/services/api/internal/identity"
	"globalvideoplatform/services/api/internal/uploads"
)

// HLSRoutes serves published HLS artifacts through the API so playback
// works behind tunnels without public buckets. Visibility rules mirror
// GetVideo; unknown videos and keys 404 without leaking existence.
func (h *Handler) HLSRoutes() http.Handler {
	r := chi.NewRouter()
	r.Get("/{videoID}/*", h.ServeHLS)
	return r
}

func hlsContentType(key string) string {
	switch strings.ToLower(path.Ext(key)) {
	case ".m3u8":
		return "application/vnd.apple.mpegurl"
	case ".ts":
		return "video/mp2t"
	case ".m4s":
		return "video/iso.segment"
	case ".mp4":
		return "video/mp4"
	case ".vtt":
		return "text/vtt"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func isPublicStill(rel string) bool {
	rel = path.Clean("/" + rel)[1:]
	if strings.HasPrefix(rel, "storyboard/") {
		return true
	}
	switch {
	case rel == "poster.jpg":
		return true
	case strings.HasPrefix(rel, "custom."):
		return true
	case strings.HasPrefix(rel, "auto_") && strings.HasSuffix(rel, ".jpg") && !strings.Contains(rel, "/"):
		return true
	default:
		return false
	}
}

func hlsCacheControl(key string) string {
	base := path.Base(key)
	if strings.HasSuffix(base, ".m3u8") {
		return "no-cache"
	}
	if strings.HasPrefix(base, "custom.") {
		return "public, max-age=60"
	}
	return "public, max-age=31536000, immutable"
}

func (h *Handler) ServeHLS(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "videoID")
	rel := strings.TrimPrefix(chi.URLParam(r, "*"), "/")
	if videoID == "" || rel == "" || strings.Contains(rel, "..") || strings.ContainsRune(rel, '\\') {
		identity.WriteError(w, r, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid artifact path")
		return
	}
	video, ok := h.Store.Videos[videoID]
	if !ok || video.Status == "DELETED" {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
		return
	}
	// Thumbnails and stills are capability URLs (unguessable video IDs):
	// <img> tags cannot send Authorization headers, so posters stay
	// reachable while playlists and segments keep full gating.
	// A valid ?token= playback token authorizes players that cannot send
	// Authorization headers (Shaka fetches segments directly).
	token := r.URL.Query().Get("token")
	authorized := token != "" && VerifyPlaybackToken(token, videoID) == nil
	if !isPublicStill(rel) && !authorized {
		user := identity.GetUserFromContext(r.Context())
		if ClassifyViewAccess(user, video, h.Store.Channels) == ViewHide {
			identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Video not found")
			return
		}
	}
	s3cfg := uploads.S3ConfigFromEnv()
	if !s3cfg.Configured() {
		identity.WriteError(w, r, http.StatusServiceUnavailable, "R2_UNAVAILABLE", "Object storage is unreachable")
		return
	}
	resp, err := s3cfg.GetObject(r.Context(), s3cfg.HLSBucket, "hls/"+videoID+"/"+rel)
	if err != nil {
		identity.WriteError(w, r, http.StatusNotFound, "NOT_FOUND", "Artifact not found")
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", hlsContentType(rel))
	w.Header().Set("Cache-Control", hlsCacheControl(rel))
	// Consoles embed stills (<img>) and fetch playlists/segments (Shaka)
	// cross-origin, so the gateway opts out of the global same-origin
	// CORP. Artifacts stay protected by unguessable IDs, visibility
	// gating, and signed playback tokens — not by origin isolation.
	w.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
	// Propagate the playback token into nested playlist URIs so Shaka can
	// follow variant playlists and segments without headers or cookies.
	if authorized && strings.HasSuffix(strings.ToLower(rel), ".m3u8") {
		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if err != nil {
			identity.WriteError(w, r, http.StatusBadGateway, "R2_UNAVAILABLE", "Object storage read failed")
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(rewritePlaylistToken(string(body), token))
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, resp.Body)
}

// rewritePlaylistToken appends ?token= to every relative URI in an HLS
// playlist (segment lines plus URI="..." attributes), keeping tags intact.
func rewritePlaylistToken(body, token string) []byte {
	var out strings.Builder
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		switch {
		case trimmed == "" || strings.HasPrefix(trimmed, "#EXT"):
			out.WriteString(rewriteURIAttributes(line, token))
		case strings.HasPrefix(trimmed, "#"):
			out.WriteString(line)
		default:
			out.WriteString(withToken(strings.TrimSpace(line), token))
		}
		out.WriteString("\n")
	}
	return []byte(out.String())
}

func withToken(uri, token string) string {
	if uri == "" || strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://") || strings.HasPrefix(uri, "/") || strings.HasPrefix(uri, "data:") {
		return uri
	}
	sep := "?"
	if strings.Contains(uri, "?") {
		sep = "&"
	}
	return uri + sep + "token=" + token
}

func rewriteURIAttributes(line, token string) string {
	const prefix = `URI="`
	var out strings.Builder
	rest := line
	for {
		idx := strings.Index(rest, prefix)
		if idx < 0 {
			out.WriteString(rest)
			return out.String()
		}
		out.WriteString(rest[:idx+len(prefix)])
		rest = rest[idx+len(prefix):]
		end := strings.Index(rest, `"`)
		if end < 0 {
			out.WriteString(rest)
			return out.String()
		}
		out.WriteString(withToken(rest[:end], token))
		rest = rest[end:]
	}
}
