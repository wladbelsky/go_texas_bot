package embedutil

import (
	"strings"
	"testing"
)

func TestChunkLines_Empty(t *testing.T) {
	if got := ChunkLines(nil, 10); len(got) != 0 {
		t.Fatalf("ChunkLines(nil) = %q, want no chunks", got)
	}
}

func TestChunkLines_FitsInOne(t *testing.T) {
	got := ChunkLines([]string{"a", "b", "c"}, 10)
	if len(got) != 1 || got[0] != "a\nb\nc" {
		t.Fatalf("ChunkLines() = %q, want one chunk \"a\\nb\\nc\"", got)
	}
}

func TestChunkLines_SplitsAtLimit(t *testing.T) {
	got := ChunkLines([]string{"aaaa", "bbbb", "cccc"}, 9)
	want := []string{"aaaa\nbbbb", "cccc"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("ChunkLines() = %q, want %q", got, want)
	}
	for _, c := range got {
		if len(c) > 9 {
			t.Fatalf("chunk %q is %d bytes, over the 9-byte limit", c, len(c))
		}
	}
}

func TestChunkLines_OversizedLineGetsOwnChunk(t *testing.T) {
	got := ChunkLines([]string{"a", "0123456789", "b"}, 5)
	want := []string{"a", "0123456789", "b"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("ChunkLines() = %q, want %q", got, want)
	}
}
