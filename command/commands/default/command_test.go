package defaultcmd

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
)

func TestTargetDisplayName_PrefersMemberNickname(t *testing.T) {
	nick := "Nickname"
	user := discord.User{Username: "username"}
	member := discord.ResolvedMember{Member: discord.Member{User: user, Nick: &nick}}

	if got := targetDisplayName(member, true, user); got != "Nickname" {
		t.Fatalf("targetDisplayName() = %q, want %q", got, "Nickname")
	}
}

func TestTargetDisplayName_FallsBackToUserWithoutMember(t *testing.T) {
	user := discord.User{Username: "username"}

	if got := targetDisplayName(discord.ResolvedMember{}, false, user); got != "username" {
		t.Fatalf("targetDisplayName() without a member = %q, want %q", got, "username")
	}
}
