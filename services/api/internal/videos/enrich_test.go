package videos_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetVideo_EnrichesChannelStats(t *testing.T) {
	srv := newVideosTestServer()
	defer srv.Close()

	req, rec := videosReq(http.MethodGet, "/videos/00000000-0000-0000-0000-000000000001", "", nil)
	srv.Config.Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out struct {
		ChannelSubscriberCount int64 `json:"channel_subscriber_count"`
		ChannelVideoCount      int   `json:"channel_video_count"`
		ChannelIsVerified      bool  `json:"channel_is_verified"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode video: %v", err)
	}
	if out.ChannelSubscriberCount != 482000 || out.ChannelVideoCount != 142 || !out.ChannelIsVerified {
		t.Fatalf("channel stats not enriched: %+v", out)
	}
}
