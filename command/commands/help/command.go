// Package help implements /help: a Go-idiomatic replacement for the
// original Python bot's prefix-command, cog-introspecting help.py.
// Slash commands are already self-documenting in Discord's own command
// picker (name + description show up there for free), so /help just lists
// every registered command and its description in one place, built
// straight from registry.Commands rather than a hand-maintained text
// (which would drift out of sync, the very thing the plan called out
// wanting to avoid).
package help

import (
	"fmt"
	"sort"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"go_texas_bot/command/command_selector"
	"go_texas_bot/registry"
)

const embedColor = 0xff9900 // matches the original bot's config.json embed_color

func init() {
	command_selector.CommandSelector.AddCommand(command_selector.Key("help", discord.ApplicationCommandTypeSlash), helpCommandListener)
}

func helpCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	embed := discord.Embed{
		Title:  "Помощь, угу...",
		Color:  embedColor,
		Fields: commandFields(),
		Footer: &discord.EmbedFooter{Text: "Все команды доступны через / -- их описания видно прямо в списке команд Discord."},
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed).WithEphemeral(true))
}

// commandFields renders every slash command's "/name — description" as a
// series of embed fields, each kept under Discord's 1024-char field-value
// limit (user/message context-menu entries are skipped since they
// duplicate a slash command already listed, and have no description of
// their own to show).
func commandFields() []discord.EmbedField {
	lines := commandLines()

	const maxFieldLen = 1000
	var fields []discord.EmbedField
	var current strings.Builder
	for _, line := range lines {
		if current.Len() > 0 && current.Len()+len(line)+1 > maxFieldLen {
			fields = append(fields, commandsField(len(fields), current.String()))
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte('\n')
		}
		current.WriteString(line)
	}
	if current.Len() > 0 {
		fields = append(fields, commandsField(len(fields), current.String()))
	}
	return fields
}

func commandsField(index int, value string) discord.EmbedField {
	name := "Команды"
	if index > 0 {
		name = fmt.Sprintf("Команды (продолжение %d)", index+1)
	}
	return discord.EmbedField{Name: name, Value: value}
}

func commandLines() []string {
	lines := make([]string, 0, len(registry.Commands))
	for _, cmd := range registry.Commands {
		slash, ok := cmd.(discord.SlashCommandCreate)
		if !ok {
			continue
		}
		lines = append(lines, fmt.Sprintf("**/%s** — %s", slash.Name, slash.Description))
	}
	sort.Strings(lines)
	return lines
}
