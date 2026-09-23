package music

import (
	"strings"
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

func TestFailureMessage(t *testing.T) {
	got := failureMessage(track("Song"), lavalink.Exception{
		Message: "(yts.version: 1.18.2) All clients failed to load the item.\nClient [WEB] failed: `x`",
	})
	want := "Не удалось воспроизвести **Song**: `(yts.version: 1.18.2) All clients failed to load the item.`"
	if got != want {
		t.Fatalf("failureMessage() = %q, want %q", got, want)
	}
}

func TestFailureMessage_NoReasonAndLongReason(t *testing.T) {
	if got := failureMessage(track("Song"), lavalink.Exception{}); got != "Не удалось воспроизвести **Song**" {
		t.Fatalf("failureMessage() without a reason = %q", got)
	}

	long := failureMessage(track("Song"), lavalink.Exception{Message: strings.Repeat("я`", 300)})
	if strings.Count(long, "`") != 2 {
		t.Fatalf("backticks from the reason must not break the code span: %q", long)
	}
	if n := len([]rune(long)); n > 260 {
		t.Fatalf("failureMessage() is %d runes long, want the reason truncated", n)
	}
}

func TestOnTrackException_NotifiesQueueChannel(t *testing.T) {
	type sent struct {
		channel snowflake.ID
		content string
	}
	var got []sent
	SetNotifier(func(channelID snowflake.ID, content string) { got = append(got, sent{channelID, content}) })
	t.Cleanup(func() { SetNotifier(nil) })

	guildID := snowflake.ID(7001)
	t.Cleanup(func() { Queues.Delete(guildID) })

	event := lavalink.TrackExceptionEvent{Track: track("Song"), Exception: lavalink.Exception{Message: "boom"}, GuildID_: guildID}

	onTrackException(nil, event)
	if len(got) != 0 {
		t.Fatalf("notified %v, but the guild has no text channel yet", got)
	}

	Queues.Get(guildID).SetTextChannel(99)
	onTrackException(nil, event)
	if len(got) != 1 || got[0].channel != 99 || !strings.Contains(got[0].content, "Song") {
		t.Fatalf("notifications = %+v, want one message about Song in channel 99", got)
	}
}
