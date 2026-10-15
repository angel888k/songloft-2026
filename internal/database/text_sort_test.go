package database

import "testing"

func TestTextSortKeyPreservesNonChineseText(t *testing.T) {
	for _, tt := range []struct{ text, want string }{
		{"", ""}, {" \t\n", ""}, {"  夜曲 Live  ", "yequ live"},
		{"愛你", "aini"}, {"爱你", "aini"}, {"01. 夜曲", "01. yequ"},
		{"(Live) β音楽 🎵", "(live) βyinle 🎵"},
	} {
		if got := textSortKey(tt.text); got != tt.want {
			t.Errorf("textSortKey(%q) = %q, want %q", tt.text, got, tt.want)
		}
	}
}
