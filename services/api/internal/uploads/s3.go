// Package uploads implements resumable multipart video ingest.
//
// Production mode talks real S3 multipart APIs (Cloudflare R2 / MinIO) using
// SigV4 request signing implemented with the standard library only.
// Local mode relays chunk bytes through the API itself so development, CI and
// E2E environments ingest real bytes without any object storage.
package uploads

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

// S3Config describes the object-storage target for multipart uploads.
type S3Config struct {
	Endpoint       string // server-side base URL, e.g. http://r2-storage:9000
	PublicEndpoint string // browser-facing base URL for presigned PUTs
	Bucket         string
	AccessKey      string
	SecretKey      string
	Region         string
}

// S3ConfigFromEnv loads object-storage settings. Production deployments must
// set R2_ACCESS_KEY_ID / R2_SECRET_ACCESS_KEY; local development falls back
// to the docker-compose MinIO defaults.
func S3ConfigFromEnv() S3Config {
	cfg := S3Config{
		Endpoint:       firstNonEmpty(os.Getenv("R2_ENDPOINT"), "http://localhost:9002"),
		PublicEndpoint: firstNonEmpty(os.Getenv("R2_PUBLIC_ENDPOINT"), os.Getenv("R2_ENDPOINT"), "http://localhost:9002"),
		Bucket:         firstNonEmpty(os.Getenv("R2_BUCKET_RAW"), "gvp-raw-uploads"),
		AccessKey:      os.Getenv("R2_ACCESS_KEY_ID"),
		SecretKey:      os.Getenv("R2_SECRET_ACCESS_KEY"),
		Region:         firstNonEmpty(os.Getenv("R2_REGION"), "auto"),
	}
	if cfg.AccessKey == "" && cfg.SecretKey == "" {
		// Local MinIO development defaults (see infra/docker/docker-compose.yml).
		cfg.AccessKey = "gvp_r2_local_key"
		cfg.SecretKey = "gvp_r2_local_secret"
	}
	return cfg
}

// Configured reports whether production S3 mode is explicitly enabled.
func (c S3Config) Configured() bool {
	return os.Getenv("R2_ACCESS_KEY_ID") != "" && os.Getenv("R2_SECRET_ACCESS_KEY") != ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// PartETag pairs a completed part number with the ETag returned by storage.
type PartETag struct {
	PartNumber int
	ETag       string
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func signingKey(secret, date, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(date))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

// encodeS3Path percent-encodes each path segment while preserving slashes.
func encodeS3Path(path string) string {
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		segments[i] = encodeS3Segment(seg)
	}
	return strings.Join(segments, "/")
}

func encodeS3Segment(seg string) string {
	var out strings.Builder
	for i := 0; i < len(seg); i++ {
		c := seg[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			out.WriteByte(c)
		} else {
			fmt.Fprintf(&out, "%%%02X", c)
		}
	}
	return out.String()
}

func encodeQueryComponent(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

func canonicalQueryString(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, encodeQueryComponent(k)+"="+encodeQueryComponent(params[k]))
	}
	return strings.Join(parts, "&")
}

// PresignPartUpload builds a SigV4 presigned PUT URL for one multipart part.
// The URL is browser-facing (PublicEndpoint) so clients upload bytes directly.
func (c S3Config) PresignPartUpload(key, s3UploadID string, partNumber int, expires time.Duration, now time.Time) (string, error) {
	if c.AccessKey == "" || c.SecretKey == "" {
		return "", fmt.Errorf("object storage credentials are not configured")
	}
	if partNumber <= 0 || s3UploadID == "" || key == "" {
		return "", fmt.Errorf("key, upload ID and a positive part number are required")
	}
	base := strings.TrimSuffix(c.PublicEndpoint, "/")
	publicURL, err := url.Parse(base)
	if err != nil || publicURL.Scheme == "" || publicURL.Host == "" {
		return "", fmt.Errorf("invalid public object-storage endpoint")
	}

	now = now.UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, c.Region)

	canonicalURI := encodeS3Path("/" + c.Bucket + "/" + key)
	params := map[string]string{
		"partNumber":          fmt.Sprintf("%d", partNumber),
		"uploadId":            s3UploadID,
		"X-Amz-Algorithm":     "AWS4-HMAC-SHA256",
		"X-Amz-Credential":    c.AccessKey + "/" + credentialScope,
		"X-Amz-Date":          amzDate,
		"X-Amz-Expires":       fmt.Sprintf("%d", int64(expires/time.Second)),
		"X-Amz-SignedHeaders": "host",
	}
	canonicalRequest := strings.Join([]string{
		"PUT",
		canonicalURI,
		canonicalQueryString(params),
		"host:" + publicURL.Host,
		"",
		"host",
		"UNSIGNED-PAYLOAD",
	}, "\n")

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(c.SecretKey, dateStamp, c.Region, "s3"), []byte(stringToSign)))

	params["X-Amz-Signature"] = signature
	return fmt.Sprintf("%s%s?%s", base, canonicalURI, canonicalQueryString(params)), nil
}

