package defaultcmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/disgoorg/snowflake/v2"
	"go_texas_bot/db"
)

func openTestDB(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	if err := db.Init(path); err != nil {
		t.Fatalf("db.Init(%q) failed: %v", path, err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	})
}

func TestMustParseSnowflake_RoundTrip(t *testing.T) {
	id := snowflake.ID(123456789)
	if got := mustParseSnowflake(id.String()); got != id {
		t.Fatalf("mustParseSnowflake(%q) = %v, want %v", id.String(), got, id)
	}
}

func TestMustParseSnowflake_Malformed(t *testing.T) {
	if got := mustParseSnowflake("not-a-snowflake"); got != 0 {
		t.Fatalf("mustParseSnowflake(malformed) = %v, want 0", got)
	}
}

func TestFormatGerLeaderboard_Empty(t *testing.T) {
	got := formatGerLeaderboard(nil, func(c db.UserGerCounter) int { return c.Uses })
	if got != "Статистика не найдена" {
		t.Fatalf("formatGerLeaderboard(nil) = %q", got)
	}
}

func TestFormatGerLeaderboard_SkipsZeroValues(t *testing.T) {
	counters := []db.UserGerCounter{{UserID: "1", Uses: 0}, {UserID: "2", Uses: 5}}
	got := formatGerLeaderboard(counters, func(c db.UserGerCounter) int { return c.Uses })
	if strings.Count(got, "\n") != 0 {
		t.Fatalf("expected exactly one leaderboard line, got %q", got)
	}
	if !strings.Contains(got, "5") {
		t.Fatalf("expected the nonzero entry to be listed, got %q", got)
	}
}

func TestTotalArkRolls_SumsAcrossUsers(t *testing.T) {
	openTestDB(t)

	db.DB.Create(&db.ArkCollectionEntry{UserID: "1", OperatorName: "A", Rarity: 5, Count: 3})
	db.DB.Create(&db.ArkCollectionEntry{UserID: "2", OperatorName: "B", Rarity: 6, Count: 2})

	total, err := totalArkRolls()
	if err != nil {
		t.Fatalf("totalArkRolls() failed: %v", err)
	}
	if total != 5 {
		t.Fatalf("totalArkRolls() = %d, want 5", total)
	}
}

func TestTotalArkRolls_EmptyDB(t *testing.T) {
	openTestDB(t)

	total, err := totalArkRolls()
	if err != nil {
		t.Fatalf("totalArkRolls() failed: %v", err)
	}
	if total != 0 {
		t.Fatalf("totalArkRolls() = %d, want 0", total)
	}
}

func TestTopSixStarCollector(t *testing.T) {
	openTestDB(t)

	db.DB.Create(&db.ArkCollectionEntry{UserID: "1", OperatorName: "A", Rarity: 6, Count: 2})
	db.DB.Create(&db.ArkCollectionEntry{UserID: "2", OperatorName: "B", Rarity: 6, Count: 5})
	db.DB.Create(&db.ArkCollectionEntry{UserID: "2", OperatorName: "C", Rarity: 5, Count: 100}) // not rarity 6, shouldn't count

	userID, count, err := topSixStarCollector()
	if err != nil {
		t.Fatalf("topSixStarCollector() failed: %v", err)
	}
	if userID != "2" || count != 5 {
		t.Fatalf("topSixStarCollector() = (%q, %d), want (\"2\", 5)", userID, count)
	}
}

func TestInfoEmbed_BuildsWithoutError(t *testing.T) {
	openTestDB(t)

	embed, err := infoEmbed("Someone")
	if err != nil {
		t.Fatalf("infoEmbed() failed: %v", err)
	}
	if embed.Footer == nil || embed.Footer.Text != "Requested by Someone" {
		t.Fatalf("Footer = %+v, want \"Requested by Someone\"", embed.Footer)
	}
	if len(embed.Fields) == 0 {
		t.Fatal("expected infoEmbed to populate fields")
	}
}
