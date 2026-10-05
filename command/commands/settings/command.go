// Package settings implements /settings, letting server admins opt into
// optional/sensitive bot features on a per-guild basis.
package settings

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
	"go_texas_bot/db"
	"go_texas_bot/guildsettings"
)

func init() {
	command_selector.CommandSelector.AddCommand(command_selector.Key("settings", discord.ApplicationCommandTypeSlash), settingsCommandListener)
}

// canManageGuild reports whether the member may change bot settings. It
// needs two separate Has calls: disgo's Permissions.Has requires *all* the
// given bits, not any of them.
func canManageGuild(p discord.Permissions) bool {
	return p.Has(discord.PermissionManageGuild) || p.Has(discord.PermissionAdministrator)
}

func settingsCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID := event.GuildID()
	if guildID == nil {
		return respond(event, "Эту команду можно использовать только на сервере.")
	}

	member := event.Member()
	if member == nil || !canManageGuild(member.Permissions) {
		return respond(event, "Нужны права на управление сервером.")
	}

	current, err := guildsettings.Get(guildID.String())
	if err != nil {
		return err
	}

	data := event.SlashCommandInteractionData()
	changed := false
	if v, ok := data.OptBool("nsfw-content"); ok {
		current.NSFWEnabled = v
		changed = true
	}
	if v, ok := data.OptBool("toxic-greetings"); ok {
		current.ToxicGreetingsEnabled = v
		changed = true
	}
	if changed {
		if err = db.DB.Save(&current).Error; err != nil {
			return err
		}
	}

	return respond(event, describe(current))
}

func describe(s db.GuildSettings) string {
	return fmt.Sprintf(
		"Настройки сервера:\n• NSFW-контент (симулятор кейсов Arknights, /nsfw, /ger): %s\n• Сообщение при выходе участника (в канал «основной»): %s",
		onOff(s.NSFWEnabled), onOff(s.ToxicGreetingsEnabled),
	)
}

func onOff(v bool) string {
	if v {
		return "включено"
	}
	return "выключено"
}

func respond(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().
		WithContent(content).
		WithEphemeral(true),
	)
}
