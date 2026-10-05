package music

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
	musicpkg "go_texas_bot/music"
)

const embedColor = 0xff9900 // matches the original bot's config.json embed_color

func formatDuration(d lavalink.Duration) string {
	return fmt.Sprintf("%d:%02d", d.Minutes(), d.SecondsPart())
}

const progressBarLength = 20

func progressBar(position, length lavalink.Duration) string {
	if length <= 0 {
		return ""
	}
	filled := int(int64(position) * progressBarLength / int64(length))
	if filled > progressBarLength {
		filled = progressBarLength
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("🟦", filled) + strings.Repeat("⬛", progressBarLength-filled)
}

func playStateEmoji(player disgolink.Player) string {
	switch {
	case player == nil || player.Track() == nil:
		return "⏹️"
	case player.Paused():
		return "⏸️"
	default:
		return "▶️"
	}
}

// nowPlayingEmbed mirrors the original bot's PlayerControls.generate_player_embed.
func nowPlayingEmbed(player disgolink.Player, queue *musicpkg.Queue) discord.Embed {
	embed := discord.Embed{
		Title: playStateEmoji(player) + " Сейчас играет",
		Color: embedColor,
	}

	track := trackOrNil(player)
	if track == nil {
		embed.Fields = []discord.EmbedField{
			blockField("Название трека", "Очередь пуста, добавь музыку командой /play"),
		}
		return embed
	}

	embed.Fields = []discord.EmbedField{
		blockField("Название трека", track.Info.Title),
		blockField("Исполнитель", track.Info.Author),
		blockField("Время",
			fmt.Sprintf("%s/%s\n%s", formatDuration(player.Position()), formatDuration(track.Info.Length),
				progressBar(player.Position(), track.Info.Length)),
		),
	}
	if requesterID, ok := musicpkg.Requester(*track); ok {
		embed.Footer = &discord.EmbedFooter{Text: "Запросил " + discord.UserMention(mustParseSnowflake(requesterID))}
	}
	return embed
}

func trackOrNil(player disgolink.Player) *lavalink.Track {
	if player == nil {
		return nil
	}
	return player.Track()
}

// queueEmbed lists the upcoming tracks, capped at `show`.
func queueEmbed(queue *musicpkg.Queue, show int) discord.Embed {
	embed := discord.Embed{
		Title: "Очередь",
		Color: embedColor,
	}

	current, hasCurrent := queue.Current()
	nowPlaying := "Ничего сейчас не играет(\nЗапроси трек командой /play"
	if hasCurrent {
		nowPlaying = fmt.Sprintf("(%d/%d) %s", queue.Position()+1, queue.Length(), current.Info.Title)
	}
	embed.Fields = []discord.EmbedField{
		blockField("Сейчас играет", nowPlaying),
		blockField("Всего треков", fmt.Sprintf("%d", queue.Length())),
	}

	upcoming := queue.Upcoming()
	if len(upcoming) > show {
		upcoming = upcoming[:show]
	}
	if len(upcoming) > 0 {
		lines := make([]string, len(upcoming))
		for i, t := range upcoming {
			lines[i] = fmt.Sprintf("%d) %s", queue.Position()+2+i, t.Info.Title)
		}
		embed.Fields = append(embed.Fields, blockField(fmt.Sprintf("Следующие %d треков", len(upcoming)), strings.Join(lines, "\n")))
	}
	return embed
}

func blockField(name, value string) discord.EmbedField {
	if value == "" {
		value = "—"
	}
	return discord.EmbedField{Name: name, Value: value}
}

const (
	customIDPrev    = "music:prev"
	customIDPause   = "music:pause"
	customIDStop    = "music:stop"
	customIDNext    = "music:next"
	customIDRefresh = "music:refresh"
)

func controlsComponents() []discord.LayoutComponent {
	return []discord.LayoutComponent{
		discord.NewActionRow(
			discord.NewSecondaryButton("", customIDPrev).WithEmoji(discord.ComponentEmoji{Name: "⏮️"}),
			discord.NewSecondaryButton("", customIDPause).WithEmoji(discord.ComponentEmoji{Name: "⏯️"}),
			discord.NewSecondaryButton("", customIDStop).WithEmoji(discord.ComponentEmoji{Name: "⏹️"}),
			discord.NewSecondaryButton("", customIDNext).WithEmoji(discord.ComponentEmoji{Name: "⏭️"}),
			discord.NewSecondaryButton("", customIDRefresh).WithEmoji(discord.ComponentEmoji{Name: "🔄"}),
		),
	}
}
