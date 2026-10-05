package command

import (
	"errors"
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestBuildErrorMessage(t *testing.T) {
	msg := buildErrorMessage(errors.New("boom"))

	want := "The command failed to execute.\nError: boom"
	if msg.Content != want {
		t.Fatalf("Content = %q, want %q", msg.Content, want)
	}
	if !msg.Flags.Has(discord.MessageFlagEphemeral) {
		t.Fatalf("expected error message to be ephemeral, flags = %v", msg.Flags)
	}
}
