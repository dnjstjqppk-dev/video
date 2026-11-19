package store_test

import (
	"testing"
	"time"

	"globalvideoplatform/services/api/internal/store"
)

func TestHourlySeries_ZeroFillsAndSums(t *testing.T) {
	s := store.NewMemoryStore()
	now := time.Now().UTC().Truncate(time.Hour)
	s.RecordHourlyView("ch1", now.Add(-2*time.Hour), 5)
	s.RecordHourlyView("ch1", now, 3)
	s.RecordHourlyView("ch1", now, 2)
	s.RecordHourlyView("ch2", now, 99)

	series := s.HourlySeries("ch1", 48)
	if len(series) != 48 {
		t.Fatalf("expected 48 buckets, got %d", len(series))
	}
	var total int64
	for i, p := range series {
		total += p.Views
		if i > 0 && !p.Hour.After(series[i-1].Hour) {
			t.Fatal("series must be oldest-first ascending")
		}
	}
	if total != 10 {
		t.Fatalf("expected total 10, got %d", total)
	}
	if series[47].Views != 5 {
		t.Fatalf("expected current hour 5, got %d", series[47].Views)
	}
	if series[45].Views != 5 {
		t.Fatalf("expected hour-2 bucket 5, got %d", series[45].Views)
	}
	if got := s.HourlySeries("missing", 48); len(got) != 48 || got[0].Views != 0 {
		t.Fatal("unknown channel must yield zero-filled series")
	}
}

func TestNoteCountedView_BucketsCurrentHour(t *testing.T) {
	s := store.NewMemoryStore()
	v := &store.Video{ID: "v1", ChannelID: "ch9"}
	s.NoteCountedView(v)
	s.NoteCountedView(nil)
	s.NoteCountedView(&store.Video{ID: "v2"})
	series := s.HourlySeries("ch9", 48)
	if series[47].Views != 1 {
		t.Fatalf("expected 1 counted view in current hour, got %d", series[47].Views)
	}
}
