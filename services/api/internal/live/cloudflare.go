package live

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// StreamInput is a provisioned Cloudflare Stream live input.
type StreamInput struct {
	UID       string
	RtmpsURL  string
	RtmpsKey  string
	SrtURL    string
	HlsURL    string
	WebRTCURL string
}

// StreamClient provisions Cloudflare Stream live inputs (P1-15). BaseURL is
// overridable for tests; production uses the Cloudflare API endpoint.
type StreamClient struct {
	AccountID  string
	Token      string
	BaseURL    string
	HTTPClient *http.Client
}

// StreamClientFromEnv builds the client from operator credentials. It
// returns nil when live ingest is not provisioned (honest idle mode).
func StreamClientFromEnv() *StreamClient {
	accountID := strings.TrimSpace(os.Getenv("CF_STREAM_ACCOUNT_ID"))
	token := strings.TrimSpace(os.Getenv("CF_STREAM_API_TOKEN"))
	if accountID == "" || token == "" {
		return nil
	}
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("CF_STREAM_BASE_URL")), "/")
	if base == "" {
		base = "https://api.cloudflare.com/client/v4"
	}
	return &StreamClient{AccountID: accountID, Token: token, BaseURL: base}
}

func (c *StreamClient) client() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// CreateLiveInput provisions a live input and returns ingest/playback URLs.
func (c *StreamClient) CreateLiveInput(ctx context.Context, name string) (*StreamInput, error) {
	payload, _ := json.Marshal(map[string]interface{}{
		"meta":                     map[string]string{"name": name},
		"recording":                map[string]interface{}{"mode": "automatic"},
		"deleteRecordingAfterDays": 30,
	})
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.BaseURL+"/accounts/"+c.AccountID+"/stream/live_inputs", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("stream live input request failed: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stream live input returned HTTP %d", resp.StatusCode)
	}
	var decoded struct {
		Success bool `json:"success"`
		Result  struct {
			UID   string `json:"uid"`
			Rtmps struct {
				URL       string `json:"url"`
				StreamKey string `json:"streamKey"`
			} `json:"rtmps"`
			Srt struct {
				URL        string `json:"url"`
				StreamID   string `json:"streamId"`
				Passphrase string `json:"passphrase"`
			} `json:"srt"`
			Playback struct {
				HLS  string `json:"hls"`
				DASH string `json:"dash"`
			} `json:"playback"`
			WebRTC struct {
				URL string `json:"url"`
			} `json:"webRTC"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil || !decoded.Success {
		return nil, fmt.Errorf("stream live input response invalid")
	}
	return &StreamInput{
		UID:       decoded.Result.UID,
		RtmpsURL:  decoded.Result.Rtmps.URL,
		RtmpsKey:  decoded.Result.Rtmps.StreamKey,
		SrtURL:    decoded.Result.Srt.URL,
		HlsURL:    decoded.Result.Playback.HLS,
		WebRTCURL: decoded.Result.WebRTC.URL,
	}, nil
}

// DeleteLiveInput releases a provisioned live input (best effort cleanup).
func (c *StreamClient) DeleteLiveInput(ctx context.Context, uid string) error {
	if uid == "" {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		c.BaseURL+"/accounts/"+c.AccountID+"/stream/live_inputs/"+uid, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.client().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("stream live input delete returned HTTP %d", resp.StatusCode)
	}
	return nil
}
