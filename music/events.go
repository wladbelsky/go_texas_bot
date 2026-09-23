package music

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
)

// Notifier posts a plain-text message to a Discord channel. main wires it to
// the bot's REST client; the music package itself doesn't talk to Discord.
type Notifier func(channelID snowflake.ID, content string)

var notify Notifier = func(snowflake.ID, string) {}

// SetNotifier sets where playback problems are reported. Call it before
// Connect; it isn't safe to change while events are being handled.
func SetNotifier(n Notifier) {
	if n == nil {
		n = func(snowflake.ID, string) {}
	}
	notify = n
}

// notifyGuild reports a playback problem in the text channel the guild's
// queue was last controlled from, if there is one.
func notifyGuild(guildID snowflake.ID, content string) {
	channelID := Queues.Get(guildID).TextChannel()
	if channelID == 0 {
		return
	}
	notify(channelID, content)
}

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
	queue := Queues.Get(event.GuildID())

	var next lavalink.Track
	var ok bool
	if event.Reason == lavalink.TrackEndReasonLoadFailed {
		// The failure itself was already reported by onTrackException.
		var exhausted bool
		next, ok, exhausted = queue.SkipFailed()
		if exhausted {
			notifyGuild(event.GuildID(), "Не удалось воспроизвести ни один трек из очереди, остановила воспроизведение")
		}
	} else {
		queue.ResetFailures()
		next, ok = queue.Advance()
	}
	if !ok {
		return
	}
	if err := player.Update(context.Background(), lavalink.WithTrack(next)); err != nil {
		slog.Error("music: failed to auto-advance queue", "err", err)
	}
}

// onTrackException reports a track Lavalink couldn't play (YouTube refusing
// the stream, a region block, a removed video...). Lavalink follows it with
// a TrackEnd(loadFailed) when the track couldn't start, which onTrackEnd
// handles by moving on to the next track.
func onTrackException(_ disgolink.Player, event lavalink.TrackExceptionEvent) {
	slog.Warn("music: track playback failed",
		"guild", event.GuildID(), "track", event.Track.Info.Title,
		"severity", event.Exception.Severity, "err", event.Exception.Message, "cause", event.Exception.Cause)
	notifyGuild(event.GuildID(), failureMessage(event.Track, event.Exception))
}

// onTrackStuck skips a track that stopped producing audio. Lavalink doesn't
// end stuck tracks on its own, so without this the player would sit silent.
func onTrackStuck(player disgolink.Player, event lavalink.TrackStuckEvent) {
	slog.Warn("music: track stuck", "guild", event.GuildID(), "track", event.Track.Info.Title, "threshold", event.Threshold)
	notifyGuild(event.GuildID(), fmt.Sprintf("Трек **%s** завис, переключаю дальше", event.Track.Info.Title))

	next, ok, _ := Queues.Get(event.GuildID()).SkipFailed()
	update := lavalink.WithNullTrack()
	if ok {
		update = lavalink.WithTrack(next)
	}
	if err := player.Update(context.Background(), update); err != nil {
		slog.Error("music: failed to skip stuck track", "err", err)
	}
}

// maxReasonLen keeps Lavalink's error text short enough for a chat message;
// youtube-source messages can include multi-line per-client details.
const maxReasonLen = 200

// failureMessage is the user-facing text for a track that failed to play.
func failureMessage(track lavalink.Track, exception lavalink.Exception) string {
	reason, _, _ := strings.Cut(strings.TrimSpace(exception.Message), "\n")
	if runes := []rune(reason); len(runes) > maxReasonLen {
		reason = string(runes[:maxReasonLen]) + "…"
	}
	msg := fmt.Sprintf("Не удалось воспроизвести **%s**", track.Info.Title)
	if reason != "" {
		msg += fmt.Sprintf(": `%s`", strings.ReplaceAll(reason, "`", "'"))
	}
	return msg
}
