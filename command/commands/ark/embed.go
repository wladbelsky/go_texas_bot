package ark

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"go_texas_bot/arknights"
)

const embedColor = 0xff9900 // matches the original bot's config.json embed_color

func inlineField(name, value string) discord.EmbedField {
	inline := true
	return discord.EmbedField{Name: name, Value: value, Inline: &inline}
}

func blockField(name, value string) discord.EmbedField {
	return discord.EmbedField{Name: name, Value: value}
}

func stars(rarity int) string {
	return strings.Repeat("⭐", rarity)
}

// characterEmbed renders a single rolled/looked-up character, mirroring the
// original bot's ark_embed_and_view.
func characterEmbed(character arknights.Character, requestedBy string) discord.Embed {
	skins := arknights.SkinsFor(character.ID)
	skinNames := make([]string, 0, len(skins))
	for _, s := range skins {
		skinNames = append(skinNames, s.DisplayName())
	}
	skinsValue := "—"
	if len(skinNames) > 0 {
		skinsValue = strings.Join(skinNames, ", ")
	}

	return discord.Embed{
		Title:       character.DisplayName(),
		Description: stars(character.Rarity),
		URL:         fmt.Sprintf("https://aceship.github.io/AN-EN-Tags/akhrchars.html?opname=%s", character.DisplayName()),
		Color:       embedColor,
		Fields: []discord.EmbedField{
			blockField("Description", character.ItemUsage+"\n"+character.ItemDesc),
			inlineField("Position", character.Position),
			inlineField("Profession", character.Profession),
			inlineField("Tags", strings.Join(character.Tags, ", ")),
			blockField("Traits", character.CleanTraits()),
			blockField("Skins", skinsValue),
		},
		Image: &discord.EmbedResource{
			URL: fmt.Sprintf("https://raw.githubusercontent.com/Aceship/Arknight-Images/main/characters/%s_1.png", character.ID),
		},
		Footer: &discord.EmbedFooter{Text: "Requested by " + requestedBy},
	}
}

// collectionEmbed renders a user's full collection, grouped by rarity, like
// the original /myark with no character argument.
func collectionEmbed(displayName string, owned int, total int, byRarity map[int][]collectionRow) discord.Embed {
	embed := discord.Embed{
		Title: fmt.Sprintf("%s's collection %d/%d (%.2f%%)", displayName, owned, total, percent(owned, total)),
		Color: embedColor,
	}

	if owned == 0 {
		embed.Fields = []discord.EmbedField{
			blockField("Ты бомж", "иди покрути девочек!"),
		}
		return embed
	}

	for rarity := 6; rarity >= 1; rarity-- {
		rows := byRarity[rarity]
		if len(rows) == 0 {
			continue
		}
		lines := make([]string, 0, len(rows))
		for _, r := range rows {
			lines = append(lines, fmt.Sprintf("%s x %d", r.Name, r.Count))
		}
		embed.Fields = append(embed.Fields, blockField(stars(rarity), strings.Join(lines, "\n")))
	}

	embed.Footer = &discord.EmbedFooter{Text: "Используй команду /myark <имя>, чтоб посмотреть на персонажа."}
	return embed
}

type collectionRow struct {
	Name  string
	Count int
}

func percent(owned, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(owned) / float64(total) * 100
}
