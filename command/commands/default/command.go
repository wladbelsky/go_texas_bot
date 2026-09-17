// Package defaultcmd implements /ping, /announce, /f, /o7, /avatar, /info
// and /invite -- the Go port of the original Python bot's default.py cog
// (named defaultcmd since "default" is a Go keyword).
package defaultcmd

import (
	"fmt"
	"math/rand"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"go_texas_bot/command/command_selector"
)

const embedColor = 0xff9900 // matches the original bot's config.json embed_color

func init() {
	add := func(name string, cmdType discord.ApplicationCommandType, fn command_selector.CommandFunc) {
		command_selector.CommandSelector.AddCommand(command_selector.Key(name, cmdType), fn)
	}
	add("ping", discord.ApplicationCommandTypeSlash, pingCommandListener)
	add("announce", discord.ApplicationCommandTypeSlash, announceCommandListener)
	add("info", discord.ApplicationCommandTypeSlash, infoCommandListener)
	add("invite", discord.ApplicationCommandTypeSlash, inviteCommandListener)

	add("f", discord.ApplicationCommandTypeSlash, fCommandListener)
	add("f", discord.ApplicationCommandTypeUser, fCommandListener)
	add("o7", discord.ApplicationCommandTypeSlash, o7CommandListener)
	add("o7", discord.ApplicationCommandTypeUser, o7CommandListener)
	add("avatar", discord.ApplicationCommandTypeSlash, avatarCommandListener)
	add("avatar", discord.ApplicationCommandTypeUser, avatarCommandListener)
}

func pingCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	latency := event.Client().Gateway.Latency()
	return event.CreateMessage(discord.NewMessageCreate().
		WithContentf("Pong! %.1f ms", float64(latency.Microseconds())/1000),
	)
}

func announceCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	member := event.Member()
	if member == nil || !member.Permissions.Has(discord.PermissionAdministrator) {
		return respondEphemeral(event, "Нужны права администратора.")
	}

	data := event.SlashCommandInteractionData()
	channel := data.Channel("channel")
	message := data.String("message")

	if _, err := event.Client().Rest.CreateMessage(channel.ID, discord.NewMessageCreate().WithContent(message)); err != nil {
		return err
	}
	return respondEphemeral(event, "Alldone, boss")
}

func inviteCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	url := fmt.Sprintf(
		"https://discord.com/oauth2/authorize?client_id=%s&scope=applications.commands%%20bot&permissions=3401792",
		event.Client().ApplicationID,
	)
	return respondEphemeral(event, url)
}

// targetMember resolves the "other member" for f/o7/avatar: the "member"
// slash option if this is a slash invocation, or the right-clicked member
// for a user-context-menu invocation. ok is false for a slash invocation
// with no member option given.
func targetMember(event *events.ApplicationCommandInteractionCreate) (discord.Member, bool) {
	switch event.Data.Type() {
	case discord.ApplicationCommandTypeUser:
		return event.UserCommandInteractionData().TargetMember().Member, true
	default:
		m, ok := event.SlashCommandInteractionData().OptMember("member")
		return m.Member, ok
	}
}

func fCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	author := event.Member()
	target, hasTarget := targetMember(event)

	title := fmt.Sprintf("**%s** заплатил увожение. o7", memberName(event, author))
	if hasTarget {
		title = fmt.Sprintf("**%s** заплатил увожение за %s", memberName(event, author), target.EffectiveName())
	}

	embed := discord.Embed{
		Title: title,
		Color: embedColor,
		Image: &discord.EmbedResource{URL: "https://pbs.twimg.com/media/D-5sUKNXYAA5K9l.jpg"},
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed))
}

func o7CommandListener(event *events.ApplicationCommandInteractionCreate) error {
	author := event.Member()
	target, hasTarget := targetMember(event)

	title := fmt.Sprintf("**%s** приветствует вас командиры. o7", memberName(event, author))
	if hasTarget {
		title = fmt.Sprintf("**%s** приветствует %s. o7", memberName(event, author), target.EffectiveName())
	}

	embed := discord.Embed{
		Title: title,
		Color: embedColor,
		Image: &discord.EmbedResource{URL: pressfImages[rand.Intn(len(pressfImages))]},
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed))
}

func avatarCommandListener(event *events.ApplicationCommandInteractionCreate) error {
	target, ok := targetMember(event)
	if !ok {
		return respondEphemeral(event, "Укажи участника через параметр member.")
	}

	embed := discord.Embed{
		Title: "Avatarr",
		Image: &discord.EmbedResource{URL: target.User.EffectiveAvatarURL()},
	}
	return event.CreateMessage(discord.NewMessageCreate().WithEmbeds(embed).WithEphemeral(true))
}

func memberName(event *events.ApplicationCommandInteractionCreate, member *discord.ResolvedMember) string {
	if member != nil {
		return member.EffectiveName()
	}
	return event.User().EffectiveName()
}

func respondEphemeral(event *events.ApplicationCommandInteractionCreate, content string) error {
	return event.CreateMessage(discord.NewMessageCreate().WithContent(content).WithEphemeral(true))
}

// mustParseSnowflake parses a stored user ID, falling back to 0 (an
// invalid/unmentionable ID) if it's ever malformed -- which shouldn't
// happen since we're the only ones who write these IDs to the database.
func mustParseSnowflake(s string) snowflake.ID {
	id, err := snowflake.Parse(s)
	if err != nil {
		return 0
	}
	return id
}
