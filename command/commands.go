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
	if sendErr := event.CreateMessage(buildErrorMessage(err)); sendErr != nil {
		slog.Error("error sending error message", "err", sendErr)
	}
}

func buildErrorMessage(err error) discord.MessageCreate {
	return discord.NewMessageCreate().
		WithContent("The command failed to execute.\nError: " + err.Error()).
		WithEphemeral(true)
}
