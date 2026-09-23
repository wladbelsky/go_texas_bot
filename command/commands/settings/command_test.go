package settings

import (
	"strings"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"go_texas_bot/db"
)

func TestCanManageGuild(t *testing.T) {
	cases := []struct {
		name  string
		perms discord.Permissions
		want  bool
	}{
		{"manage guild only", discord.PermissionManageGuild, true},
		{"administrator only", discord.PermissionAdministrator, true},
		{"both", discord.PermissionManageGuild | discord.PermissionAdministrator, true},
		{"neither", discord.PermissionSendMessages, false},
		{"none", 0, false},
	}
	for _, c := range cases {
		if got := canManageGuild(c.perms); got != c.want {
			t.Errorf("%s: canManageGuild() = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestDescribe_ReflectsBothFlags(t *testing.T) {
	got := describe(db.GuildSettings{NSFWEnabled: true, ToxicGreetingsEnabled: false})
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("describe() = %q, want a header plus one line per setting", got)
	}
	if !strings.HasSuffix(lines[1], "включено") || !strings.HasSuffix(lines[2], "выключено") {
		t.Fatalf("describe() = %q, want NSFW on and toxic greetings off", got)
	}
}
