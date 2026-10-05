// Package arknights implements the gacha-style operator collection game
// ("case simulator") ported from the original Python bot's ark.py cog.
//
// Static game data (character_table.json / skin_table.json) is sourced from
// https://github.com/Aceship/AN-EN-Tags and trimmed down to only the fields
// this package needs; see data/characters.json and data/skins.json.
package arknights

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed data/characters.json data/skins.json
var dataFS embed.FS

// Character is a collectible Arknights operator (3-6 stars, obtainable
// through the case simulator).
type Character struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Rarity     int      `json:"rarity"`
	Profession string   `json:"profession"`
	Position   string   `json:"position"`
	Tags       []string `json:"tags"`
	ItemUsage  string   `json:"item_usage"`
	ItemDesc   string   `json:"item_desc"`
	Traits     string   `json:"traits"`
}

// DisplayName is the name as shown in embeds and stored in the database:
// spaces replaced with underscores and apostrophes removed, matching the
// original Python bot's normalization.
func (c Character) DisplayName() string {
	return normalizeName(c.Name)
}

// CleanTraits strips the Arknights rich-text markup (e.g. "<@ba.kw>...</>")
// from the trait description.
func (c Character) CleanTraits() string {
	s := strings.NewReplacer("<", "", "@", "", ".", "", ">", "", "/", "").Replace(c.Traits)
	return strings.ReplaceAll(s, "bakw", "")
}

func normalizeName(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	return strings.ReplaceAll(name, "'", "")
}

// Skin is a cosmetic skin/outfit available for a Character.
type Skin struct {
	CharID        string `json:"char_id"`
	PortraitID    string `json:"portrait_id"`
	SkinName      string `json:"skin_name"`
	ModelName     string `json:"model_name"`
	SkinGroupName string `json:"skin_group_name"`
}

// DisplayName mirrors the Python bot's fallback: the skin name if it has
// one, otherwise "<model name>#<portrait suffix>".
func (s Skin) DisplayName() string {
	if s.SkinName != "" {
		return s.SkinName
	}
	parts := strings.Split(s.PortraitID, "_")
	return fmt.Sprintf("%s#%s", s.ModelName, parts[len(parts)-1])
}

var (
	allCharacters      []Character
	charactersByName   map[string]Character
	charactersByRarity map[int][]Character
	skinsByCharID      map[string][]Skin
)

func init() {
	charactersJSON, err := dataFS.ReadFile("data/characters.json")
	if err != nil {
		panic("arknights: failed to read embedded characters.json: " + err.Error())
	}
	if err = json.Unmarshal(charactersJSON, &allCharacters); err != nil {
		panic("arknights: failed to parse embedded characters.json: " + err.Error())
	}

	skinsJSON, err := dataFS.ReadFile("data/skins.json")
	if err != nil {
		panic("arknights: failed to read embedded skins.json: " + err.Error())
	}
	var skins []Skin
	if err = json.Unmarshal(skinsJSON, &skins); err != nil {
		panic("arknights: failed to parse embedded skins.json: " + err.Error())
	}

	charactersByName = make(map[string]Character, len(allCharacters))
	charactersByRarity = make(map[int][]Character)
	for _, c := range allCharacters {
		charactersByName[strings.ToLower(c.DisplayName())] = c
		charactersByRarity[c.Rarity] = append(charactersByRarity[c.Rarity], c)
	}

	skinsByCharID = make(map[string][]Skin, len(skins))
	for _, s := range skins {
		skinsByCharID[s.CharID] = append(skinsByCharID[s.CharID], s)
	}
}

// All returns every collectible character.
func All() []Character { return allCharacters }

// Count returns the number of collectible characters.
func Count() int { return len(allCharacters) }

// ByRarity returns every character of the given star rarity (3-6).
func ByRarity(rarity int) []Character { return charactersByRarity[rarity] }

// ByName looks up a character by its display name (case-insensitive,
// same normalization as DisplayName).
func ByName(name string) (Character, bool) {
	c, ok := charactersByName[strings.ToLower(normalizeName(name))]
	return c, ok
}

// SkinsFor returns every skin registered for the given character id.
func SkinsFor(charID string) []Skin { return skinsByCharID[charID] }
