// Package ger implements /ger, the Go port of the original Python bot's
// toilet-humor "farts on a random member" cog.
package ger

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go_texas_bot/command/command_selector"
	"go_texas_bot/command/cooldown"
	"go_texas_bot/db"
	"go_texas_bot/guildsettings"
)

func init() {
	command_selector.CommandSelector.AddCommand(command_selector.Key("ger", discord.ApplicationCommandTypeSlash), gerCommandListener)
}

// selfChance is out of the same [0,101] roll range as the original bot's
// `randint(0, 101) >= self_ger_chance`.
const selfChance = 10

// Ported verbatim from the original bot's config.json ger.phrase_variants
// and ger.self_phrase_variants.
var phrases = []string{
	"пернул в ротешник", "насрал в рот", "высрал какулю на лицо", "пернул в ухо",
	"пустил шептуна за щеку", "запустил струю подливы в ротешник",
}

var selfPhrases = []string{
	"обосрался с подливой", "напрудил в штанишки", "пустил шептуна",
	"выдал звучную трель", "громогласно перданул", "жоплодирует!",
}

// gerCooldown mirrors the original bot's per-guild-per-user /ger cooldown
// (config.json ger.ger_cooldown=79200s).
var gerCooldown = cooldown.New(22 * time.Hour)

var errNoOtherMembers = errors.New("на сервере больше никого нет, не в кого пукать")

func gerCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := guildsettings.RequireGuildNSFW(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}

	authorID := event.User().ID
	if remaining := gerCooldown.Reserve(guildID.String(), authorID.String()); remaining > 0 {
		return respondEphemeral(event, fmt.Sprintf("Не так быстро! Попробуй снова через %s.", remaining.Round(time.Second)))
	}

	target, ok := randomOtherMember(event, guildID, authorID)
	if !ok {
		gerCooldown.Release(guildID.String(), authorID.String())
		return respondEphemeral(event, errNoOtherMembers.Error())
	}

	isSelf := rand.Intn(102) < selfChance
	if err = recordStats(event, authorID, target, isSelf); err != nil {
		gerCooldown.Release(guildID.String(), authorID.String())
		return err
	}

	if isSelf {
		return event.CreateMessage(discord.NewMessageCreate().
			WithContent(fmt.Sprintf("%s %s", discord.UserMention(authorID), selfPhrases[rand.Intn(len(selfPhrases))])),
		)
	}
	return event.CreateMessage(discord.NewMessageCreate().
		WithContent(fmt.Sprintf("%s %s %s", discord.UserMention(authorID), phrases[rand.Intn(len(phrases))], discord.UserMention(target.User.ID))),
	)
}

func randomOtherMember(event *events.ApplicationCommandInteractionCreate, guildID, authorID snowflake.ID) (discord.Member, bool) {
	var candidates []discord.Member
	for member := range event.Client().Caches.Members(guildID) {
		if member.User.ID != authorID {
			candidates = append(candidates, member)
		}
	}
	if len(candidates) == 0 {
		return discord.Member{}, false
	}
	return candidates[rand.Intn(len(candidates))], true
}

func recordStats(event *events.ApplicationCommandInteractionCreate, authorID snowflake.ID, target discord.Member, isSelf bool) error {
	var stats db.GerStats
	if err := db.DB.FirstOrCreate(&stats, db.GerStats{ID: 1}).Error; err != nil {
		return err
	}
	stats.Total++
	if target.User.Bot {
		stats.Bot++
	}
	if target.User.ID == event.Client().ApplicationID {
		stats.Me++
	}
	if isSelf {
		stats.Self++
	}
	if err := db.DB.Save(&stats).Error; err != nil {
		return err
	}

	if err := bumpUserCounter(target.User.ID, func(c *db.UserGerCounter) { c.Hits++ }); err != nil {
		return err
	}
	return bumpUserCounter(authorID, func(c *db.UserGerCounter) { c.Uses++ })
}

func bumpUserCounter(userID snowflake.ID, apply func(*db.UserGerCounter)) error {
	var counter db.UserGerCounter
	if err := db.DB.FirstOrCreate(&counter, db.UserGerCounter{UserID: userID.String()}).Error; err != nil {
		return err
	}
	apply(&counter)
	return db.DB.Save(&counter).Error
}

func respondEphemeral(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().WithContent(content).WithEphemeral(true))
}
