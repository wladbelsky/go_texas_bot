package music

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

// onTrackEnd advances the guild's queue and starts the next track whenever
// Lavalink reports the current one ending naturally (finished, or failed to
// load) -- mirrors the original bot's on_wavelink_track_end handler.
// Manual stops/replacements (skip, stop, /play while already playing) are
// left alone since the command handlers that caused them already know what
// should play next.
func onTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	if !event.Reason.MayStartNext() {
		return
	}

	next, ok := Queues.Get(event.GuildID()).Advance()
	if !ok {
		return
	}
	if err := player.Update(context.Background(), lavalink.WithTrack(next)); err != nil {
		slog.Error("music: failed to auto-advance queue", "err", err)
	}
}
