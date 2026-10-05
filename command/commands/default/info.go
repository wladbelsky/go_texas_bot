package defaultcmd

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/arknights"
	"go_texas_bot/db"
)

func infoCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	embed, err := infoEmbed(event.User().EffectiveName())
	if err != nil {
		return err
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed).WithEphemeral(true))
}

func infoEmbed(requestedBy string) (discord.Embed, error) {
	arkTotal, err := arknights.TotalGranted()
	if err != nil {
		return discord.Embed{}, err
	}
	topUserID, topSixCount, err := topSixStarCollector()
	if err != nil {
		return discord.Embed{}, err
	}
	gerStats, err := gerGlobalStats()
	if err != nil {
		return discord.Embed{}, err
	}
	topByUses, err := topGerCounters("uses")
	if err != nil {
		return discord.Embed{}, err
	}
	topByHits, err := topGerCounters("hits")
	if err != nil {
		return discord.Embed{}, err
	}

	topSixLine := "Пока никто ничего не собрал"
	if topUserID != "" {
		topSixLine = fmt.Sprintf(
			"%s с количеством аж %d шестизведочных персонажей. Поздравляем Вас и вручаем вам самый ценный подарок: **наше увожение**",
			discord.UserMention(mustParseSnowflake(topUserID)), topSixCount,
		)
	}

	return discord.Embed{
		Title: "Texass",
		Color: embedColor,
		Fields: []discord.EmbedField{
			blockField("Описание", "Тупая деффка еще и бот"),
			blockField("Статистика арков", fmt.Sprintf("Арков выкручено за все время: %d", arkTotal)),
			blockField("Больше всего 6★ собрано", topSixLine),
			blockField("Статистика пуков", fmt.Sprintf(
				"Пуков за все время: %d\nИз них самообсеров: %d\nПопаданий по ботам: %d\nПопаданий по мне: %d",
				gerStats.Total, gerStats.Self, gerStats.Bot, gerStats.Me,
			)),
			blockField("Топ засранцев", formatGerLeaderboard(topByUses, func(c db.UserGerCounter) int { return c.Uses })),
			blockField("Обосрали больше всего", formatGerLeaderboard(topByHits, func(c db.UserGerCounter) int { return c.Hits })),
		},
		Thumbnail: &discord.EmbedResource{URL: "https://aceship.github.io/AN-EN-Tags/img/factions/logo_rhodes.png"},
		Image:     &discord.EmbedResource{URL: "https://aceship.github.io/AN-EN-Tags/img/characters/char_102_texas_2.png"},
		Footer:    &discord.EmbedFooter{Text: "Requested by " + requestedBy},
	}, nil
}

func topSixStarCollector() (userID string, count int, err error) {
	var row struct {
		UserID string
		Total  int
	}
	err = db.DB.Model(&db.ArkCollectionEntry{}).
		Select("user_id, SUM(count) as total").
		Where("rarity = ?", 6).
		Group("user_id").
		Order("total DESC").
		Limit(1).
		Scan(&row).Error
	return row.UserID, row.Total, err
}

func gerGlobalStats() (db.GerStats, error) {
	var stats db.GerStats
	err := db.DB.FirstOrCreate(&stats, db.GerStats{ID: 1}).Error
	return stats, err
}

func topGerCounters(orderColumn string) ([]db.UserGerCounter, error) {
	var counters []db.UserGerCounter
	err := db.DB.Order(orderColumn + " DESC").Limit(5).Find(&counters).Error
	return counters, err
}

func formatGerLeaderboard(counters []db.UserGerCounter, value func(db.UserGerCounter) int) string {
	if len(counters) == 0 {
		return "Статистика не найдена"
	}
	lines := make([]string, 0, len(counters))
	for _, c := range counters {
		if v := value(c); v > 0 {
			lines = append(lines, fmt.Sprintf("%s: %d", discord.UserMention(mustParseSnowflake(c.UserID)), v))
		}
	}
	if len(lines) == 0 {
		return "Статистика не найдена"
	}
	return strings.Join(lines, "\n")
}

func blockField(name, value string) discord.EmbedField {
	if value == "" {
		value = "—"
	}
	return discord.EmbedField{Name: name, Value: value}
}
