package music

import (
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func track(title string) lavalink.Track {
	return lavalink.Track{Info: lavalink.TrackInfo{Title: title}}
}

func titles(tracks []lavalink.Track) []string {
	out := make([]string, len(tracks))
	for i, t := range tracks {
		out[i] = t.Info.Title
	}
	return out
}

func TestQueue_EmptyByDefault(t *testing.T) {
	q := NewQueue()
	if !q.IsEmpty() {
		t.Fatal("expected a new queue to be empty")
	}
	if _, ok := q.Current(); ok {
		t.Fatal("expected Current() to report nothing on an empty queue")
	}
}

func TestQueue_AddAndAdvance(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"), track("c"))

	if q.IsEmpty() {
		t.Fatal("expected queue to be non-empty after Add")
	}
	if q.Length() != 3 {
		t.Fatalf("Length() = %d, want 3", q.Length())
	}

	first, ok := q.Advance()
	if !ok || first.Info.Title != "a" {
		t.Fatalf("Advance() = (%v, %v), want (\"a\", true)", first.Info.Title, ok)
	}
	second, ok := q.Advance()
	if !ok || second.Info.Title != "b" {
		t.Fatalf("Advance() = (%v, %v), want (\"b\", true)", second.Info.Title, ok)
	}
}

func TestQueue_AdvancePastEndWithoutRepeatStops(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"))

	if _, ok := q.Advance(); !ok {
		t.Fatal("expected first Advance() to succeed")
	}
	if _, ok := q.Advance(); ok {
		t.Fatal("expected second Advance() to report nothing left to play")
	}
}

func TestQueue_RepeatOne_StaysOnCurrentTrack(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"))
	q.SetRepeatMode(RepeatOne)

	first, _ := q.Advance()
	if first.Info.Title != "a" {
		t.Fatalf("first track = %q, want a", first.Info.Title)
	}
	again, ok := q.Advance()
	if !ok || again.Info.Title != "a" {
		t.Fatalf("Advance() with RepeatOne = (%q, %v), want (\"a\", true)", again.Info.Title, ok)
	}
}

func TestQueue_RepeatAll_WrapsAround(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"))
	q.SetRepeatMode(RepeatAll)

	q.Advance() // a
	q.Advance() // b
	wrapped, ok := q.Advance()
	if !ok || wrapped.Info.Title != "a" {
		t.Fatalf("Advance() past the end with RepeatAll = (%q, %v), want (\"a\", true)", wrapped.Info.Title, ok)
	}
}

func TestQueue_Previous(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"))
	q.Advance()
	q.Advance()

	prev, ok := q.Previous()
	if !ok || prev.Info.Title != "a" {
		t.Fatalf("Previous() = (%q, %v), want (\"a\", true)", prev.Info.Title, ok)
	}
	if _, ok := q.Previous(); ok {
		t.Fatal("expected Previous() to report nothing before the first track")
	}
}

func TestQueue_UpcomingAndHistory(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"), track("c"))
	q.Advance() // a

	if got := titles(q.History()); len(got) != 0 {
		t.Fatalf("History() = %v, want empty before advancing past the first track", got)
	}
	if got := titles(q.Upcoming()); len(got) != 2 || got[0] != "b" || got[1] != "c" {
		t.Fatalf("Upcoming() = %v, want [b c]", got)
	}

	q.Advance() // b
	if got := titles(q.History()); len(got) != 1 || got[0] != "a" {
		t.Fatalf("History() = %v, want [a]", got)
	}
}

func TestQueue_Shuffle_KeepsHistoryInPlace(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"), track("c"), track("d"), track("e"))
	q.Advance() // a is now history, b..e is upcoming

	q.Shuffle()

	all := titles(q.All())
	if all[0] != "a" {
		t.Fatalf("All()[0] = %q, want the already-played \"a\" to stay in place", all[0])
	}
	// The shuffled tail should still be exactly {b,c,d,e}, just reordered.
	seen := map[string]bool{}
	for _, title := range all[1:] {
		seen[title] = true
	}
	for _, want := range []string{"b", "c", "d", "e"} {
		if !seen[want] {
			t.Fatalf("All() after Shuffle() is missing %q: %v", want, all)
		}
	}
}

func TestQueue_Clear(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"))
	q.Advance()
	q.Clear()

	if !q.IsEmpty() {
		t.Fatal("expected queue to be empty after Clear()")
	}
	if q.Position() != -1 {
		t.Fatalf("Position() = %d, want -1 after Clear()", q.Position())
	}
}

func TestQueue_SkipTo(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"), track("c"))

	got, ok := q.SkipTo(2)
	if !ok || got.Info.Title != "c" {
		t.Fatalf("SkipTo(2) = (%q, %v), want (\"c\", true)", got.Info.Title, ok)
	}
	if _, ok := q.SkipTo(99); ok {
		t.Fatal("expected SkipTo with an out-of-range index to fail")
	}
}

func TestQueue_RemoveAt(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"), track("c"))
	q.Advance() // a is current, position 0

	removed, ok := q.RemoveAt(2)
	if !ok || removed.Info.Title != "c" {
		t.Fatalf("RemoveAt(2) = (%q, %v), want (\"c\", true)", removed.Info.Title, ok)
	}
	if q.Length() != 2 {
		t.Fatalf("Length() = %d, want 2 after removing a track", q.Length())
	}

	if _, ok := q.RemoveAt(0); ok {
		t.Fatal("expected RemoveAt on the currently playing track to fail")
	}
}

func TestWithRequesterAndRequester_RoundTrip(t *testing.T) {
	tr := track("a")
	userID := snowflake.ID(123456789)

	withRequester := WithRequester(tr, userID)
	got, ok := Requester(withRequester)
	if !ok || got != userID.String() {
		t.Fatalf("Requester() = (%q, %v), want (%q, true)", got, ok, userID.String())
	}
}

func TestRequester_MissingUserData(t *testing.T) {
	if _, ok := Requester(track("a")); ok {
		t.Fatal("expected Requester() to report no requester for a track with no user data")
	}
}

func TestManager_GetCreatesAndReuses(t *testing.T) {
	m := NewManager()
	guildID := snowflake.ID(1)

	q1 := m.Get(guildID)
	q1.Add(track("a"))

	q2 := m.Get(guildID)
	if q2.Length() != 1 {
		t.Fatalf("Get() returned a different queue for the same guild: Length() = %d, want 1", q2.Length())
	}
}

func TestManager_Delete(t *testing.T) {
	m := NewManager()
	guildID := snowflake.ID(1)

	m.Get(guildID).Add(track("a"))
	m.Delete(guildID)

	if m.Get(guildID).Length() != 0 {
		t.Fatal("expected a fresh, empty queue after Delete()")
	}
}

func TestQueue_AdvanceAfterExhaustionPlaysNewlyAddedTrack(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"))
	q.Advance() // a plays
	if _, ok := q.Advance(); ok {
		t.Fatal("expected the queue to be exhausted after the only track")
	}

	q.Add(track("b"))
	next, ok := q.Advance()
	if !ok || next.Info.Title != "b" {
		t.Fatalf("Advance() after exhaustion + Add = (%q, %v), want (\"b\", true)", next.Info.Title, ok)
	}
}

func TestQueue_PreviousAfterExhaustionReturnsEarlierTrack(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"))
	q.Advance() // a
	q.Advance() // b
	q.Advance() // exhausted

	prev, ok := q.Previous()
	if !ok || prev.Info.Title != "a" {
		t.Fatalf("Previous() after exhaustion = (%q, %v), want (\"a\", true)", prev.Info.Title, ok)
	}
}

func TestQueue_At(t *testing.T) {
	q := NewQueue()
	q.Add(track("a"), track("b"))

	got, ok := q.At(1)
	if !ok || got.Info.Title != "b" {
		t.Fatalf("At(1) = (%q, %v), want (\"b\", true)", got.Info.Title, ok)
	}
	if q.Position() != -1 {
		t.Fatalf("At() moved the position to %d, want it untouched at -1", q.Position())
	}
	if _, ok := q.At(2); ok {
		t.Fatal("expected At() with an out-of-range index to fail")
	}
	if _, ok := q.At(-1); ok {
		t.Fatal("expected At() with a negative index to fail")
	}
}