// signHeaderRequest attaches SigV4 Authorization headers for server-side calls.
func (c S3Config) signHeaderRequest(req *http.Request, payloadHash string, now time.Time) {
	now = now.UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, c.Region)

	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	signedHeaders := "host;x-amz-content-sha256;x-amz-date"
	canonicalHeaders := "host:" + req.URL.Host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"

	rawQuery := req.URL.Query()
	flat := make(map[string]string, len(rawQuery))
	for k, values := range rawQuery {
		if len(values) > 0 {
			flat[k] = values[0]
		} else {
			flat[k] = ""
		}
	}
	canonicalRequest := strings.Join([]string{
		req.Method,
		encodeS3Path(req.URL.Path),
		canonicalQueryString(flat),
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(signingKey(c.SecretKey, dateStamp, c.Region, "s3"), []byte(stringToSign)))
	req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.AccessKey, credentialScope, signedHeaders, signature))
}

func (c S3Config) doSigned(ctx context.Context, method, rawURL string, body []byte, contentType string) (*http.Response, []byte, error) {
	payloadHash := sha256Hex(body)
	req, err := http.NewRequestWithContext(ctx, method, rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c.signHeaderRequest(req, payloadHash, time.Now())
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return resp, nil, err
	}
	return resp, respBody, nil
}

// EnsureBucket creates the raw-upload bucket when missing (idempotent).
func (c S3Config) EnsureBucket(ctx context.Context) error {
	target := strings.TrimSuffix(c.Endpoint, "/") + "/" + c.Bucket
	resp, body, err := c.doSigned(ctx, http.MethodPut, target, nil, "")
	if err != nil {
		return fmt.Errorf("ensure bucket request failed: %w", err)
	}
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusConflict {
		return nil
	}
	if resp.StatusCode == http.StatusBadRequest && bytes.Contains(body, []byte("BucketAlreadyOwnedByYou")) {
		return nil
	}
	return fmt.Errorf("ensure bucket returned HTTP %d", resp.StatusCode)
}

type initiateMultipartResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	UploadID string   `xml:"UploadId"`
}

// CreateMultipartUpload opens a server-side multipart upload and returns the S3 upload ID.
func (c S3Config) CreateMultipartUpload(ctx context.Context, key, contentType string) (string, error) {
	target := strings.TrimSuffix(c.Endpoint, "/") + "/" + c.Bucket + "/" + encodeS3Path(key) + "?uploads"
	resp, body, err := c.doSigned(ctx, http.MethodPost, target, nil, contentType)
	if err != nil {
		return "", fmt.Errorf("create multipart upload failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("create multipart upload returned HTTP %d", resp.StatusCode)
	}
	var result initiateMultipartResult
	if err := xml.Unmarshal(body, &result); err != nil || result.UploadID == "" {
		return "", fmt.Errorf("invalid create multipart upload response")
	}
	return result.UploadID, nil
}

// CompleteMultipartUpload finalizes the S3 multipart object from reported part ETags.
func (c S3Config) CompleteMultipartUpload(ctx context.Context, key, s3UploadID string, parts []PartETag) error {
	var buf strings.Builder
	buf.WriteString("<CompleteMultipartUpload>")
	for _, p := range parts {
		etag := strings.Trim(p.ETag, "\"")
		fmt.Fprintf(&buf, "<Part><PartNumber>%d</PartNumber><ETag>\"%s\"</ETag></Part>", p.PartNumber, etag)
	}
	buf.WriteString("</CompleteMultipartUpload>")
	target := strings.TrimSuffix(c.Endpoint, "/") + "/" + c.Bucket + "/" + encodeS3Path(key) +
		"?uploadId=" + encodeQueryComponent(s3UploadID)
	resp, _, err := c.doSigned(ctx, http.MethodPost, target, []byte(buf.String()), "application/xml")
	if err != nil {
		return fmt.Errorf("complete multipart upload failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("complete multipart upload returned HTTP %d", resp.StatusCode)
	}
	return nil
}
