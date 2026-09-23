package cooldown

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

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

func mustReserve(t *testing.T, c *Tracker, guildID, userID string) time.Duration {
	t.Helper()
	remaining, err := c.Reserve(guildID, userID)
	if err != nil {
		t.Fatalf("Reserve(%q, %q) failed: %v", guildID, userID, err)
	}
	return remaining
}

func TestReserve_FirstUseIsFree(t *testing.T) {
	openTestDB(t)
	c := New("test", time.Hour)
	if remaining := mustReserve(t, c, "g", "u"); remaining != 0 {
		t.Fatalf("Reserve() = %v, want 0 for a first use", remaining)
	}
}

func TestReserve_SecondUseIsBlocked(t *testing.T) {
	openTestDB(t)
	c := New("test", time.Hour)
	mustReserve(t, c, "g", "u")
	if remaining := mustReserve(t, c, "g", "u"); remaining <= 0 || remaining > time.Hour {
		t.Fatalf("second Reserve() = %v, want a remaining duration in (0, 1h]", remaining)
	}
}

func TestReserve_IsPerGuildUserAndCommand(t *testing.T) {
	openTestDB(t)
	c := New("test", time.Hour)
	mustReserve(t, c, "g1", "u")
	if remaining := mustReserve(t, c, "g2", "u"); remaining != 0 {
		t.Fatalf("Reserve() in another guild = %v, want 0", remaining)
	}
	if remaining := mustReserve(t, c, "g1", "other"); remaining != 0 {
		t.Fatalf("Reserve() for another user = %v, want 0", remaining)
	}
	if remaining := mustReserve(t, New("other-command", time.Hour), "g1", "u"); remaining != 0 {
		t.Fatalf("Reserve() for another command = %v, want 0", remaining)
	}
}

func TestRelease_RefundsTheCooldown(t *testing.T) {
	openTestDB(t)
	c := New("test", time.Hour)
	mustReserve(t, c, "g", "u")
	c.Release("g", "u")
	if remaining := mustReserve(t, c, "g", "u"); remaining != 0 {
		t.Fatalf("Reserve() after Release() = %v, want 0", remaining)
	}
}

func TestReserve_ExpiresAfterDuration(t *testing.T) {
	openTestDB(t)
	c := New("test", 5*time.Millisecond)
	mustReserve(t, c, "g", "u")
	time.Sleep(20 * time.Millisecond)
	if remaining := mustReserve(t, c, "g", "u"); remaining != 0 {
		t.Fatalf("Reserve() after the cooldown elapsed = %v, want 0", remaining)
	}
	if remaining := mustReserve(t, c, "g", "u"); remaining <= 0 {
		t.Fatalf("Reserve() right after re-reserving = %v, want the new cooldown to be running", remaining)
	}
}

// The point of storing cooldowns in the database: a new Tracker (as after a
// bot restart) still sees a cooldown started by the old one.
func TestReserve_SurvivesRestart(t *testing.T) {
	openTestDB(t)
	mustReserve(t, New("test", time.Hour), "g", "u")

	if remaining := mustReserve(t, New("test", time.Hour), "g", "u"); remaining <= 0 {
		t.Fatalf("Reserve() on a fresh Tracker = %v, want the stored cooldown to still be running", remaining)
	}
}

func TestReserve_ConcurrentCallsOnlyOneWins(t *testing.T) {
	openTestDB(t)
	c := New("test", time.Hour)

	const callers = 10
	var wg sync.WaitGroup
	results := make(chan time.Duration, callers)
	for i := 0; i < callers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			remaining, err := c.Reserve("g", "u")
			if err != nil {
				t.Errorf("Reserve() failed: %v", err)
				return
			}
			results <- remaining
		}()
	}
	wg.Wait()
	close(results)

	free := 0
	for remaining := range results {
		if remaining == 0 {
			free++
		}
	}
	if free != 1 {
		t.Fatalf("%d concurrent Reserve() calls got through, want exactly 1", free)
	}
}
