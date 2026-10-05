package help

import (
	"sort"
	"strings"
	"testing"
)

func TestCommandLines_SortedAndContainsKnownCommand(t *testing.T) {
	lines := commandLines()
	if len(lines) == 0 {
		t.Fatal("expected at least one command line")
	}
	if !sort.StringsAreSorted(lines) {
		t.Fatal("expected commandLines() to be sorted")
	}
	found := false
	for _, l := range lines {
		if strings.Contains(l, "**/say**") {
			found = true
		}
	}
	if !found {
		t.Fatal("expected commandLines() to include the /say command")
	}
}

func TestCommandFields_RespectsFieldLengthLimit(t *testing.T) {
	fields := commandFields()
	if len(fields) == 0 {
		t.Fatal("expected at least one field")
	}
	for _, f := range fields {
		if len(f.Value) > 1024 {
			t.Fatalf("field %q value is %d chars, exceeds Discord's 1024 limit", f.Name, len(f.Value))
		}
	}
}

func TestCommandFields_NoDataLost(t *testing.T) {
	fields := commandFields()
	var totalLines int
	for _, f := range fields {
		totalLines += strings.Count(f.Value, "\n") + 1
	}
	if totalLines != len(commandLines()) {
		t.Fatalf("commandFields() produced %d lines total, want %d", totalLines, len(commandLines()))
	}
}
