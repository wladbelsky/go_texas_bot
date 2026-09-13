package music

import (
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"go_texas_bot/music"
)

var (
	errNotInGuild     = errors.New("эту команду можно использовать только на сервере")
	errLavalinkDown   = errors.New("не удалось подключиться к музыкальному серверу, попробуй позже")
	errNothingPlaying = errors.New("сейчас ничего не играет")
	errNotYourTrack   = errors.New("нельзя управлять чужим треком")
)

// mustParseSnowflake parses a snowflake ID stashed via music.WithRequester,
// falling back to 0 (an invalid/unmentionable ID) if it's ever malformed --
// which shouldn't happen since we're the only ones who write it.
func mustParseSnowflake(s string) snowflake.ID {
	id, err := snowflake.Parse(s)
	if err != nil {
		return 0
	}
	return id
}

func requireGuild(event *events.ApplicationCommandInteractionCreate) (snowflake.ID, error) {
	guildID := event.GuildID()
	if guildID == nil {
		return 0, errNotInGuild
	}
	return *guildID, nil
}

// existingPlayer returns the guild's player if one is already connected, or
// nil if there isn't one (or Lavalink itself is unreachable).
func existingPlayer(guildID snowflake.ID) disgolink.Player {
	if music.Lavalink == nil {
		return nil
	}
	return music.Lavalink.ExistingPlayer(guildID)
}

// requireTrackOwnerOrAdmin mirrors the original bot's permission check for
// skip/stop/prev: server admins may always act, otherwise only the member
// who requested the track may -- and a track with no recorded requester is
// treated as the caller's own (matching the Python
// "(get_track_owner() or ctx.user.id) == ctx.user.id" fallback).
func requireTrackOwnerOrAdmin(member *discord.ResolvedMember, userID snowflake.ID, track lavalink.Track) error {
	if member != nil && member.Permissions.Has(discord.PermissionAdministrator) {
		return nil
	}
	requesterID, ok := music.Requester(track)
	if !ok || requesterID == userID.String() {
		return nil
	}
	return errNotYourTrack
}
