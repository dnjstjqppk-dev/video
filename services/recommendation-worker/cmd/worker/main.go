package main

import (
	"context"
	"log/slog"
	"math"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"
)

type VideoCandidate struct {
	ID                  string
	ChannelID           string
	Category            string
	Status              string // READY, PROCESSING, DELETED
	Visibility          string // public, private, unlisted
	AgeRestricted       bool
	Impressions         int64
	AvgWatchDurationSec float64
	PublishedAt         time.Time
	Score               float64
}

type UserProfile struct {
	UserID            string
	PreferredCategory string
	AgeVerified       bool
	BlockedChannels   map[string]bool
}

// FilterEligible removes non-public, non-ready, blocked, or age-restricted items.
func FilterEligible(candidates []VideoCandidate, user UserProfile) []VideoCandidate {
	var out []VideoCandidate
	for _, c := range candidates {
		if c.Status != "READY" || c.Visibility != "public" {
			continue
		}
		if c.AgeRestricted && !user.AgeVerified {
			continue
		}
		if user.BlockedChannels != nil && user.BlockedChannels[c.ChannelID] {
			continue
		}
		out = append(out, c)
	}
	return out
}

// RankCandidates applies multi-objective scoring (relevance, freshness, watch duration, cold-start boost, channel diversity).
func RankCandidates(candidates []VideoCandidate, user UserProfile, now time.Time) []VideoCandidate {
	eligible := FilterEligible(candidates, user)

	for i := range eligible {
		c := &eligible[i]

		// 1. Relevance score
		relevance := 1.0
		if strings.EqualFold(c.Category, user.PreferredCategory) {
			relevance = 2.0
		}

		// 2. Freshness decay (half-life ~ 48 hours)
		ageHours := math.Max(0, now.Sub(c.PublishedAt).Hours())
		freshness := math.Exp(-ageHours / 72.0)

		// 3. Watch time quality signal
		watchQuality := math.Log1p(math.Max(0, c.AvgWatchDurationSec)) / 6.0

		// 4. Cold-start exploration booster for newly published videos with < 200 impressions
		coldStartBoost := 0.0
		if c.Impressions < 200 && ageHours <= 24 {
			coldStartBoost = 0.75
		}

		c.Score = (relevance * 0.40) + (freshness * 0.25) + (watchQuality * 0.25) + coldStartBoost
	}

	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].Score > eligible[j].Score
	})

	// 5. Channel diversity penalty: dampen consecutive appearances from the same channel
	seenChannelCount := make(map[string]int)
	for i := range eligible {
		ch := eligible[i].ChannelID
		count := seenChannelCount[ch]
		if count > 0 {
			eligible[i].Score *= math.Pow(0.75, float64(count))
		}
		seenChannelCount[ch]++
	}

	sort.Slice(eligible, func(i, j int) bool {
		return eligible[i].Score > eligible[j].Score
	})

	return eligible
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	slog.Info("Starting Recommendation Ranking & Cold-Start Worker")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
	case <-time.After(100 * time.Millisecond):
	}
}
