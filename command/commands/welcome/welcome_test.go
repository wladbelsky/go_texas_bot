package welcome

import (
	"encoding/json"
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func textChannel(t *testing.T, id, name string) discord.GuildChannel {
	t.Helper()
	var ch discord.GuildTextChannel
	payload := `{"id":"` + id + `","name":"` + name + `"}`
	if err := json.Unmarshal([]byte(payload), &ch); err != nil {
		t.Fatalf("failed to build test channel: %v", err)
	}
	return ch
}

func TestFindChannelByName_Found(t *testing.T) {
	channels := []discord.GuildChannel{
		textChannel(t, "1", "general"),
		textChannel(t, "2", "основной"),
		textChannel(t, "3", "random"),
	}

	id, ok := findChannelByName(channels, leaveChannelName)
	if !ok {
		t.Fatal("expected to find the channel named \"основной\"")
	}
	if id.String() != "2" {
		t.Fatalf("id = %v, want 2", id)
	}
}

func TestFindChannelByName_NotFound(t *testing.T) {
	channels := []discord.GuildChannel{
		textChannel(t, "1", "general"),
		textChannel(t, "3", "random"),
	}

	if _, ok := findChannelByName(channels, leaveChannelName); ok {
		t.Fatal("expected findChannelByName to report no match")
	}
}

func TestFindChannelByName_EmptyList(t *testing.T) {
	if _, ok := findChannelByName(nil, leaveChannelName); ok {
		t.Fatal("expected findChannelByName to report no match on an empty list")
	}
}
