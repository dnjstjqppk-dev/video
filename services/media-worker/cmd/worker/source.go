package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Source kinds ResolveSourceRef can return.
type SourceKind int

const (
	SourceLocal SourceKind = iota
	SourceRemote
)

// ResolveSourceRef classifies a transcode source reference without fetching.
// Accepted forms:
//
//   - file:///absolute/path — only under MEDIA_SOURCE_ROOTS allowlist
//   - https://... — presigned object-storage GET, downloaded to scratch
//
// Bare storage keys are rejected: the dispatcher owns credentials and must
// presign. file:// outside the allowlist is rejected (path traversal guard).
func ResolveSourceRef(ref string) (SourceKind, string, error) {
	if strings.HasPrefix(ref, "file://") {
		raw := strings.TrimPrefix(ref, "file://")
		if !filepath.IsAbs(raw) {
			// Windows file URIs arrive as file:///C:/... which is absolute after trim.
			if len(raw) < 3 || raw[1] != ':' {
				return 0, "", errors.New("file URI must be absolute")
			}
			// Normalize file:///C:/x to C:/x.
			raw = strings.TrimPrefix(raw, "/")
		}
		clean := filepath.Clean(raw)
		allowed := false
		for _, root := range mediaSourceRoots() {
			rel, err := filepath.Rel(root, clean)
			if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return 0, "", fmt.Errorf("file source outside MEDIA_SOURCE_ROOTS: %s", clean)
		}
		return SourceLocal, clean, nil
	}
	if strings.HasPrefix(ref, "https://") || strings.HasPrefix(ref, "http://") {
		parsed, err := url.ParseRequestURI(ref)
		if err != nil || parsed.Host == "" {
			return 0, "", fmt.Errorf("invalid source URL: %w", err)
		}
		// Remote sources arrive as dispatcher-presigned URLs on this internal
		// endpoint; the file:// allowlist below stays strict regardless.
		return SourceRemote, ref, nil
	}
	if strings.Contains(ref, "://") {
		return 0, "", fmt.Errorf("unsupported source URL scheme in %q", ref)
	}
	return 0, "", errors.New("bare storage keys are not fetchable; dispatcher must presign a download URL")
}

func mediaSourceRoots() []string {
	var roots []string
	if v := os.Getenv("MEDIA_SOURCE_ROOTS"); v != "" {
		for _, part := range strings.Split(v, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				roots = append(roots, filepath.Clean(trimmed))
			}
		}
	}
	roots = append(roots, filepath.Clean(os.TempDir()))
	if local := os.Getenv("GVP_LOCAL_MEDIA_DIR"); local != "" {
		roots = append(roots, filepath.Clean(local))
	}
	if cwd, err := os.Getwd(); err == nil {
		roots = append(roots, filepath.Join(cwd, "var", "media"))
	}
	return roots
}

// FetchRemoteSource downloads a presigned source URL into scratch space.
func FetchRemoteSource(ctx context.Context, sourceURL, destPath string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("source download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("source download rejected with HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()
	const maxSourceBytes = 5 << 30
	n, err := io.Copy(out, io.LimitReader(resp.Body, maxSourceBytes+1))
	if err != nil {
		return fmt.Errorf("source download interrupted: %w", err)
	}
	if n > maxSourceBytes {
		_ = os.Remove(destPath)
		return errors.New("source exceeds the 5 GiB worker limit")
	}
	return nil
}
