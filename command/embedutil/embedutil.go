// Package embedutil holds helpers for fitting content into Discord embeds.
package embedutil

import "strings"

// MaxFieldValueLen stays a little under Discord's 1024-char embed field
// value limit. It's compared against byte length, which is never less than
// the character count Discord measures, so it's a safe upper bound.
const MaxFieldValueLen = 1000

// ChunkLines joins lines with newlines into as few chunks as possible, each
// at most maxLen bytes. A single line longer than maxLen gets a chunk of
// its own rather than being split.
func ChunkLines(lines []string, maxLen int) []string {
	var chunks []string
	var current strings.Builder
	for _, line := range lines {
		if current.Len() > 0 && current.Len()+1+len(line) > maxLen {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}
