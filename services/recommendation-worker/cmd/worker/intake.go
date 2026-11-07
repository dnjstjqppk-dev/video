package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// intake.go exposes the recommendation execution API: candidate ranking with
// eligibility filtering, freshness decay, cold-start boost and channel
// diversity. Feed builders POST here behind RECOMMENDATION_WORKER_URL.

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

type rankCandidateRequest struct {
	ID                  string  `json:"id"`
	ChannelID           string  `json:"channelId"`
	Category            string  `json:"category"`
	Status              string  `json:"status"`
	Visibility          string  `json:"visibility"`
	AgeRestricted       bool    `json:"ageRestricted"`
	Impressions         int64   `json:"impressions"`
	AvgWatchDurationSec float64 `json:"avgWatchDurationSec"`
	PublishedAt         string  `json:"publishedAt"`
}

type rankUserRequest struct {
	UserID            string   `json:"userId"`
	PreferredCategory string   `json:"preferredCategory"`
	AgeVerified       bool     `json:"ageVerified"`
	BlockedChannels   []string `json:"blockedChannels"`
}

type rankRequest struct {
	Candidates []rankCandidateRequest `json:"candidates"`
	User       rankUserRequest        `json:"user"`
	Limit      int                    `json:"limit"`
}

func (h *healthServer) handleRank(w http.ResponseWriter, r *http.Request) {
	var req rankRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid request body")
		return
	}
	if len(req.Candidates) == 0 {
		writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "candidates must not be empty")
		return
	}
	candidates := make([]VideoCandidate, 0, len(req.Candidates))
	for _, c := range req.Candidates {
		publishedAt := time.Now().UTC()
		if c.PublishedAt != "" {
			parsed, err := time.Parse(time.RFC3339, c.PublishedAt)
			if err != nil {
				writeIntakeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "publishedAt must be RFC3339")
				return
			}
			publishedAt = parsed
		}
		candidates = append(candidates, VideoCandidate{
			ID:                  c.ID,
			ChannelID:           c.ChannelID,
			Category:            c.Category,
			Status:              c.Status,
			Visibility:          c.Visibility,
			AgeRestricted:       c.AgeRestricted,
			Impressions:         c.Impressions,
			AvgWatchDurationSec: c.AvgWatchDurationSec,
			PublishedAt:         publishedAt,
		})
	}
	blocked := make(map[string]bool, len(req.User.BlockedChannels))
	for _, ch := range req.User.BlockedChannels {
		blocked[ch] = true
	}
	ranked := RankCandidates(candidates, UserProfile{
		UserID:            req.User.UserID,
		PreferredCategory: req.User.PreferredCategory,
		AgeVerified:       req.User.AgeVerified,
		BlockedChannels:   blocked,
	}, time.Now().UTC())
	if req.Limit > 0 && len(ranked) > req.Limit {
		ranked = ranked[:req.Limit]
	}
	out := make([]map[string]interface{}, 0, len(ranked))
	for _, c := range ranked {
		out = append(out, map[string]interface{}{
			"id":        c.ID,
			"channelId": c.ChannelID,
			"score":     c.Score,
		})
	}
	writeIntakeJSON(w, http.StatusOK, map[string]interface{}{"ranked": out})
}

func (h *healthServer) registerIntakeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /rank", h.handleRank)
}
