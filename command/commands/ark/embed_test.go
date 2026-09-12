package ark

import (
	"strings"
	"testing"

	"go_texas_bot/arknights"
)

func TestStars(t *testing.T) {
	if got := stars(3); got != "⭐⭐⭐" {
		t.Fatalf("stars(3) = %q, want 3 stars", got)
	}
	if got := stars(0); got != "" {
		t.Fatalf("stars(0) = %q, want empty string", got)
	}
}

func TestPercent(t *testing.T) {
	if got := percent(1, 2); got != 50 {
		t.Fatalf("percent(1,2) = %v, want 50", got)
	}
	if got := percent(0, 0); got != 0 {
		t.Fatalf("percent(0,0) = %v, want 0 (no division by zero)", got)
	}
}

func TestCharacterEmbed_UsesCharacterFields(t *testing.T) {
	c := arknights.Character{
		ID:         "char_test",
		Name:       "Test Op",
		Rarity:     5,
		Profession: "CASTER",
		Position:   "RANGED",
		Tags:       []string{"DPS", "Nuker"},
		ItemUsage:  "usage text",
		ItemDesc:   "desc text",
		Traits:     "Deals <@ba.kw>Arts damage</>",
	}

	embed := characterEmbed(c, "Someone")

	if embed.Title != "Test_Op" {
		t.Fatalf("Title = %q, want %q", embed.Title, "Test_Op")
	}
	if embed.Description != "⭐⭐⭐⭐⭐" {
		t.Fatalf("Description = %q, want 5 stars", embed.Description)
	}
	if embed.Footer == nil || embed.Footer.Text != "Requested by Someone" {
		t.Fatalf("Footer = %+v, want \"Requested by Someone\"", embed.Footer)
	}

	foundTraits := false
	for _, f := range embed.Fields {
		if f.Name == "Traits" {
			foundTraits = true
			if strings.Contains(f.Value, "<@") {
				t.Fatalf("Traits field still contains markup: %q", f.Value)
			}
		}
	}
	if !foundTraits {
		t.Fatal("expected a Traits field in the embed")
	}
}

func TestCollectionEmbed_EmptyCollection(t *testing.T) {
	embed := collectionEmbed("Someone", 0, 10, nil)
	if len(embed.Fields) != 1 || embed.Fields[0].Name != "Ты бомж" {
		t.Fatalf("expected the empty-collection field, got %+v", embed.Fields)
	}
}

func TestCollectionEmbed_ListsRaritiesHighToLow(t *testing.T) {
	byRarity := map[int][]collectionRow{
		3: {{Name: "Low", Count: 2}},
		6: {{Name: "High", Count: 1}},
	}
	embed := collectionEmbed("Someone", 2, 10, byRarity)

	if len(embed.Fields) != 2 {
		t.Fatalf("expected 2 rarity fields, got %d", len(embed.Fields))
	}
	if !strings.Contains(embed.Fields[0].Value, "High") {
		t.Fatalf("expected the 6-star field first, got %+v", embed.Fields[0])
	}
	if !strings.Contains(embed.Fields[1].Value, "Low") {
		t.Fatalf("expected the 3-star field second, got %+v", embed.Fields[1])
	}
}
