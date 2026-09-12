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
	err := event.CreateMessage(discord.NewMessageCreate().
		WithContent(data.String("message")).
		WithEphemeral(data.Bool("ephemeral")),
	)
	return err
}
