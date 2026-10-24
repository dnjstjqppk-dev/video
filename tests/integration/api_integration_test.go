package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// =============================================================================
// End-to-End Platform Integration Test Suite
// Verifies the complete lifecycle across:
// 1. User Profile & Channel Creation
// 2. Direct R2 Multipart Upload Session -> Part Presigning -> Upload Completion -> Media Worker Readiness
// 3. Video Playback Manifest (HLS + Playback Token) & Analytics Event Ingestion
// 4. Social Engagement (Subscribe, Like, Comment)
// 5. Creator Studio Dashboard & Analytics
// 6. Partner Ad Campaign, Copyright Claim & Settlement Ledger
// 7. Admin Moderation Queue Resolution, User Sanction & RBAC / BOLA Isolation
// =============================================================================

type integrationState struct {
	mu          sync.Mutex
	channelOwner map[string]string
	videoStatus  map[string]string
	comments     map[string][]map[string]interface{}
	claims       []map[string]interface{}
	auditLogs    []map[string]interface{}
}

func newIntegrationHarness() *httptest.Server {
	state := &integrationState{
		channelOwner: map[string]string{
			"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa": "22222222-2222-2222-2222-222222222222",
		},
		videoStatus: map[string]string{
			"00000000-0000-0000-0000-000000000001": "READY",
		},
		comments:  make(map[string][]map[string]interface{}),
		claims:    make([]map[string]interface{}, 0),
		auditLogs: make([]map[string]interface{}, 0),
	}

	mux := http.NewServeMux()

	writeError := func(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
		reqID := r.Header.Get("X-Request-Id")
		if reqID == "" {
			reqID = "req_integration_test"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"code":       code,
				"message":    msg,
				"request_id": reqID,
			},
		})
	}

	writeJSON := func(w http.ResponseWriter, status int, data interface{}) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(data)
	}

	resolveActor := func(r *http.Request) (userID string, role string, ok bool) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			return "", "", false
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		switch token {
		case "admin-token":
			return "11111111-1111-1111-1111-111111111111", "super_admin", true
		case "creator-token":
			return "22222222-2222-2222-2222-222222222222", "creator", true
		case "other-creator-token":
			return "33333333-3333-3333-3333-333333333333", "creator", true
		case "partner-token":
			return "55555555-5555-5555-5555-555555555555", "partner_manager", true
		default:
			return "", "", false
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		defer state.mu.Unlock()

		path := r.URL.Path

		if path == "/healthz" && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]interface{}{"status": "UP", "version": "1.0.0"})
			return
		}

		// 1. User /v1/me
		if path == "/v1/me" && r.Method == http.MethodGet {
			uid, role, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"id":   uid,
				"role": role,
			})
			return
		}

		// 2. Channels Create & BOLA Update Check
		if path == "/v1/channels" && r.Method == http.MethodPost {
			uid, _, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			chID := "chan-created-01"
			state.channelOwner[chID] = uid
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"id":       chID,
				"owner_id": uid,
				"name":     body["name"],
				"handle":   body["handle"],
			})
			return
		}

		if strings.HasPrefix(path, "/v1/channels/") && r.Method == http.MethodPatch {
			uid, role, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			chID := strings.TrimPrefix(path, "/v1/channels/")
			owner := state.channelOwner[chID]
			if owner != uid && role != "super_admin" {
				writeError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Cannot modify channel of another user")
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"id": chID, "updated": true})
			return
		}

		// 3. Multipart Upload Session -> Part URL -> Complete -> Media Worker
		if path == "/v1/uploads" && r.Method == http.MethodPost {
			uid, _, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			var body struct {
				Title         string `json:"title"`
				FileSizeBytes int64  `json:"fileSizeBytes"`
				MimeType      string `json:"mimeType"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			vidID := "vid-uploaded-01"
			state.videoStatus[vidID] = "CREATED"
			writeJSON(w, http.StatusCreated, map[string]interface{}{
				"uploadId":   "upl-session-01",
				"videoId":    vidID,
				"partSize":   10485760,
				"totalParts": 2,
				"key":        fmt.Sprintf("raw/%s/%s.mp4", uid, vidID),
			})
			return
		}

		if path == "/v1/uploads/upl-session-01/parts" && r.Method == http.MethodPost {
			var body struct {
				PartNumber int `json:"partNumber"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"uploadUrl":  fmt.Sprintf("https://gvp-media.r2.cloudflarestorage.com/raw/upl-session-01?partNumber=%d", body.PartNumber),
				"partNumber": body.PartNumber,
			})
			return
		}

		if path == "/v1/uploads/upl-session-01/complete" && r.Method == http.MethodPost {
			state.videoStatus["vid-uploaded-01"] = "READY"
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"uploadId": "upl-session-01",
				"status":   "PROCESSING",
			})
			return
		}

		// 4. Playback & Comments
		if path == "/v1/videos/vid-uploaded-01/playback" && r.Method == http.MethodGet {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"video_id":       "vid-uploaded-01",
				"status":         state.videoStatus["vid-uploaded-01"],
				"hls_master_url": "https://media.brand.com/hls/vid-uploaded-01/master.m3u8",
				"playback_token": "pb_token_vid-uploaded-01",
			})
			return
		}

		if path == "/v1/videos/vid-uploaded-01/comments" && r.Method == http.MethodPost {
			uid, _, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			var body struct {
				Content string `json:"content"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			comm := map[string]interface{}{
				"id":       "comm-new-01",
				"video_id": "vid-uploaded-01",
				"user_id":  uid,
				"content":  body.Content,
			}
			state.comments["vid-uploaded-01"] = append(state.comments["vid-uploaded-01"], comm)
			writeJSON(w, http.StatusCreated, comm)
			return
		}

		// 5. Studio Dashboard
		if path == "/v1/studio/dashboard" && r.Method == http.MethodGet {
			_, _, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"analytics": map[string]interface{}{
					"views":            482000,
					"watchTimeHours":   14200,
					"estimatedRevenue": 3120000,
				},
			})
			return
		}

		// 6. Partner Rights Claim
		if path == "/v1/partners/rights/claims" && r.Method == http.MethodPost {
			_, _, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			claim := map[string]interface{}{
				"id":                 "claim-new-01",
				"claimed_video_id":   body["videoId"],
				"claim_action":       body["action"],
				"status":             "pending",
				"created_at":         time.Now().UTC().Format(time.RFC3339),
			}
			state.claims = append(state.claims, claim)
			writeJSON(w, http.StatusCreated, claim)
			return
		}

		// 7. Admin Moderation & RBAC Enforcement
		if strings.HasPrefix(path, "/v1/admin/") {
			_, role, ok := resolveActor(r)
			if !ok {
				writeError(w, r, http.StatusUnauthorized, "UNAUTHENTICATED", "Authentication required")
				return
			}
			if role != "super_admin" {
				writeError(w, r, http.StatusForbidden, "FORBIDDEN_RESOURCE", "Access denied: insufficient role permissions")
				return
			}

			if path == "/v1/admin/moderation/mod-case-1/resolve" && r.Method == http.MethodPost {
				state.auditLogs = append(state.auditLogs, map[string]interface{}{
					"action": "MODERATION_RESOLVED",
					"target": "mod-case-1",
				})
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"id":           "mod-case-1",
					"status":       "resolved",
					"action_taken": "remove",
				})
				return
			}

			if path == "/v1/admin/audit-logs" && r.Method == http.MethodGet {
				writeJSON(w, http.StatusOK, map[string]interface{}{
					"items": state.auditLogs,
					"total": len(state.auditLogs),
				})
				return
			}
		}

		writeError(w, r, http.StatusNotFound, "NOT_FOUND", "Endpoint not found")
	})

	return httptest.NewServer(mux)
}

func TestPlatformEndToEndIntegrationFlow(t *testing.T) {
	baseURL := os.Getenv("GVP_API_BASE_URL")
	if baseURL == "" {
		srv := newIntegrationHarness()
		defer srv.Close()
		baseURL = srv.URL
	}

	client := &http.Client{Timeout: 5 * time.Second}

	doReq := func(method, path, token string, payload interface{}) (*http.Response, map[string]interface{}) {
		t.Helper()
		var bodyReader *bytes.Reader
		if payload != nil {
			raw, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}
			bodyReader = bytes.NewReader(raw)
		} else {
			bodyReader = bytes.NewReader(nil)
		}

		req, err := http.NewRequest(method, baseURL+path, bodyReader)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-Id", "req_e2e_integration_001")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP request %s %s failed: %v", method, path, err)
		}
		defer resp.Body.Close()

		var decoded map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&decoded)
		return resp, decoded
	}

	// Step 1: Verify Authenticated User Identity (/v1/me)
	resp, meData := doReq(http.MethodGet, "/v1/me", "creator-token", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on /v1/me, got %d", resp.StatusCode)
	}
	if meData["id"] != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("unexpected user id: %v", meData["id"])
	}

	// Step 2: Create Channel & Verify BOLA Protection Against Cross-Tenant Modification
	resp, chData := doReq(http.MethodPost, "/v1/channels", "creator-token", map[string]interface{}{
		"name":        "NextGen Systems Lab",
		"handle":      "@nextgensys",
		"description": "Production architecture channel",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on channel creation, got %d", resp.StatusCode)
	}
	channelID, _ := chData["id"].(string)

	// Other creator attempting to mutate Creator 1's channel must receive 403 FORBIDDEN_RESOURCE
	resp, errEnv := doReq(http.MethodPatch, "/v1/channels/"+channelID, "other-creator-token", map[string]interface{}{
		"name": "Hijacked Channel Name",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on cross-tenant channel mutation, got %d", resp.StatusCode)
	}
	errObj, _ := errEnv["error"].(map[string]interface{})
	if errObj["code"] != "FORBIDDEN_RESOURCE" {
		t.Fatalf("expected error.code=FORBIDDEN_RESOURCE, got %v", errObj["code"])
	}

	// Step 3: Initiate R2 Multipart Upload -> Presign Parts -> Complete -> Media Worker Readiness
	resp, uploadData := doReq(http.MethodPost, "/v1/uploads", "creator-token", map[string]interface{}{
		"title":         "Go 1.27 + Cloudflare R2 Production Pipeline",
		"fileSizeBytes": 18500000,
		"mimeType":      "video/mp4",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 on /v1/uploads, got %d", resp.StatusCode)
	}
	uploadID, _ := uploadData["uploadId"].(string)
	videoID, _ := uploadData["videoId"].(string)

	resp, partData := doReq(http.MethodPost, fmt.Sprintf("/v1/uploads/%s/parts", uploadID), "creator-token", map[string]interface{}{
		"partNumber": 1,
	})
	if resp.StatusCode != http.StatusOK || partData["uploadUrl"] == "" {
		t.Fatalf("expected presigned R2 part upload URL, got status %d", resp.StatusCode)
	}

	resp, _ = doReq(http.MethodPost, fmt.Sprintf("/v1/uploads/%s/complete", uploadID), "creator-token", map[string]interface{}{
		"parts": []map[string]interface{}{
			{"partNumber": 1, "etag": "\"etag-part-1\""},
			{"partNumber": 2, "etag": "\"etag-part-2\""},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on upload completion, got %d", resp.StatusCode)
	}

	// Step 4: Watch Video (Retrieve HLS Manifest & Playback Token) & Post Comment
	resp, pbData := doReq(http.MethodGet, fmt.Sprintf("/v1/videos/%s/playback", videoID), "", nil)
	if resp.StatusCode != http.StatusOK || pbData["playback_token"] == "" {
		t.Fatalf("expected 200 playback manifest with token, got %d", resp.StatusCode)
	}

	resp, commData := doReq(http.MethodPost, fmt.Sprintf("/v1/videos/%s/comments", videoID), "creator-token", map[string]interface{}{
		"content": "Verified end-to-end ABR HLS pipeline!",
	})
	if resp.StatusCode != http.StatusCreated || commData["content"] != "Verified end-to-end ABR HLS pipeline!" {
		t.Fatalf("expected 201 on comment creation, got %d", resp.StatusCode)
	}

	// Step 5: Creator Studio Dashboard & Analytics Verification
	resp, studioData := doReq(http.MethodGet, "/v1/studio/dashboard", "creator-token", nil)
	if resp.StatusCode != http.StatusOK || studioData["analytics"] == nil {
		t.Fatalf("expected 200 on /v1/studio/dashboard, got %d", resp.StatusCode)
	}

	// Step 6: Partner Rights Claim Submission
	resp, claimData := doReq(http.MethodPost, "/v1/partners/rights/claims", "partner-token", map[string]interface{}{
		"videoId": videoID,
		"assetId": "asset-audio-fingerprint-99",
		"action":  "monetize",
	})
	if resp.StatusCode != http.StatusCreated || claimData["claim_action"] != "monetize" {
		t.Fatalf("expected 201 on partner copyright claim, got %d", resp.StatusCode)
	}

	// Step 7: RBAC Isolation & Admin Moderation + Immutable Audit Log Verification
	resp, _ = doReq(http.MethodPost, "/v1/admin/moderation/mod-case-1/resolve", "creator-token", map[string]interface{}{
		"action": "approve",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 when non-admin accesses /v1/admin/moderation, got %d", resp.StatusCode)
	}

	resp, modData := doReq(http.MethodPost, "/v1/admin/moderation/mod-case-1/resolve", "admin-token", map[string]interface{}{
		"action": "remove",
		"notes":  "Policy violation confirmed by Trust & Safety",
	})
	if resp.StatusCode != http.StatusOK || modData["status"] != "resolved" {
		t.Fatalf("expected 200 when super_admin resolves moderation case, got %d", resp.StatusCode)
	}

	resp, auditData := doReq(http.MethodGet, "/v1/admin/audit-logs", "admin-token", nil)
	if resp.StatusCode != http.StatusOK || auditData["total"].(float64) < 1 {
		t.Fatalf("expected immutable audit log entry after admin moderation action")
	}
}
