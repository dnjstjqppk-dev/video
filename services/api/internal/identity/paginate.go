package identity

import (
	"net/http"
	"strconv"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

// ParsePagination extracts a clamped 1-based page/limit pair from query params.
func ParsePagination(r *http.Request, defaultLimit int) (page, limit int) {
	page = 1
	limit = defaultLimit
	if limit <= 0 {
		limit = DefaultPageLimit
	}
	if raw := r.URL.Query().Get("page"); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil && p > 0 {
			page = p
		}
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if l, err := strconv.Atoi(raw); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	return page, limit
}

// PaginateSlice returns the requested window plus the total length.
func PaginateSlice[T any](items []T, page, limit int) ([]T, int) {
	total := len(items)
	if total == 0 || limit <= 0 || page <= 0 {
		if items == nil {
			return []T{}, total
		}
		return items, total
	}
	start := (page - 1) * limit
	if start >= total {
		return []T{}, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return items[start:end], total
}

// PageEnvelope builds the standard list response with paging metadata.
func PageEnvelope(items interface{}, total, page, limit int) map[string]interface{} {
	envelope := map[string]interface{}{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	}
	if page*limit < total {
		envelope["nextCursor"] = "page:" + strconv.Itoa(page+1)
	}
	return envelope
}
