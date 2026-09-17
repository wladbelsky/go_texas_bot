package main

import (
	"context"
	"flag"
	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"go_texas_bot/command"
	musiccommands "go_texas_bot/command/commands/music"
	"go_texas_bot/config"
	"go_texas_bot/db"
	"go_texas_bot/music"
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
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagVoiceStates|cache.FlagMembers)),
		bot.WithEventListenerFunc(command.Listener),
		bot.WithEventListenerFunc(musiccommands.ComponentListener),
		bot.WithEventListenerFunc(onVoiceStateUpdate),
		bot.WithEventListenerFunc(onVoiceServerUpdate),
	)
	if err != nil {
		log.Panicln("error creating client:", err)
	}

	if err = music.Init(mainContext, client.ApplicationID, config.LavalinkHost(), config.LavalinkPort(), config.LavalinkPassword()); err != nil {
		slog.Error("error connecting to lavalink node, music commands will be unavailable", "err", err)
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

// onVoiceStateUpdate forwards the bot's own voice state changes to
// disgolink, which needs them to keep its players' voice connections in
// sync, and drops the guild's music queue once the bot leaves the channel.
func onVoiceStateUpdate(event *events.GuildVoiceStateUpdate) {
	if music.Lavalink == nil || event.VoiceState.UserID != event.Client().ApplicationID {
		return
	}
	music.Lavalink.OnVoiceStateUpdate(context.Background(), event.VoiceState.GuildID, event.VoiceState.ChannelID, event.VoiceState.SessionID)
	if event.VoiceState.ChannelID == nil {
		music.Queues.Delete(event.VoiceState.GuildID)
	}
}

// onVoiceServerUpdate forwards voice server updates to disgolink.
func onVoiceServerUpdate(event *events.VoiceServerUpdate) {
	if music.Lavalink == nil || event.Endpoint == nil {
		return
	}
	music.Lavalink.OnVoiceServerUpdate(context.Background(), event.GuildID, event.Token, *event.Endpoint)
}
