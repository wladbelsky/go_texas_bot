package arknights

import (
	"path/filepath"
	"testing"

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

func TestRoll_AddsToCollectionAndTracksPity(t *testing.T) {
	openTestDB(t)

	character, err := Roll("user-1")
	if err != nil {
		t.Fatalf("Roll() failed: %v", err)
	}

	entry, owned, err := OwnedEntry("user-1", character)
	if err != nil {
		t.Fatalf("OwnedEntry() failed: %v", err)
	}
	if !owned {
		t.Fatal("expected the rolled character to be in the user's collection")
	}
	if entry.Count != 1 {
		t.Fatalf("Count = %d, want 1", entry.Count)
	}

	pity, err := getPity("user-1")
	if err != nil {
		t.Fatalf("getPity() failed: %v", err)
	}
	if character.Rarity == 6 {
		if pity.SixMiss != 0 {
			t.Fatalf("SixMiss = %d, want 0 after a six-star roll", pity.SixMiss)
		}
	} else if pity.SixMiss != 1 {
		t.Fatalf("SixMiss = %d, want 1 after a non-six-star roll", pity.SixMiss)
	}
}

func TestRoll_IncrementsCountOnRepeat(t *testing.T) {
	openTestDB(t)

	character, ok := RandomOfRarity(5)
	if !ok {
		t.Fatal("expected at least one 5-star character")
	}

	for i := 0; i < 3; i++ {
		if err := addToCollection("user-1", character); err != nil {
			t.Fatalf("addToCollection() failed: %v", err)
		}
	}

	entry, owned, err := OwnedEntry("user-1", character)
	if err != nil || !owned {
		t.Fatalf("OwnedEntry() = (%v, %v, %v)", entry, owned, err)
	}
	if entry.Count != 3 {
		t.Fatalf("Count = %d, want 3", entry.Count)
	}
}

func TestOwnedEntry_NotOwned(t *testing.T) {
	openTestDB(t)

	character, ok := RandomOfRarity(3)
	if !ok {
		t.Fatal("expected at least one 3-star character")
	}

	_, owned, err := OwnedEntry("nobody", character)
	if err != nil {
		t.Fatalf("OwnedEntry() failed: %v", err)
	}
	if owned {
		t.Fatal("expected OwnedEntry to report the character as not owned")
	}
}

func TestCollection_GroupsByRarity(t *testing.T) {
	openTestDB(t)

	five, _ := RandomOfRarity(5)
	six, _ := RandomOfRarity(6)
	if err := addToCollection("user-1", five); err != nil {
		t.Fatal(err)
	}
	if err := addToCollection("user-1", six); err != nil {
		t.Fatal(err)
	}

	collection, err := Collection("user-1")
	if err != nil {
		t.Fatalf("Collection() failed: %v", err)
	}
	if len(collection[5]) != 1 || len(collection[6]) != 1 {
		t.Fatalf("Collection() = %v, want one entry each in rarities 5 and 6", collection)
	}

	owned, err := OwnedCount("user-1")
	if err != nil {
		t.Fatalf("OwnedCount() failed: %v", err)
	}
	if owned != 2 {
		t.Fatalf("OwnedCount() = %d, want 2", owned)
	}
}

func TestBarter_ExchangesDuplicatesForHigherRarity(t *testing.T) {
	openTestDB(t)

	fourStar, ok := RandomOfRarity(4)
	if !ok {
		t.Fatal("expected at least one 4-star character")
	}
	for i := 0; i < 6; i++ {
		if err := addToCollection("user-1", fourStar); err != nil {
			t.Fatal(err)
		}
	}

	results, err := Barter("user-1")
	if err != nil {
		t.Fatalf("Barter() failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected Barter to grant at least one exchanged character")
	}
	for _, c := range results {
		if c.Rarity != fourStar.Rarity+1 {
			t.Fatalf("exchanged character rarity = %d, want %d", c.Rarity, fourStar.Rarity+1)
		}
	}

	entry, owned, err := OwnedEntry("user-1", fourStar)
	if err != nil || !owned {
		t.Fatalf("OwnedEntry() = (%v, %v, %v)", entry, owned, err)
	}
	if entry.Count >= 6 {
		t.Fatalf("Count = %d, want fewer than 6 after bartering away the duplicates", entry.Count)
	}
}

func TestBarter_NoDuplicatesToExchange(t *testing.T) {
	openTestDB(t)

	results, err := Barter("user-1")
	if err != nil {
		t.Fatalf("Barter() failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Barter() = %v, want no exchanges for a user with no duplicates", results)
	}
}

func TestTotalGranted_OnlyGrowsThroughBarter(t *testing.T) {
	openTestDB(t)

	fourStar, _ := RandomOfRarity(4)
	for i := 0; i < 6; i++ {
		if err := addToCollection("user-1", fourStar); err != nil {
			t.Fatal(err)
		}
	}
	before, err := TotalGranted()
	if err != nil {
		t.Fatalf("TotalGranted() failed: %v", err)
	}
	if before != 6 {
		t.Fatalf("TotalGranted() = %d, want 6", before)
	}

	granted, err := Barter("user-1")
	if err != nil {
		t.Fatalf("Barter() failed: %v", err)
	}
	after, err := TotalGranted()
	if err != nil {
		t.Fatalf("TotalGranted() failed: %v", err)
	}
	if after != before+len(granted) {
		t.Fatalf("TotalGranted() after barter = %d, want %d (+%d exchanged, duplicates given up don't count down)", after, before+len(granted), len(granted))
	}
}
