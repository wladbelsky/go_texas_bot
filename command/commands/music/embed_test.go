package music

import (
	"strings"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	musicpkg "go_texas_bot/music"
)

func TestFormatDuration(t *testing.T) {
	if got := formatDuration(90 * lavalink.Second); got != "1:30" {
		t.Fatalf("formatDuration(90s) = %q, want %q", got, "1:30")
	}
	if got := formatDuration(5 * lavalink.Second); got != "0:05" {
		t.Fatalf("formatDuration(5s) = %q, want %q", got, "0:05")
	}
}

func TestProgressBar(t *testing.T) {
	if got := progressBar(0, 0); got != "" {
		t.Fatalf("progressBar with zero length = %q, want empty", got)
	}
	full := progressBar(100*lavalink.Second, 100*lavalink.Second)
	if strings.Count(full, "🟦") != progressBarLength {
		t.Fatalf("progressBar at 100%% = %q, want all filled", full)
	}
	half := progressBar(50*lavalink.Second, 100*lavalink.Second)
	if strings.Count(half, "🟦") != progressBarLength/2 {
		t.Fatalf("progressBar at 50%% = %q, want half filled", half)
	}
}

func TestPlayStateEmoji_NilPlayer(t *testing.T) {
	if got := playStateEmoji(nil); got != "⏹️" {
		t.Fatalf("playStateEmoji(nil) = %q, want stopped emoji", got)
	}
}

func TestNowPlayingEmbed_NilPlayer(t *testing.T) {
	embed := nowPlayingEmbed(nil, musicpkg.NewQueue())
	if len(embed.Fields) != 1 {
		t.Fatalf("expected a single placeholder field, got %+v", embed.Fields)
	}
	if !strings.Contains(embed.Fields[0].Value, "/play") {
		t.Fatalf("expected the empty-queue field to mention /play, got %q", embed.Fields[0].Value)
	}
}

func TestQueueEmbed_Empty(t *testing.T) {
	embed := queueEmbed(musicpkg.NewQueue(), 10)
	if embed.Fields[0].Name != "Сейчас играет" {
		t.Fatalf("expected a \"Сейчас играет\" field, got %+v", embed.Fields)
	}
	if !strings.Contains(embed.Fields[0].Value, "/play") {
		t.Fatalf("expected the empty-queue message to mention /play, got %q", embed.Fields[0].Value)
	}
}

func TestQueueEmbed_ListsUpcomingCappedAtShow(t *testing.T) {
	q := musicpkg.NewQueue()
	q.Add(track("a"), track("b"), track("c"), track("d"))
	q.Advance() // a is now playing

	embed := queueEmbed(q, 2)

	var upcomingField *string
	for _, f := range embed.Fields {
		if f.Name == "Следующие 2 треков" {
			upcomingField = &f.Value
		}
	}
	if upcomingField == nil {
		t.Fatalf("expected a capped \"upcoming\" field, got %+v", embed.Fields)
	}
	if strings.Contains(*upcomingField, "d") {
		t.Fatalf("expected the upcoming list to be capped at 2 tracks, got %q", *upcomingField)
	}
}

func track(title string) lavalink.Track {
	return lavalink.Track{Info: lavalink.TrackInfo{Title: title}}
}
