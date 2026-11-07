package security

import (
	"strings"
	"testing"
)

func TestSanitizeUserText(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", "안녕하세요 반갑습니다", "안녕하세요 반갑습니다"},
		{"script", `<script>alert(1)</script>hello`, "alert(1)hello"},
		{"img handler", `<img src=x onerror=alert(1)>pic`, "pic"},
		{"anchor", `click <a href="https://evil.example">here</a> now`, "click here now"},
		{"style", `<style>body{display:none}</style>text`, "body{display:none}text"},
		{"comparison survives", "1 < 2 and 5 > 3", "1 < 2 and 5 > 3"},
		{"heart emoticon survives", "love <3 you", "love <3 you"},
		{"unclosed bracket survives", "a < b", "a < b"},
		{"nul removed", "ab\x00cd", "abcd"},
		{"trimmed", "  hi  ", "hi"},
		{"blank lines collapsed", "a\n\n\n\nb", "a\n\nb"},
		{"only tags becomes empty", "<div><br/></div>", ""},
		{"svg handler", `<svg onload=alert(1)>x</svg>`, "x"},
		{"mixed case tag", `<ScRiPt>alert(1)</sCrIpT>ok`, "alert(1)ok"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SanitizeUserText(tc.input, 2000); got != tc.want {
				t.Fatalf("SanitizeUserText(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSanitizeUserTextTruncates(t *testing.T) {
	got := SanitizeUserText(strings.Repeat("가", 3000), 2000)
	if len([]rune(got)) != 2000 {
		t.Fatalf("expected 2000 runes, got %d", len([]rune(got)))
	}
}
