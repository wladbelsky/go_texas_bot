package music

import (
	"context"
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	musicpkg "go_texas_bot/music"
)

func controlsCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	player := existingPlayer(guildID)
	queue := musicpkg.Queues.Get(guildID)

	return event.CreateMessage(discord.NewMessageCreate().
		WithEmbeds(nowPlayingEmbed(player, queue)).
		WithComponents(controlsComponents()...),
	)
}

// ComponentListener handles clicks on the /controls buttons. It's
// registered as its own top-level event listener (alongside
// command.Listener) since button presses arrive as a different interaction
// type than slash commands.
func ComponentListener(event *events.ComponentInteractionCreate) {
	customID := event.ButtonInteractionData().CustomID()
	guildID := event.GuildID()
	if guildID == nil {
		return
	}

	player := existingPlayer(*guildID)
	queue := musicpkg.Queues.Get(*guildID)

	var actionErr error
	switch customID {
	case customIDPrev:
		actionErr = handlePrev(event, player, queue)
	case customIDPause:
		actionErr = handlePause(player)
	case customIDStop:
		actionErr = handleStop(event, player, queue)
	case customIDNext:
		actionErr = handleNext(event, player, queue)
	case customIDRefresh:
		// nothing to do, we re-render below either way
	default:
		return
	}

	if actionErr != nil {
		slog.Error("music: control button failed", "custom_id", customID, "err", actionErr)
		_ = event.CreateMessage(discord.NewMessageCreate().WithContent("Не получилось: " + actionErr.Error()).WithEphemeral(true))
		return
	}

	if err := event.UpdateMessage(discord.NewMessageUpdate().
		WithEmbeds(nowPlayingEmbed(player, queue)).
		WithComponents(controlsComponents()...),
	); err != nil {
		slog.Error("music: failed to refresh controls message", "err", err)
	}
}

func handlePrev(event *events.ComponentInteractionCreate, player disgolink.Player, queue *musicpkg.Queue) error {
	if player == nil || player.Track() == nil {
		return errNothingPlaying
	}
	if err := requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *player.Track()); err != nil {
		return err
	}
	prev, ok := queue.Previous()
	if !ok {
		return nil
	}
	return player.Update(context.Background(), lavalink.WithTrack(prev))
}

func handlePause(player disgolink.Player) error {
	if player == nil {
		return errNothingPlaying
	}
	return player.Update(context.Background(), lavalink.WithPaused(!player.Paused()))
}

func handleStop(event *events.ComponentInteractionCreate, player disgolink.Player, queue *musicpkg.Queue) error {
	if player == nil {
		return errNothingPlaying
	}
	if track := player.Track(); track != nil {
		if err := requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *track); err != nil {
			return err
		}
	}
	queue.Clear()
	return player.Update(context.Background(), lavalink.WithNullTrack())
}

func handleNext(event *events.ComponentInteractionCreate, player disgolink.Player, queue *musicpkg.Queue) error {
	if player == nil || player.Track() == nil {
		return errNothingPlaying
	}
	if err := requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *player.Track()); err != nil {
		return err
	}
	next, ok := queue.Advance()
	if !ok {
		return player.Update(context.Background(), lavalink.WithNullTrack())
	}
	return player.Update(context.Background(), lavalink.WithTrack(next))
}
