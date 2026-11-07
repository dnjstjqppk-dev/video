package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// intake.go exposes the analytics execution API: telemetry ingest with bot
// filtering and deduplication plus an explicit ClickHouse batch flush. Event
// producers POST here behind ANALYTICS_WORKER_URL.

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

type ingestEventRequest struct {
	EventID      string  `json:"eventId"`
	EventType    string  `json:"eventType"`
	VideoID      string  `json:"videoId"`
	UserID       string  `json:"userId"`
	UserAgent    string  `json:"userAgent"`
	WatchSeconds float64 `json:"watchSeconds"`
	PlaybackRate float64 `json:"playbackRate"`
	OccurredAt   string  `json:"occurredAt"`
}

func (h *healthServer) handleIngest(w http.ResponseWriter, r *http.Request) {
	var req ingestEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	occurredAt := time.Now().UTC()
	if req.OccurredAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.OccurredAt)
		if err != nil {
			writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "occurredAt must be RFC3339")
			return
		}
		occurredAt = parsed
	}
	accepted, err := h.proc.Ingest(TelemetryEvent{
		EventID:      req.EventID,
		EventType:    req.EventType,
		VideoID:      req.VideoID,
		UserID:       req.UserID,
		UserAgent:    req.UserAgent,
		WatchSeconds: req.WatchSeconds,
		PlaybackRate: req.PlaybackRate,
		OccurredAt:   occurredAt,
	})
	if err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_EVENT", err.Error())
		return
	}
	writeIntakeJSON(w, http.StatusAccepted, map[string]interface{}{
		"eventId":  req.EventID,
		"accepted": accepted,
	})
}

func (h *healthServer) handleFlush(w http.ResponseWriter, r *http.Request) {
	flushed := h.proc.FlushToClickHouse()
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"flushed": flushed})
}

func (h *healthServer) registerIntakeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /ingest", h.handleIngest)
	mux.HandleFunc("POST /flush", h.handleFlush)
}
