package say

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestBuildSayMessage(t *testing.T) {
	msg := buildSayMessage("hello there", true)

	if msg.Content != "hello there" {
		t.Fatalf("Content = %q, want %q", msg.Content, "hello there")
	}
	if !msg.Flags.Has(discord.MessageFlagEphemeral) {
		t.Fatalf("expected message to be ephemeral, flags = %v", msg.Flags)
	}
}

func TestBuildSayMessage_NotEphemeral(t *testing.T) {
	msg := buildSayMessage("public message", false)

	if msg.Content != "public message" {
		t.Fatalf("Content = %q, want %q", msg.Content, "public message")
	}
	if msg.Flags.Has(discord.MessageFlagEphemeral) {
		t.Fatalf("expected message to not be ephemeral, flags = %v", msg.Flags)
	}
}
