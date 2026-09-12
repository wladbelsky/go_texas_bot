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
	discord.SlashCommandCreate{
		Name:        "ark",
		Description: "Кидает рандомную девочку (или кунчика) и сохраняет её в коллекцию",
		NSFW:        boolPtr(true),
	},
	discord.SlashCommandCreate{
		Name:        "myark",
		Description: "Показать свою коллекцию операторов Arknights (или конкретного персонажа)",
		NSFW:        boolPtr(true),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "character",
				Description: "Имя персонажа",
			},
			discord.ApplicationCommandOptionBool{
				Name:        "public",
				Description: "Показать ответ всем, а не только тебе",
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "barter",
		Description: "Обменять дубликаты операторов на более редких",
		NSFW:        boolPtr(true),
	},
	discord.SlashCommandCreate{
		Name:        "settings",
		Description: "Настройки бота для этого сервера (только для администраторов)",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{
				Name:        "nsfw-content",
				Description: "Включить NSFW-контент (симулятор кейсов Arknights и т.п.) на этом сервере",
			},
		},
	},
}

func boolPtr(v bool) *bool { return &v }

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
	slog.Error("command error", "err", err)
	if sendErr := event.CreateMessage(buildErrorMessage(err)); sendErr != nil {
		slog.Error("error sending error message", "err", sendErr)
	}
}

func buildErrorMessage(err error) discord.MessageCreate {
	return discord.NewMessageCreate().
		WithContent("The command failed to execute.\nError: " + err.Error()).
		WithEphemeral(true)
}
