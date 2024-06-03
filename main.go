package main

import (
	"context"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"go_texas_bot/command"
	"go_texas_bot/config"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	token := config.Token()
	if token == "" {
		log.Panicln("token is required")
	}
	mainContext := context.Background()
	client, err := disgo.New(
		token,
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentDirectMessages,
				gateway.IntentDirectMessageReactions,
				gateway.IntentGuildMessageReactions,
				gateway.IntentGuildMembers,
				gateway.IntentGuildVoiceStates,
			),
		),
		bot.WithEventListenerFunc(command.Listener))
	if err != nil {
		log.Panicln("error creating client:", err)
	}
	client.Rest().SetGuildCommands(client.ApplicationID(), 630848078181826580, command.Commands) //TEST GUILD ID
	//client.Rest().SetGlobalCommands(client.ApplicationID(), command.Commands)
	defer client.Close(mainContext)
	if err = client.OpenGateway(mainContext); err != nil {
		log.Panicln("error opening gateway:", err)
	}
	slog.Info("Bot is running")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutdown signal received, shutting down client")
}
