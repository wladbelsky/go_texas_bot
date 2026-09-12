package ark

import (
	"errors"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/db"
)

// cooldown mirrors the original bot's per-guild-per-user 4h cooldown on
// /ark. It's kept in memory (reset on restart) rather than persisted, since
// losing a partial cooldown on redeploy is a fine trade-off for a personal
// bot.
const cooldown = 4 * time.Hour

var lastRoll sync.Map // map[string]time.Time, keyed by "guildID:userID"

var (
	errNotInGuild   = errors.New("эту команду можно использовать только на сервере")
	errNSFWChannel  = errors.New("эту команду можно использовать только в NSFW-канале")
	errNSFWDisabled = errors.New("NSFW-контент отключён на этом сервере: попроси администратора включить его через /settings")
)

// requireGuildNSFW checks that the interaction happened in a guild, in a
// channel flagged NSFW by Discord, and that the guild has opted into NSFW
// content via /settings. It returns the guild ID on success.
func requireGuildNSFW(event *events.ApplicationCommandInteractionCreate) (string, error) {
	guildID := event.GuildID()
	if guildID == nil {
		return "", errNotInGuild
	}

	channel, ok := event.Channel().MessageChannel.(discord.GuildMessageChannel)
	if !ok || !channel.NSFW() {
		return "", errNSFWChannel
	}

	settings, err := getGuildSettings(guildID.String())
	if err != nil {
		return "", err
	}
	if !settings.NSFWEnabled {
		return "", errNSFWDisabled
	}

	return guildID.String(), nil
}

func getGuildSettings(guildID string) (db.GuildSettings, error) {
	var settings db.GuildSettings
	err := db.DB.FirstOrCreate(&settings, db.GuildSettings{GuildID: guildID}).Error
	return settings, err
}

func checkCooldown(guildID, userID string) time.Duration {
	key := guildID + ":" + userID
	now := time.Now()
	if v, ok := lastRoll.Load(key); ok {
		if remaining := cooldown - now.Sub(v.(time.Time)); remaining > 0 {
			return remaining
		}
	}
	lastRoll.Store(key, now)
	return 0
}
