package comments

import (
	"os"
	"strings"
	"unicode"
)

// blockedWords parses COMMENT_BLOCKLIST (comma-separated, case-insensitive).
// Empty by default; operators set it per locale.
func blockedWords() []string {
	raw := os.Getenv("COMMENT_BLOCKLIST")
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	for _, w := range strings.Split(raw, ",") {
		if trimmed := strings.ToLower(strings.TrimSpace(w)); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// BlockedTerm returns the first blocklisted whole word in content, if any.
func BlockedTerm(content string) (string, bool) {
	blocked := blockedWords()
	if len(blocked) == 0 {
		return "", false
	}
	set := make(map[string]bool, len(blocked))
	for _, w := range blocked {
		set[w] = true
	}
	for _, field := range strings.FieldsFunc(content, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r))
	}) {
		if lower := strings.ToLower(field); set[lower] {
			return lower, true
		}
	}
	return "", false
}
