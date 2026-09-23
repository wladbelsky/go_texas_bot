package guildsettings

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

func TestGet_CreatesDefaultRow(t *testing.T) {
	openTestDB(t)

	settings, err := Get("guild-1")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if settings.NSFWEnabled {
		t.Fatal("expected NSFWEnabled to default to false")
	}
}

func TestGet_ReturnsExistingRow(t *testing.T) {
	openTestDB(t)

	first, err := Get("guild-1")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	first.NSFWEnabled = true
	if err = db.DB.Save(&first).Error; err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	again, err := Get("guild-1")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if !again.NSFWEnabled {
		t.Fatal("expected the previously saved NSFWEnabled=true to persist")
	}
}
