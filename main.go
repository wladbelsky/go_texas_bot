package main

import (
	"context"
	"flag"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"
	"go_texas_bot/command"
	"go_texas_bot/config"
	"go_texas_bot/db"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	flag.Parse()
	token := config.Token()
	if token == "" {
		log.Panicln("token is required")
	}

	if err := db.Init(config.DBPath()); err != nil {
		log.Panicln("error opening database:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("error closing database", "err", err)
		}
	}()

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

	if _, err = client.Rest.SetGlobalCommands(client.ApplicationID, command.Commands); err != nil {
		log.Panicln("error setting global commands:", err)
	}
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
