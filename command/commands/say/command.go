package say

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
)

func init() {
	command_selector.CommandSelector.AddCommand("say", sayCommandListener)
}

func sayCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	data := event.SlashCommandInteractionData()
	return event.CreateMessage(buildSayMessage(data.String("message"), data.Bool("ephemeral")))
}

func buildSayMessage(message string, ephemeral bool) discord.MessageCreate {
	return discord.NewMessageCreate().
		WithContent(message).
		WithEphemeral(ephemeral)
}
