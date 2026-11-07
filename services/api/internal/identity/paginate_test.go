package identity

import (
	"net/http/httptest"
	"testing"
)

func TestParsePagination(t *testing.T) {
	cases := []struct {
		name      string
		query     string
		def       int
		wantPage  int
		wantLimit int
	}{
		{"defaults", "", 20, 1, 20},
		{"explicit", "?page=3&limit=10", 20, 3, 10},
		{"clamps-limit", "?page=1&limit=500", 20, 1, MaxPageLimit},
		{"ignores-garbage", "?page=abc&limit=-5", 24, 1, 24},
		{"ignores-zero-page", "?page=0", 20, 1, 20},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/x"+tc.query, nil)
			page, limit := ParsePagination(req, tc.def)
			if page != tc.wantPage || limit != tc.wantLimit {
				t.Fatalf("got page=%d limit=%d, want page=%d limit=%d", page, limit, tc.wantPage, tc.wantLimit)
			}
		})
	}
}

func TestPaginateSlice(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}

	window, total := PaginateSlice(items, 1, 2)
	if total != 5 || len(window) != 2 || window[0] != 1 || window[1] != 2 {
		t.Fatalf("unexpected first window: %+v total=%d", window, total)
	}

	window, total = PaginateSlice(items, 3, 2)
	if total != 5 || len(window) != 1 || window[0] != 5 {
		t.Fatalf("unexpected last window: %+v total=%d", window, total)
	}

	window, total = PaginateSlice(items, 9, 2)
	if total != 5 || len(window) != 0 {
		t.Fatalf("out-of-range page must return an empty window: %+v", window)
	}

	window, total = PaginateSlice([]int(nil), 1, 20)
	if total != 0 || window == nil || len(window) != 0 {
		t.Fatalf("nil input must return an empty non-nil window")
	}
}

func TestPageEnvelope(t *testing.T) {
	first := PageEnvelope([]int{1, 2}, 5, 1, 2)
	if first["total"] != 5 || first["nextCursor"] != "page:2" {
		t.Fatalf("first page must carry a next cursor: %+v", first)
	}
	last := PageEnvelope([]int{5}, 5, 3, 2)
	if _, ok := last["nextCursor"]; ok {
		t.Fatalf("final page must not carry a next cursor: %+v", last)
	}
}
