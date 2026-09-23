package command

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
	"log/slog"
)

func Listener(event *events.ApplicationCommandInteractionCreate) {
	key := command_selector.Key(event.Data.CommandName(), event.Data.Type())
	command, ok := command_selector.CommandSelector.GetCommand(key)
	if !ok {
		slog.Error("command not found", "key", key)
		return
	}
	err := command(event)
	if err != nil {
		sendError(event, err)
	}
}

func sendError(event *events.ApplicationCommandInteractionCreate, err error) {
	slog.Error("command error", "err", err)
	msg := buildErrorMessage(err)
	if createErr := event.CreateMessage(msg); createErr == nil {
		return
	}
	// Handlers that called DeferCreateMessage have already acknowledged the
	// interaction, so Discord rejects a second create; edit the deferred
	// response instead, or the user is left looking at "thinking..." forever.
	if _, updateErr := event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(),
		discord.NewMessageUpdate().WithContent(msg.Content)); updateErr != nil {
		slog.Error("error sending error message", "err", updateErr)
	}
}

func buildErrorMessage(err error) discord.MessageCreate {
	return discord.NewMessageCreate().
		WithContent("The command failed to execute.\nError: " + err.Error()).
		WithEphemeral(true)
}
