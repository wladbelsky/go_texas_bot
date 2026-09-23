// Package ark implements the /ark, /myark and /barter slash commands, the
// Go port of the original Python bot's ark.py "case simulator" cog.
package ark

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/arknights"
	"go_texas_bot/command/command_selector"
	"go_texas_bot/guildsettings"
)

func init() {
	command_selector.CommandSelector.AddCommand(command_selector.Key("ark", discord.ApplicationCommandTypeSlash), arkCommandListener)
	command_selector.CommandSelector.AddCommand(command_selector.Key("myark", discord.ApplicationCommandTypeSlash), myArkCommandListener)
	command_selector.CommandSelector.AddCommand(command_selector.Key("barter", discord.ApplicationCommandTypeSlash), barterCommandListener)
}

func arkCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := guildsettings.RequireGuildNSFW(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}

	userID := event.User().ID.String()
	if remaining := rollCooldown.Reserve(guildID.String(), userID); remaining > 0 {
		return respondEphemeral(event, fmt.Sprintf("Не так быстро! Попробуй снова через %s.", remaining.Round(time.Second)))
	}

	character, err := arknights.Roll(userID)
	if err != nil {
		rollCooldown.Release(guildID.String(), userID)
		return err
	}

	return event.CreateMessage(discord.NewMessageCreate().
		WithEmbeds(characterEmbed(character, event.User().EffectiveName())),
	)
}

func myArkCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	if _, err := guildsettings.RequireGuildNSFW(event); err != nil {
		return respondEphemeral(event, err.Error())
	}

	data := event.SlashCommandInteractionData()
	userID := event.User().ID.String()
	public, _ := data.OptBool("public")
	charName, hasCharName := data.OptString("character")

	if !hasCharName || charName == "" {
		return respondMyArkCollection(event, userID, public)
	}
	return respondMyArkCharacter(event, userID, charName, public)
}

func respondMyArkCollection(event *events.ApplicationCommandInteractionCreate, userID string, public bool) error {
	collection, err := arknights.Collection(userID)
	if err != nil {
		return err
	}

	byRarity := make(map[int][]collectionRow, len(collection))
	owned := 0
	for rarity, entries := range collection {
		for _, e := range entries {
			byRarity[rarity] = append(byRarity[rarity], collectionRow{Name: e.OperatorName, Count: e.Count})
			owned++
		}
	}

	embed := collectionEmbed(event.User().EffectiveName(), owned, arknights.Count(), byRarity)
	return event.CreateMessage(discord.NewMessageCreate().
		WithEmbeds(embed).
		WithEphemeral(!public),
	)
}

func respondMyArkCharacter(event *events.ApplicationCommandInteractionCreate, userID, charName string, public bool) error {
	character, ok := arknights.ByName(charName)
	if !ok {
		return respondEphemeral(event, "***Лошара, даже имя своей вайфу не запомнил((***")
	}

	_, owned, err := arknights.OwnedEntry(userID, character)
	if err != nil {
		return err
	}
	if !owned {
		return respondEphemeral(event, "***Лох, у тебя нет такой дивочки***")
	}

	return event.CreateMessage(discord.NewMessageCreate().
		WithEmbeds(characterEmbed(character, event.User().EffectiveName())).
		WithEphemeral(!public),
	)
}

func barterCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	_, err := guildsettings.RequireGuildNSFW(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}

	userID := event.User().ID.String()
	results, err := arknights.Barter(userID)
	if err != nil {
		return err
	}
	if len(results) == 0 {
		return respondEphemeral(event, "***Нет операторов на обмен***")
	}

	names := make([]string, 0, len(results))
	for _, c := range results {
		names = append(names, fmt.Sprintf("%s %s", stars(c.Rarity), c.DisplayName()))
	}

	embed := discord.Embed{
		Title:       "Обмен завершён!",
		Description: "Наслаждайся!",
		Color:       embedColor,
		Fields:      []discord.EmbedField{blockField("Получено", strings.Join(names, "\n"))},
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed))
}

func respondEphemeral(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().
		WithContent(content).
		WithEphemeral(true),
	)
}
