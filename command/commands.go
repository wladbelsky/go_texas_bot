package command

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
	"log/slog"
)

var Commands = []discord.ApplicationCommandCreate{
	discord.SlashCommandCreate{
		Name: "say",
		NameLocalizations: map[discord.Locale]string{
			discord.LocaleEnglishGB: "say",
			discord.LocaleRussian:   "скажи",
		},
		Description: "says what you say",
		DescriptionLocalizations: map[discord.Locale]string{
			discord.LocaleEnglishGB: "says what you say",
			discord.LocaleRussian:   "говорит то, что вы говорите",
		},
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name: "message",
				NameLocalizations: map[discord.Locale]string{
					discord.LocaleEnglishGB: "message",
					discord.LocaleRussian:   "сообщение",
				},
				Description: "What to say",
				DescriptionLocalizations: map[discord.Locale]string{
					discord.LocaleEnglishGB: "What to say",
					discord.LocaleRussian:   "Что сказать",
				},
				Required: true,
			},
			discord.ApplicationCommandOptionBool{
				Name: "ephemeral",
				NameLocalizations: map[discord.Locale]string{
					discord.LocaleEnglishGB: "ephemeral",
					discord.LocaleRussian:   "скрытый",
				},
				Description: "If the response should only be visible to you",
				DescriptionLocalizations: map[discord.Locale]string{
					discord.LocaleEnglishGB: "If the response should only be visible to you",
					discord.LocaleRussian:   "Если ответ должен быть виден только вам",
				},
				Required: true,
			},
		},
	},
}

func Listener(event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	command, ok := command_selector.CommandSelector.GetCommand(data.CommandName())
	if !ok {
		slog.Error("command not found")
		return
	}
	err := command(event)
	if err != nil {
		sendError(event, err)
	}
}

func sendError(event *events.ApplicationCommandInteractionCreate, err error) {
	slog.Error("command error:", err)
	err = event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("The command failed to execute.\nError: " + err.Error()).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		slog.Error("error sending error message:", err)
	}
}
