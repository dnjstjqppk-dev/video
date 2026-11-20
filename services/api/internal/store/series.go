package store

import "sort"

// EpisodesForSeries returns the videos attached to a series ordered by
// season then episode number. Deleted videos are skipped.
func (s *MemoryStore) EpisodesForSeries(seriesID string) []*Video {
	var out []*Video
	for _, v := range s.Videos {
		if v == nil || v.Status == "DELETED" || v.SeriesID != seriesID {
			continue
		}
		out = append(out, v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].SeasonNo != out[j].SeasonNo {
			return out[i].SeasonNo < out[j].SeasonNo
		}
		if out[i].EpisodeNo != out[j].EpisodeNo {
			return out[i].EpisodeNo < out[j].EpisodeNo
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// SeriesHasApprovedAIEpisode reports whether the series holds at least one
// approved AI episode (deleted videos excluded). AiOriginal mirrors it.
func (s *MemoryStore) SeriesHasApprovedAIEpisode(seriesID string) bool {
	for _, v := range s.EpisodesForSeries(seriesID) {
		if v != nil && v.AiGenerated && v.AiReviewStatus == "approved" {
			return true
		}
	}
	return false
}

// RefreshSeriesOriginalFlag re-derives AiOriginal from the current episodes
// and persists the series when the flag changes. Call it after every
// transition that can add or remove a qualifying episode: attach (both the
// old and the new series on moves), detach, review decisions, AI unflagging,
// and deletion.
func (s *MemoryStore) RefreshSeriesOriginalFlag(seriesID string) {
	if seriesID == "" {
		return
	}
	srs, ok := s.Series[seriesID]
	if !ok || srs == nil {
		return
	}
	if want := s.SeriesHasApprovedAIEpisode(seriesID); srs.AiOriginal != want {
		srs.AiOriginal = want
		s.PersistSeries(srs)
	}
}

// ToSeriesEpisode projects a video onto its series-episode list entry.
func ToSeriesEpisode(v *Video) SeriesEpisode {
	thumb := v.ThumbnailURL
	if v.CustomThumbnailURL != "" {
		thumb = v.CustomThumbnailURL
	}
	return SeriesEpisode{
		VideoID:      v.ID,
		Title:        v.Title,
		ThumbnailURL: thumb,
		DurationSec:  v.DurationSeconds,
		SeasonNo:     v.SeasonNo,
		EpisodeNo:    v.EpisodeNo,
		ViewCount:    v.ViewCount,
		Visibility:   v.Visibility,
		Status:       v.Status,
		AiGenerated:  v.AiGenerated,
		AiTool:       v.AiTool,
	}
}

// SeriesContext describes where a video sits inside its series, including
// the next episode for binge autoplay. It returns nil for standalone
// videos. Visibility gating stays with the caller.
func (s *MemoryStore) SeriesContext(video *Video) map[string]interface{} {
	if video == nil || video.SeriesID == "" {
		return nil
	}
	series, ok := s.Series[video.SeriesID]
	if !ok || series == nil {
		return nil
	}
	episodes := s.EpisodesForSeries(series.ID)
	ctx := map[string]interface{}{
		"id":            series.ID,
		"title":         series.Title,
		"status":        series.Status,
		"ai_original":   series.AiOriginal,
		"episode_total": len(episodes),
		"season_no":     video.SeasonNo,
		"episode_no":    video.EpisodeNo,
	}
	for i, ep := range episodes {
		if ep.ID != video.ID {
			continue
		}
		ctx["episode_index"] = i + 1
		if i+1 < len(episodes) {
			ctx["next_episode"] = ToSeriesEpisode(episodes[i+1])
		}
		break
	}
	return ctx
}
