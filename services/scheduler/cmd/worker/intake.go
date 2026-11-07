package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

// intake.go exposes the scheduler execution API: enqueue outbox events,
// scheduled videos, upload sessions and settlement periods, then run the
// periodic sweep via /tick (outbox relay + due-video publish + expired
// upload cleanup) or close a settlement period explicitly.

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

func newOutboxID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "obx_" + hex.EncodeToString(b)
}

// EnqueueOutboxEvent stages a transactional outbox event for the next relay.
func (s *SchedulerEngine) EnqueueOutboxEvent(topic, payloadJSON string) *OutboxEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	ev := &OutboxEvent{
		ID:          newOutboxID(),
		Topic:       topic,
		PayloadJSON: payloadJSON,
		CreatedAt:   time.Now().UTC(),
	}
	s.Outbox = append(s.Outbox, ev)
	return ev
}

// ScheduleVideo registers a video for future publication.
func (s *SchedulerEngine) ScheduleVideo(id string, publishAt time.Time) *ScheduledVideo {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := &ScheduledVideo{ID: id, Visibility: "scheduled", PublishAt: publishAt}
	s.Videos = append(s.Videos, v)
	return v
}

// TrackUpload registers a multipart upload session for expiry cleanup.
func (s *SchedulerEngine) TrackUpload(uploadID string, expiresAt time.Time) *UploadSession {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := &UploadSession{UploadID: uploadID, Status: "pending", ExpiresAt: expiresAt}
	s.Uploads = append(s.Uploads, u)
	return u
}

// OpenSettlementPeriod registers an open financial period.
func (s *SchedulerEngine) OpenSettlementPeriod(periodID string) *SettlementPeriod {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := &SettlementPeriod{PeriodID: periodID, Status: "open"}
	s.Settlements[periodID] = p
	return p
}

func parseIntakeTime(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339, raw)
}

func (h *healthServer) handleEnqueueOutbox(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Topic   string `json:"topic"`
		Payload string `json:"payload"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Topic == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "topic is required")
		return
	}
	ev := h.eng.EnqueueOutboxEvent(req.Topic, req.Payload)
	writeIntakeJSON(w, http.StatusCreated, map[string]interface{}{
		"id": ev.ID, "topic": ev.Topic, "published": ev.Published,
	})
}

func (h *healthServer) handleScheduleVideo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID        string `json:"id"`
		PublishAt string `json:"publishAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" || req.PublishAt == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "id and publishAt (RFC3339) are required")
		return
	}
	publishAt, err := parseIntakeTime(req.PublishAt)
	if err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "publishAt must be RFC3339")
		return
	}
	v := h.eng.ScheduleVideo(req.ID, publishAt)
	writeIntakeJSON(w, http.StatusCreated, map[string]interface{}{
		"id": v.ID, "visibility": v.Visibility, "publishAt": v.PublishAt.Format(time.RFC3339),
	})
}

func (h *healthServer) handleTrackUpload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UploadID  string `json:"uploadId"`
		ExpiresAt string `json:"expiresAt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UploadID == "" || req.ExpiresAt == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "uploadId and expiresAt (RFC3339) are required")
		return
	}
	expiresAt, err := parseIntakeTime(req.ExpiresAt)
	if err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "expiresAt must be RFC3339")
		return
	}
	u := h.eng.TrackUpload(req.UploadID, expiresAt)
	writeIntakeJSON(w, http.StatusCreated, map[string]interface{}{
		"uploadId": u.UploadID, "status": u.Status, "expiresAt": u.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *healthServer) handleOpenSettlement(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PeriodID string `json:"periodId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PeriodID == "" {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "periodId is required")
		return
	}
	p := h.eng.OpenSettlementPeriod(req.PeriodID)
	writeIntakeJSON(w, http.StatusCreated, map[string]interface{}{
		"periodId": p.PeriodID, "status": p.Status,
	})
}

func (h *healthServer) handleTick(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Now string `json:"now"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	now := time.Now().UTC()
	if req.Now != "" {
		parsed, err := parseIntakeTime(req.Now)
		if err != nil {
			writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "now must be RFC3339")
			return
		}
		now = parsed
	}
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{
		"relayedEvents":   h.eng.RelayOutboxEvents(),
		"publishedVideos": h.eng.PublishDueVideos(now),
		"expiredUploads":  h.eng.CleanupExpiredUploads(now),
		"tickAt":          now.Format(time.RFC3339),
	})
}

func (h *healthServer) handleCloseSettlement(w http.ResponseWriter, r *http.Request) {
	periodID := r.PathValue("id")
	if !h.eng.CloseSettlementPeriod(periodID) {
		writeIntakeError(w, http.StatusNotFound, "NOT_FOUND", "Open settlement period not found")
		return
	}
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"periodId": periodID, "status": "closed"})
}

func (h *healthServer) registerIntakeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /outbox", h.handleEnqueueOutbox)
	mux.HandleFunc("POST /scheduled-videos", h.handleScheduleVideo)
	mux.HandleFunc("POST /uploads", h.handleTrackUpload)
	mux.HandleFunc("POST /settlements", h.handleOpenSettlement)
	mux.HandleFunc("POST /settlements/{id}/close", h.handleCloseSettlement)
	mux.HandleFunc("POST /tick", h.handleTick)
}
