package config

import "testing"

func TestToken(t *testing.T) {
	t.Setenv("BOT_TOKEN", "test-token")
	if got := Token(); got != "test-token" {
		t.Fatalf("Token() = %q, want %q", got, "test-token")
	}
}

func TestLavalinkHost(t *testing.T) {
	t.Setenv("LAVALINK_HOST", "lavalink.local")
	if got := LavalinkHost(); got != "lavalink.local" {
		t.Fatalf("LavalinkHost() = %q, want %q", got, "lavalink.local")
	}
}

func TestLavalinkPort(t *testing.T) {
	t.Setenv("LAVALINK_PORT", "2333")
	if got := LavalinkPort(); got != 2333 {
		t.Fatalf("LavalinkPort() = %d, want %d", got, 2333)
	}
}

func TestLavalinkPassword(t *testing.T) {
	t.Setenv("LAVALINK_PASSWORD", "secret")
	if got := LavalinkPassword(); got != "secret" {
		t.Fatalf("LavalinkPassword() = %q, want %q", got, "secret")
	}
}

func TestDBPath_DefaultsToFlagValue(t *testing.T) {
	if got := DBPath(); got != "data/texas_bot.db" {
		t.Fatalf("DBPath() = %q, want default %q", got, "data/texas_bot.db")
	}
}

func TestDBPath_EnvOverridesDefault(t *testing.T) {
	t.Setenv("DB_PATH", "/custom/path.db")
	if got := DBPath(); got != "/custom/path.db" {
		t.Fatalf("DBPath() = %q, want %q", got, "/custom/path.db")
	}
}
