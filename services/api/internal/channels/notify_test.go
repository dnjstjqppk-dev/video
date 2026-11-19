package channels_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestSubscriptionNotify_PreferenceLifecycle(t *testing.T) {
	srv := newChannelsTestServer()
	defer srv.Close()
	h := srv.Config.Handler
	channelID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	get := func(token string) (int, map[string]interface{}) {
		req, rec := channelsReq(http.MethodGet, "/channels/"+channelID+"/subscription", token, nil)
		h.ServeHTTP(rec, req)
		var out map[string]interface{}
		_ = json.Unmarshal(rec.Body.Bytes(), &out)
		return rec.Code, out
	}

	if code, _ := get(""); code != http.StatusUnauthorized {
		t.Fatalf("anonymous subscription read must be 401, got %d", code)
	}
	if code, out := get("viewer-token"); code != http.StatusOK || out["subscribed"] != false || out["notify_level"] != "" {
		t.Fatalf("unsubscribed state wrong: %d %v", code, out)
	}

	req, rec := channelsReq(http.MethodPost, "/channels/"+channelID+"/subscribe", "viewer-token", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("subscribe must be 200, got %d", rec.Code)
	}
	var subbed map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &subbed)
	if subbed["notify_level"] != "all" {
		t.Fatalf("new subscription must default to all, got %v", subbed)
	}

	req, rec = channelsReq(http.MethodPatch, "/channels/"+channelID+"/subscribe", "viewer-token",
		map[string]string{"notify_level": "everything"})
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid level must be 400, got %d", rec.Code)
	}
	req, rec = channelsReq(http.MethodPatch, "/channels/"+channelID+"/subscribe", "viewer-token",
		map[string]string{"notify_level": "none"})
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("notify update must be 200, got %d", rec.Code)
	}
	if code, out := get("viewer-token"); code != http.StatusOK || out["notify_level"] != "none" {
		t.Fatalf("stored level must be none: %d %v", code, out)
	}

	req, rec = channelsReq(http.MethodDelete, "/channels/"+channelID+"/subscribe", "viewer-token", nil)
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("unsubscribe must be 200, got %d", rec.Code)
	}
	if code, out := get("viewer-token"); code != http.StatusOK || out["subscribed"] != false || out["notify_level"] != "" {
		t.Fatalf("unsubscribed state must clear level: %d %v", code, out)
	}

	req, rec = channelsReq(http.MethodPatch, "/channels/"+channelID+"/subscribe", "viewer-token",
		map[string]string{"notify_level": "all"})
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("notify change while unsubscribed must be 409, got %d", rec.Code)
	}
}
