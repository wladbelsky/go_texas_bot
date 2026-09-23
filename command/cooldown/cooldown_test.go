package cooldown

import (
	"testing"
	"time"
)

func TestReserve_FirstUseIsFree(t *testing.T) {
	c := New(time.Hour)
	if remaining := c.Reserve("g", "u"); remaining != 0 {
		t.Fatalf("Reserve() = %v, want 0 for a first use", remaining)
	}
}

func TestReserve_SecondUseIsBlocked(t *testing.T) {
	c := New(time.Hour)
	c.Reserve("g", "u")
	if remaining := c.Reserve("g", "u"); remaining <= 0 || remaining > time.Hour {
		t.Fatalf("second Reserve() = %v, want a remaining duration in (0, 1h]", remaining)
	}
}

func TestReserve_IsPerGuildAndUser(t *testing.T) {
	c := New(time.Hour)
	c.Reserve("g1", "u")
	if remaining := c.Reserve("g2", "u"); remaining != 0 {
		t.Fatalf("Reserve() in another guild = %v, want 0", remaining)
	}
	if remaining := c.Reserve("g1", "other"); remaining != 0 {
		t.Fatalf("Reserve() for another user = %v, want 0", remaining)
	}
}

func TestRelease_RefundsTheCooldown(t *testing.T) {
	c := New(time.Hour)
	c.Reserve("g", "u")
	c.Release("g", "u")
	if remaining := c.Reserve("g", "u"); remaining != 0 {
		t.Fatalf("Reserve() after Release() = %v, want 0", remaining)
	}
}

func TestReserve_ExpiresAfterDuration(t *testing.T) {
	c := New(time.Millisecond)
	c.Reserve("g", "u")
	time.Sleep(5 * time.Millisecond)
	if remaining := c.Reserve("g", "u"); remaining != 0 {
		t.Fatalf("Reserve() after the cooldown elapsed = %v, want 0", remaining)
	}
}
