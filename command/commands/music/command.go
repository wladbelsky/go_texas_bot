// Package music implements the /play, /skip, /stop, /queue, /disconnect,
// /shuffle, /repeat, /skipto, /pop and /controls slash commands: the Go
// port of the original Python bot's music.py cog, played through Lavalink
// via disgolink.
package music

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"go_texas_bot/command/command_selector"
	musicpkg "go_texas_bot/music"
)

func init() {
	command_selector.CommandSelector.AddCommand("play", playCommandListener)
	command_selector.CommandSelector.AddCommand("skip", skipCommandListener)
	command_selector.CommandSelector.AddCommand("stop", stopCommandListener)
	command_selector.CommandSelector.AddCommand("queue", queueCommandListener)
	command_selector.CommandSelector.AddCommand("disconnect", disconnectCommandListener)
	command_selector.CommandSelector.AddCommand("shuffle", shuffleCommandListener)
	command_selector.CommandSelector.AddCommand("repeat", repeatCommandListener)
	command_selector.CommandSelector.AddCommand("skipto", skiptoCommandListener)
	command_selector.CommandSelector.AddCommand("pop", popCommandListener)
	command_selector.CommandSelector.AddCommand("controls", controlsCommandListener)
}

func respond(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().WithContent(content))
}

func respondEphemeral(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().WithContent(content).WithEphemeral(true))
}

// playCommandListener resolves the query/URL via Lavalink (using whatever
// sources are configured on the node -- YouTube, and Spotify/Apple
// Music/Deezer through the LavaSrc plugin), joins the caller's voice
// channel if needed, and queues the result.
func playCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	if musicpkg.Lavalink == nil {
		return respondEphemeral(event, errLavalinkDown.Error())
	}

	voiceState, ok := event.Client().Caches.VoiceState(guildID, event.User().ID)
	if !ok || voiceState.ChannelID == nil {
		return respondEphemeral(event, "Сначала зайди в голосовой канал")
	}

	if err = event.DeferCreateMessage(false); err != nil {
		return err
	}

	query := event.SlashCommandInteractionData().String("query")
	identifier := query
	if !isURL(query) {
		identifier = lavalink.SearchTypeYouTube.Apply(query)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	requesterID := event.User().ID
	queue := musicpkg.Queues.Get(guildID)

	var toPlay *lavalink.Track
	var alreadyQueued bool
	var loadedMessage string
	musicpkg.Lavalink.BestNode().LoadTracksHandler(ctx, identifier, disgolink.NewResultHandler(
		func(t lavalink.Track) {
			t = musicpkg.WithRequester(t, requesterID)
			toPlay = &t
			loadedMessage = fmt.Sprintf("Добавила **%s** в очередь", t.Info.Title)
		},
		func(playlist lavalink.Playlist) {
			for i := range playlist.Tracks {
				playlist.Tracks[i] = musicpkg.WithRequester(playlist.Tracks[i], requesterID)
			}
			queue.Add(playlist.Tracks...)
			alreadyQueued = true
			if len(playlist.Tracks) > 0 {
				toPlay = &playlist.Tracks[0]
			}
			loadedMessage = fmt.Sprintf("Добавила %d треков из плейлиста **%s** в очередь", len(playlist.Tracks), playlist.Info.Name)
		},
		func(tracks []lavalink.Track) {
			if len(tracks) == 0 {
				return
			}
			t := musicpkg.WithRequester(tracks[0], requesterID)
			toPlay = &t
			loadedMessage = fmt.Sprintf("Добавила **%s** в очередь", t.Info.Title)
		},
		func() {
			loadedMessage = fmt.Sprintf("Ничего не найдено по запросу `%s`", query)
		},
		func(loadErr error) {
			loadedMessage = fmt.Sprintf("Ошибка при поиске: `%s`", loadErr)
		},
	))

	if _, err = event.Client().Rest.UpdateInteractionResponse(event.ApplicationID(), event.Token(), discord.NewMessageUpdate().WithContent(loadedMessage)); err != nil {
		return err
	}
	if toPlay == nil {
		return nil
	}
	if !alreadyQueued {
		queue.Add(*toPlay)
	}

	player := musicpkg.Lavalink.Player(guildID)
	alreadyPlaying := player.Track() != nil

	if err = event.Client().UpdateVoiceState(context.Background(), guildID, voiceState.ChannelID, false, false); err != nil {
		return err
	}

	if !alreadyPlaying {
		current, hasCurrent := queue.Advance()
		if hasCurrent {
			return player.Update(context.Background(), lavalink.WithTrack(current))
		}
	}
	return nil
}

func skipCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	player := existingPlayer(guildID)
	if player == nil || player.Track() == nil {
		return respondEphemeral(event, errNothingPlaying.Error())
	}
	if err = requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *player.Track()); err != nil {
		return respondEphemeral(event, err.Error())
	}

	queue := musicpkg.Queues.Get(guildID)
	next, ok := queue.Advance()
	if !ok {
		if err = player.Update(context.Background(), lavalink.WithNullTrack()); err != nil {
			return err
		}
		return respond(event, "Играть больше нечего")
	}
	if err = player.Update(context.Background(), lavalink.WithTrack(next)); err != nil {
		return err
	}
	return respond(event, "Играем дальше: **"+next.Info.Title+"**")
}

func stopCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	player := existingPlayer(guildID)
	if player == nil {
		return respondEphemeral(event, errNothingPlaying.Error())
	}
	if track := player.Track(); track != nil {
		if err = requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *track); err != nil {
			return respondEphemeral(event, err.Error())
		}
	}

	musicpkg.Queues.Get(guildID).Clear()
	if err = player.Update(context.Background(), lavalink.WithNullTrack()); err != nil {
		return err
	}
	return respond(event, "Остановила музыку и очистила очередь")
}

func queueCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	queue := musicpkg.Queues.Get(guildID)
	if queue.IsEmpty() {
		return respondEphemeral(event, "Очередь пуста")
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(queueEmbed(queue, 10)))
}

func disconnectCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	player := existingPlayer(guildID)
	if player == nil {
		return respondEphemeral(event, "Я и не в канале")
	}
	if track := player.Track(); track != nil {
		if err = requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *track); err != nil {
			return respondEphemeral(event, err.Error())
		}
	}

	musicpkg.Queues.Delete(guildID)
	if err = event.Client().UpdateVoiceState(context.Background(), guildID, nil, false, false); err != nil {
		return err
	}
	return respond(event, "Ну все, до новых встреч, пока!")
}

func shuffleCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	queue := musicpkg.Queues.Get(guildID)
	if queue.IsEmpty() {
		return respondEphemeral(event, errNothingPlaying.Error())
	}
	queue.Shuffle()
	return respond(event, "Перемешала очередь")
}

func repeatCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	mode := musicpkg.RepeatMode(event.SlashCommandInteractionData().String("mode"))
	musicpkg.Queues.Get(guildID).SetRepeatMode(mode)
	return respond(event, "Установила режим повтора: `"+string(mode)+"`")
}

func skiptoCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	player := existingPlayer(guildID)
	if player == nil {
		return respondEphemeral(event, errNothingPlaying.Error())
	}
	if track := player.Track(); track != nil {
		if err = requireTrackOwnerOrAdmin(event.Member(), event.User().ID, *track); err != nil {
			return respondEphemeral(event, err.Error())
		}
	}

	index := event.SlashCommandInteractionData().Int("index") - 1
	track, ok := musicpkg.Queues.Get(guildID).SkipTo(index)
	if !ok {
		return respondEphemeral(event, "Указан неверный номер")
	}
	if err = player.Update(context.Background(), lavalink.WithTrack(track)); err != nil {
		return err
	}
	return respond(event, "Играем музяку под номером "+event.SlashCommandInteractionData().String("index"))
}

func popCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	guildID, err := requireGuild(event)
	if err != nil {
		return respondEphemeral(event, err.Error())
	}
	index := event.SlashCommandInteractionData().Int("index") - 1
	track, ok := musicpkg.Queues.Get(guildID).RemoveAt(index)
	if !ok {
		return respondEphemeral(event, "Указан неверный номер (или это трек, который играет сейчас)")
	}
	return respond(event, "Удалила из очереди: **"+track.Info.Title+"**")
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}
