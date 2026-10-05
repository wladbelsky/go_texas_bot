package db

import (
	"os"
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	if err := Init(path); err != nil {
		t.Fatalf("Init(%q) failed: %v", path, err)
	}
	t.Cleanup(func() {
		if err := Close(); err != nil {
			t.Errorf("Close() failed: %v", err)
		}
	})
}

func TestInit_CreatesFileAndMigrates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "test.db")

	if err := Init(path); err != nil {
		t.Fatalf("Init(%q) failed: %v", path, err)
	}
	defer Close()

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected db file to be created: %v", err)
	}
	for _, model := range []any{&ArkCollectionEntry{}, &ArkPity{}, &GuildSettings{}} {
		if !DB.Migrator().HasTable(model) {
			t.Fatalf("expected table for %T to be migrated", model)
		}
	}
}

func TestArkCollectionEntry_CRUD(t *testing.T) {
	openTestDB(t)

	entry := ArkCollectionEntry{UserID: "1", OperatorName: "Amiya", Rarity: 5, Count: 1}
	if err := DB.Create(&entry).Error; err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var got ArkCollectionEntry
	if err := DB.First(&got, "user_id = ? AND operator_name = ?", "1", "Amiya").Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if got.Count != 1 || got.Rarity != 5 {
		t.Fatalf("unexpected entry: %+v", got)
	}

	got.Count++
	if err := DB.Save(&got).Error; err != nil {
		t.Fatalf("update failed: %v", err)
	}

	var updated ArkCollectionEntry
	if err := DB.First(&updated, "user_id = ? AND operator_name = ?", "1", "Amiya").Error; err != nil {
		t.Fatalf("read back after update failed: %v", err)
	}
	if updated.Count != 2 {
		t.Fatalf("Count = %d, want 2", updated.Count)
	}
}

func TestGuildSettings_DefaultsToZeroValue(t *testing.T) {
	openTestDB(t)

	var settings GuildSettings
	err := DB.First(&settings, "guild_id = ?", "unknown-guild").Error
	if err == nil {
		t.Fatal("expected an error for a guild with no settings row")
	}
}

func TestClose_NilDBIsNoop(t *testing.T) {
	DB = nil
	if err := Close(); err != nil {
		t.Fatalf("Close() on nil DB should be a no-op, got: %v", err)
	}
}
