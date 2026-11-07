package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// intake.go exposes the live-controller execution API on the worker HTTP
// server: session provisioning, state transitions, stream-key rotation and
// health evaluation. The API owns its inline live implementation; this
// service is the scale-out executor behind LIVE_WORKER_URL.

func writeIntakeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeIntakeError(w http.ResponseWriter, status int, code, message string) {
	writeIntakeJSON(w, status, map[string]interface{}{
		"error": map[string]string{"code": code, "message": message},
	})
}

func newSessionID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "sess_" + hex.EncodeToString(b)
}

type provisionSessionRequest struct {
	SessionID  string `json:"sessionId"`
	ChannelID  string `json:"channelId"`
	DvrEnabled bool   `json:"dvrEnabled"`
}

type transitionRequest struct {
	State string `json:"state"`
}

type healthEvalRequest struct {
	BitrateKbps int     `json:"bitrateKbps"`
	FPS         int     `json:"fps"`
	DroppedPct  float64 `json:"droppedPct"`
}

func liveSessionResponse(s *LiveSession) map[string]interface{} {
	return map[string]interface{}{
		"sessionId":      s.ID,
		"channelId":      s.ChannelID,
		"rtmpsUrl":       s.RtmpsURL,
		"srtUrl":         s.SrtURL,
		"playbackHlsUrl": s.PlaybackHLSURL,
		"state":          string(s.State),
		"dvrEnabled":     s.DvrEnabled,
		"vodVideoId":     s.VodVideoID,
		"createdAt":      s.CreatedAt.Format(time.RFC3339),
	}
}

func (h *healthServer) handleProvisionSession(w http.ResponseWriter, r *http.Request) {
	var req provisionSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	if req.ChannelID == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "channelId is required")
		return
	}
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = newSessionID()
	}
	sess := h.ctrl.ProvisionInput(sessionID, req.ChannelID, req.DvrEnabled)
	writeIntakeJSON(w, http.StatusCreated, liveSessionResponse(sess))
}

func (h *healthServer) handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.ctrl.mu.RLock()
	sess, ok := h.ctrl.Sessions[id]
	h.ctrl.mu.RUnlock()
	if !ok {
		writeIntakeError(w, http.StatusNotFound, "NOT_FOUND", "Live session not found")
		return
	}
	writeIntakeJSON(w, http.StatusOK, liveSessionResponse(sess))
}

func (h *healthServer) handleTransitionSession(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req transitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	target := LiveState(strings.ToUpper(req.State))
	if target != LiveStateLive && target != LiveStateEnded {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "state must be LIVE or ENDED")
		return
	}
	if err := h.ctrl.TransitionState(id, target); err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeIntakeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeIntakeError(w, http.StatusConflict, "INVALID_TRANSITION", err.Error())
		return
	}
	h.ctrl.mu.RLock()
	sess := h.ctrl.Sessions[id]
	h.ctrl.mu.RUnlock()
	writeIntakeJSON(w, http.StatusOK, liveSessionResponse(sess))
}

func (h *healthServer) handleRotateStreamKey(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	key, err := h.ctrl.RotateStreamKey(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			writeIntakeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		writeIntakeError(w, http.StatusConflict, "INVALID_STATE", err.Error())
		return
	}
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"sessionId": id, "streamKey": key})
}

func (h *healthServer) handleEvaluateHealth(w http.ResponseWriter, r *http.Request) {
	var req healthEvalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	health := EvaluateHealth(req.BitrateKbps, req.FPS, req.DroppedPct)
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{
		"bitrateKbps":     health.BitrateKbps,
		"fps":             health.FPS,
		"droppedFramePct": health.DroppedFramePct,
		"status":          health.Status,
	})
}

func (h *healthServer) registerIntakeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /sessions", h.handleProvisionSession)
	mux.HandleFunc("GET /sessions/{id}", h.handleGetSession)
	mux.HandleFunc("POST /sessions/{id}/transition", h.handleTransitionSession)
	mux.HandleFunc("POST /sessions/{id}/rotate-key", h.handleRotateStreamKey)
	mux.HandleFunc("POST /sessions/health", h.handleEvaluateHealth)
}
