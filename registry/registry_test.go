package registry

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestCommands_ContainsSay(t *testing.T) {
	for _, cmd := range Commands {
		slash, ok := cmd.(discord.SlashCommandCreate)
		if !ok {
			continue
		}
		if slash.Name == "say" {
			return
		}
	}
	t.Fatal("expected Commands to register a \"say\" slash command")
}

func TestCommands_NoDuplicateNamesWithinType(t *testing.T) {
	type key struct {
		name    string
		cmdType discord.ApplicationCommandType
	}
	seen := make(map[key]bool)
	for _, cmd := range Commands {
		var k key
		switch c := cmd.(type) {
		case discord.SlashCommandCreate:
			k = key{c.Name, discord.ApplicationCommandTypeSlash}
		case discord.UserCommandCreate:
			k = key{c.Name, discord.ApplicationCommandTypeUser}
		case discord.MessageCommandCreate:
			k = key{c.Name, discord.ApplicationCommandTypeMessage}
		default:
			continue
		}
		if seen[k] {
			t.Fatalf("duplicate command registration: name=%q type=%v", k.name, k.cmdType)
		}
		seen[k] = true
	}
}
