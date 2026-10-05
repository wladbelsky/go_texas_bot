// Package reactions implements the /sfw, /reaction and /nsfw slash
// commands, the Go port of the original Python bot's reactions.py cog.
package reactions

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
	"go_texas_bot/guildsettings"
	"go_texas_bot/reactions"
)

const embedColor = 0xff9900 // matches the original bot's config.json embed_color

func init() {
	command_selector.CommandSelector.AddCommand(command_selector.Key("sfw", discord.ApplicationCommandTypeSlash), sfwCommandListener)
	command_selector.CommandSelector.AddCommand(command_selector.Key("reaction", discord.ApplicationCommandTypeSlash), reactionCommandListener)
	command_selector.CommandSelector.AddCommand(command_selector.Key("nsfw", discord.ApplicationCommandTypeSlash), nsfwCommandListener)
}

func reactionEmbed(ctx context.Context, category, title string, nsfw bool) discord.Embed {
	url, ok, err := reactions.FetchImageURL(ctx, category, nsfw)
	if err != nil || !ok {
		return discord.Embed{
			Title:       "Неудалось подключиться к бд картинками(",
			Description: "Наша команда пыталась найти пикчу, но потерпела неудачу((",
			Color:       embedColor,
			Image:       &discord.EmbedResource{URL: "https://c.tenor.com/tZ2Xd8LqAnMAAAAd/typing-fast.gif"},
		}
	}
	return discord.Embed{
		Title: title,
		Color: embedColor,
		Image: &discord.EmbedResource{URL: url},
	}
}

func sfwCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}
	category := event.SlashCommandInteractionData().String("type")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(),
		discord.NewMessageUpdate().WithEmbeds(reactionEmbed(ctx, category, "", false)))
	return err
}

func reactionCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	data := event.SlashCommandInteractionData()
	category := data.String("type")
	phrase, ok := reactions.Phrases[category]
	if !ok {
		return event.CreateMessage(discord.NewMessageCreate().WithContent("Я таких картинок не знаю!"))
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	target := ""
	if member, hasMember := data.OptMember("member"); hasMember {
		target = member.EffectiveName()
	}
	text := phrase.Format(event.User().EffectiveName(), target)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(),
		discord.NewMessageUpdate().WithEmbeds(reactionEmbed(ctx, category, text, false)))
	return err
}

// nsfwFooters are format templates (one %s for the display name) for the
// /nsfw command's footer, ported verbatim from the original bot.
var nsfwFooters = map[string]string{
	"waifu":   "%s смотрит хентай!",
	"neko":    "%s любитель фурри, ясно понятно",
	"trap":    "Фу бля, %s любитель трапов походу",
	"blowjob": "%s любит сосать, курильщик походу",
}

var nsfwCategories = []string{"waifu", "neko", "trap", "blowjob"}

func nsfwCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	if _, err := guildsettings.RequireGuildNSFW(event); err != nil {
		return respondEphemeral(event, err.Error())
	}

	category, ok := event.SlashCommandInteractionData().OptString("type")
	if !ok || category == "" {
		category = nsfwCategories[rand.Intn(len(nsfwCategories))]
	}

	if err := event.DeferCreateMessage(false); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	embed := reactionEmbed(ctx, category, "", true)
	if footer, ok := nsfwFooters[category]; ok {
		embed.Footer = &discord.EmbedFooter{Text: fmt.Sprintf(footer, event.User().EffectiveName())}
	}

	_, err := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdate().WithEmbeds(embed))
	return err
}

func respondEphemeral(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().WithContent(content).WithEphemeral(true))
}
