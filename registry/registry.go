// Package registry holds the bot's Discord application command
// declarations (the payload sent to Discord's bulk command-registration
// endpoint). It's a standalone leaf package -- rather than living on
// command.Commands -- specifically so that command/commands/help can read
// it to build /help without creating an import cycle through
// command/register.go's blank imports of every command subpackage.
package registry

import (
	"github.com/disgoorg/disgo/discord"
	"go_texas_bot/reactions"
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
			discord.ApplicationCommandOptionBool{
				Name:        "toxic-greetings",
				Description: "Писать в канал «основной», когда участник выходит с сервера",
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "play",
		Description: "Найти и включить музыку (YouTube, ссылка или поиск)",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "query",
				Description: "Поисковый запрос или ссылка",
				Required:    true,
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "skip",
		Description: "Пропустить текущий трек",
	},
	discord.SlashCommandCreate{
		Name:        "stop",
		Description: "Остановить музыку и очистить очередь",
	},
	discord.SlashCommandCreate{
		Name:        "queue",
		Description: "Показать очередь",
	},
	discord.SlashCommandCreate{
		Name:        "disconnect",
		Description: "Отключиться от голосового канала",
	},
	discord.SlashCommandCreate{
		Name:        "shuffle",
		Description: "Перемешать очередь",
	},
	discord.SlashCommandCreate{
		Name:        "repeat",
		Description: "Установить режим повтора",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "mode",
				Description: "Режим повтора",
				Required:    true,
				Choices: []discord.ApplicationCommandOptionChoiceString{
					{Name: "Выключен", Value: "none"},
					{Name: "Один трек", Value: "one"},
					{Name: "Вся очередь", Value: "all"},
				},
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "skipto",
		Description: "Перейти сразу к треку под номером в очереди",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "index",
				Description: "Номер трека в очереди",
				Required:    true,
				MinValue:    intPtr(1),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "pop",
		Description: "Удалить трек из очереди по номеру",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionInt{
				Name:        "index",
				Description: "Номер трека в очереди",
				Required:    true,
				MinValue:    intPtr(1),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "controls",
		Description: "Показать плеер с кнопками управления",
	},
	discord.SlashCommandCreate{
		Name:        "sfw",
		Description: "Аниме пикчи)",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "type",
				Description: "Выбери что хочешь посмотреть",
				Required:    true,
				Choices:     stringChoices("waifu", "neko", "awoo", "shinobu", "megumin"),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "reaction",
		Description: "Аниме реакшоны",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "type",
				Description: "Выбери эмоцию",
				Required:    true,
				Choices:     stringChoices(reactions.Categories...),
			},
			discord.ApplicationCommandOptionUser{
				Name:        "member",
				Description: "Выбери кого упомянуть (для парных эмоций)",
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "nsfw",
		Description: "Пошлые аниме пикчи))",
		NSFW:        boolPtr(true),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "type",
				Description: "Выбери что хочешь посмотреть",
				Choices:     stringChoices("waifu", "neko", "trap", "blowjob"),
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "ger",
		Description: "Пукает в рандома, или в себя)",
		NSFW:        boolPtr(true),
	},
	discord.SlashCommandCreate{
		Name:        "ping",
		Description: "Замеряет задержку в развитии, твоем)",
	},
	discord.SlashCommandCreate{
		Name:        "announce",
		Description: "Я скажу все что ты хочешь, братик (только для администраторов)",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "message",
				Description: "Сообщение",
				Required:    true,
			},
			discord.ApplicationCommandOptionChannel{
				Name:         "channel",
				Description:  "В какой канал отправить",
				Required:     true,
				ChannelTypes: []discord.ChannelType{discord.ChannelTypeGuildText},
			},
		},
	},
	discord.SlashCommandCreate{
		Name:        "info",
		Description: "Информация и статистика бота",
	},
	discord.SlashCommandCreate{
		Name:        "invite",
		Description: "Показать ссылку-приглашение этого бота",
	},
	discord.SlashCommandCreate{
		Name:        "f",
		Description: "Отдать честь за почивших героев. Можно упоминанием указать кого чтим.",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "member",
				Description: "Кого почтить",
			},
		},
	},
	discord.UserCommandCreate{Name: "f"},
	discord.SlashCommandCreate{
		Name:        "o7",
		Description: "Поприветствовать командиров, а можно и кого-то конкретного",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "member",
				Description: "Кого поприветствовать",
			},
		},
	},
	discord.UserCommandCreate{Name: "o7"},
	discord.SlashCommandCreate{
		Name:        "avatar",
		Description: "Показать аватарку участника",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "member",
				Description: "Чью аватарку показать",
				Required:    true,
			},
		},
	},
	discord.UserCommandCreate{Name: "avatar"},
	discord.SlashCommandCreate{
		Name:        "help",
		Description: "Показать помощь, ня",
	},
}

func boolPtr(v bool) *bool { return &v }
func intPtr(v int) *int    { return &v }

func stringChoices(values ...string) []discord.ApplicationCommandOptionChoiceString {
	choices := make([]discord.ApplicationCommandOptionChoiceString, len(values))
	for i, v := range values {
		choices[i] = discord.ApplicationCommandOptionChoiceString{Name: v, Value: v}
	}
	return choices
}
