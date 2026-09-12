package ark

import "testing"

func TestCheckCooldown_FirstRollHasNoCooldown(t *testing.T) {
	guildID, userID := "guild-1", "user-checkcooldown-first"
	if remaining := checkCooldown(guildID, userID); remaining != 0 {
		t.Fatalf("checkCooldown() = %v, want 0 for a first roll", remaining)
	}
}

func TestCheckCooldown_SecondRollIsBlocked(t *testing.T) {
	guildID, userID := "guild-1", "user-checkcooldown-second"
	if remaining := checkCooldown(guildID, userID); remaining != 0 {
		t.Fatalf("first checkCooldown() = %v, want 0", remaining)
	}
	if remaining := checkCooldown(guildID, userID); remaining <= 0 {
		t.Fatalf("second checkCooldown() = %v, want a positive remaining duration", remaining)
	}
}

func TestCheckCooldown_IsPerGuild(t *testing.T) {
	userID := "user-checkcooldown-perguild"
	if remaining := checkCooldown("guild-a", userID); remaining != 0 {
		t.Fatalf("checkCooldown(guild-a) = %v, want 0", remaining)
	}
	if remaining := checkCooldown("guild-b", userID); remaining != 0 {
		t.Fatalf("checkCooldown(guild-b) = %v, want 0 (different guild, independent cooldown)", remaining)
	}
}
