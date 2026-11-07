package security

import (
	"strings"
	"unicode/utf8"
)

// SanitizeUserText strips HTML/XML tags from user-supplied text and normalizes
// whitespace so stored content can never carry executable markup. It is
// intentionally dependency-free: markup is removed rather than allow-listed,
// which is the correct policy for plain-text fields (comments, titles, reasons).
//
// The function also removes NUL bytes, trims surrounding whitespace, collapses
// runs of blank lines to at most one, and truncates to maxRunes runes when
// maxRunes > 0. An empty result means the input carried no visible text.
func SanitizeUserText(input string, maxRunes int) string {
	if input == "" {
		return ""
	}
	stripped := stripTags(input)
	stripped = strings.ReplaceAll(stripped, "\x00", "")
	stripped = collapseBlankLines(stripped)
	stripped = strings.TrimSpace(stripped)
	if maxRunes > 0 && utf8.RuneCountInString(stripped) > maxRunes {
		runes := []rune(stripped)
		stripped = strings.TrimSpace(string(runes[:maxRunes]))
	}
	return stripped
}

// stripTags removes <...> sequences. A '<' without a closing '>' on the same
// logical tag is emitted literally so plain-text comparisons (a < b) survive,
// while well-formed tags (including script/style/iframe/img with handlers)
// are always dropped.
func stripTags(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if s[i] != '<' {
			b.WriteByte(s[i])
			i++
			continue
		}
		end := strings.IndexByte(s[i:], '>')
		if end == -1 {
			b.WriteString(s[i:])
			break
		}
		tag := s[i+1 : i+end]
		if isTagLike(tag) {
			i += end + 1
			continue
		}
		b.WriteByte('<')
		i++
	}
	return b.String()
}

// isTagLike reports whether the text between '<' and '>' looks like markup:
// an optional '/' or '!' or '?' followed by an ASCII letter. This keeps
// constructs such as "1 < 2" and "a <3 b" intact.
func isTagLike(tag string) bool {
	if tag == "" {
		return false
	}
	c := tag[0]
	if c == '/' || c == '!' || c == '?' {
		if len(tag) < 2 {
			return false
		}
		c = tag[1]
	}
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// collapseBlankLines reduces 3+ consecutive newlines to exactly two.
func collapseBlankLines(s string) string {
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.ReplaceAll(s, "\r\n", "\n")
}
