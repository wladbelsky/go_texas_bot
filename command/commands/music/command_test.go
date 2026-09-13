package music

import "testing"

func TestIsURL(t *testing.T) {
	cases := map[string]bool{
		"https://youtube.com/watch?v=abc": true,
		"http://example.com":              true,
		"never gonna give you up":         false,
		"ytsearch:some song":              false,
		"":                                false,
	}
	for input, want := range cases {
		if got := isURL(input); got != want {
			t.Errorf("isURL(%q) = %v, want %v", input, got, want)
		}
	}
}
