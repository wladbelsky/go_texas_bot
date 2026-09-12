// Package settings implements /settings, letting server admins opt into
// optional/sensitive bot features on a per-guild basis (starting with the
// Arknights case simulator's NSFW content).
package settings

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
	"go_texas_bot/db"
)

func init() {
	command_selector.CommandSelector.AddCommand("settings", settingsCommandListener)
}

func settingsCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID := event.GuildID()
	if guildID == nil {
		return respond(event, "Эту команду можно использовать только на сервере.")
	}

	member := event.Member()
	if member == nil || !member.Permissions.Has(discord.PermissionManageGuild, discord.PermissionAdministrator) {
		return respond(event, "Нужны права на управление сервером.")
	}

	data := event.SlashCommandInteractionData()
	nsfwEnabled, hasNSFW := data.OptBool("nsfw-content")
	if !hasNSFW {
		return respond(event, "Ничего не изменилось.")
	}

	var current db.GuildSettings
	if err := db.DB.FirstOrCreate(&current, db.GuildSettings{GuildID: guildID.String()}).Error; err != nil {
		return err
	}
	current.NSFWEnabled = nsfwEnabled
	if err := db.DB.Save(&current).Error; err != nil {
		return err
	}

	status := "выключен"
	if nsfwEnabled {
		status = "включён"
	}
	return respond(event, "NSFW-контент (симулятор кейсов Arknights и т.п.) "+status+" для этого сервера.")
}

func respond(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().
		WithContent(content).
		WithEphemeral(true),
	)
}
