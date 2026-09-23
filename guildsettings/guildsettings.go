// Package guildsettings centralizes the per-guild feature toggles (starting
// with NSFW content) and the gate that command handlers use to enforce
// them. It's a standalone package (rather than living on the `command`
// package like the rest of the shared command plumbing) specifically so
// that command subpackages can import it without creating an import cycle
// through command/register.go's blank imports of those same subpackages.
package guildsettings

import (
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go_texas_bot/db"
)

var (
	ErrNotInGuild   = errors.New("эту команду можно использовать только на сервере")
	ErrNSFWChannel  = errors.New("эту команду можно использовать только в NSFW-канале")
	ErrNSFWDisabled = errors.New("NSFW-контент отключён на этом сервере: попроси администратора включить его через /settings")
)

// Get returns the guild's settings row, creating a default (all-disabled)
// one if it doesn't exist yet.
func Get(guildID string) (db.GuildSettings, error) {
	var settings db.GuildSettings
	err := db.DB.FirstOrCreate(&settings, db.GuildSettings{GuildID: guildID}).Error
	return settings, err
}

// RequireGuildNSFW checks that the interaction happened in a guild, in a
// channel flagged NSFW by Discord, and that the guild has opted into NSFW
// content via /settings. It returns the guild ID on success.
func RequireGuildNSFW(event *events.ApplicationCommandInteractionCreate) (snowflake.ID, error) {
	guildID := event.GuildID()
	if guildID == nil {
		return 0, ErrNotInGuild
	}

	channel, ok := event.Channel().MessageChannel.(discord.GuildMessageChannel)
	if !ok || !channel.NSFW() {
		return 0, ErrNSFWChannel
	}

	settings, err := Get(guildID.String())
	if err != nil {
		return 0, err
	}
	if !settings.NSFWEnabled {
		return 0, ErrNSFWDisabled
	}

	return *guildID, nil
}
