// Package welcome implements the join/leave listeners ported from the
// original Python bot's welcome.py. Unlike the other command/commands/*
// packages it registers no slash command -- OnMemberJoin/OnMemberLeave are
// wired up directly as gateway event listeners in main.go.
package welcome

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go_texas_bot/guildsettings"
)

// leaveChannelName is the guild text channel the original bot posted its
// leave message to, found by name rather than any configured ID -- ported
// as-is, so this only does anything on guilds that happen to have a
// channel named exactly this.
const leaveChannelName = "основной"

// OnMemberJoin DMs new members, mirroring the original's on_member_join
// (which, despite its docstring claiming to send rules, just sends "f").
func OnMemberJoin(event *events.GuildMemberJoin) {
	dm, err := event.Client().Rest.CreateDMChannel(event.Member.User.ID)
	if err != nil {
		slog.Error("welcome: failed to open DM channel", "err", err)
		return
	}
	if _, err = event.Client().Rest.CreateMessage(dm.ID(), discord.NewMessageCreate().WithContent("f")); err != nil {
		slog.Error("welcome: failed to send join DM", "err", err)
	}
}

// OnMemberLeave posts a hostile message to the guild's "основной" channel
// when it leaves, mirroring the original's on_member_remove -- but only if
// the guild has opted into it via /settings (toxic-greetings-enabled),
// unlike the original which always sent it unconditionally.
func OnMemberLeave(event *events.GuildMemberLeave) {
	settings, err := guildsettings.Get(event.GuildID.String())
	if err != nil {
		slog.Error("welcome: failed to load guild settings", "err", err)
		return
	}
	if !settings.ToxicGreetingsEnabled {
		return
	}

	channels, err := event.Client().Rest.GetGuildChannels(event.GuildID)
	if err != nil {
		slog.Error("welcome: failed to list guild channels", "err", err)
		return
	}
	channelID, ok := findChannelByName(channels, leaveChannelName)
	if !ok {
		return
	}

	if _, err = event.Client().Rest.CreateMessage(channelID, discord.NewMessageCreate().WithContent("fck u")); err != nil {
		slog.Error("welcome: failed to send leave message", "err", err)
	}
}

func findChannelByName(channels []discord.GuildChannel, name string) (snowflake.ID, bool) {
	for _, ch := range channels {
		if ch.Name() == name {
			return ch.ID(), true
		}
	}
	return 0, false
}
