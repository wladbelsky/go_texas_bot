package ger

import (
	"path/filepath"
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

func TestPhraseLists_AreNonEmpty(t *testing.T) {
	if len(phrases) == 0 {
		t.Error("expected phrases to be non-empty")
	}
	if len(selfPhrases) == 0 {
		t.Error("expected selfPhrases to be non-empty")
	}
}

func TestBumpUserCounter_CreatesAndIncrements(t *testing.T) {
	openTestDB(t)
	userID := snowflake.ID(42)

	if err := bumpUserCounter(userID, func(c *db.UserGerCounter) { c.Uses++ }); err != nil {
		t.Fatalf("bumpUserCounter() failed: %v", err)
	}
	if err := bumpUserCounter(userID, func(c *db.UserGerCounter) { c.Uses++ }); err != nil {
		t.Fatalf("bumpUserCounter() failed: %v", err)
	}

	var counter db.UserGerCounter
	if err := db.DB.First(&counter, "user_id = ?", userID.String()).Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if counter.Uses != 2 {
		t.Fatalf("Uses = %d, want 2", counter.Uses)
	}
}

func TestBumpUserCounter_IndependentFields(t *testing.T) {
	openTestDB(t)
	userID := snowflake.ID(7)

	if err := bumpUserCounter(userID, func(c *db.UserGerCounter) { c.Hits++ }); err != nil {
		t.Fatalf("bumpUserCounter(Hits) failed: %v", err)
	}
	if err := bumpUserCounter(userID, func(c *db.UserGerCounter) { c.Uses++ }); err != nil {
		t.Fatalf("bumpUserCounter(Uses) failed: %v", err)
	}

	var counter db.UserGerCounter
	if err := db.DB.First(&counter, "user_id = ?", userID.String()).Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if counter.Hits != 1 || counter.Uses != 1 {
		t.Fatalf("counter = %+v, want Hits=1 Uses=1", counter)
	}
}
