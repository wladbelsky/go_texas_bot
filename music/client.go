package music

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

// Lavalink is the shared disgolink client used by every music command. It's
// nil until Init is called from main().
var Lavalink disgolink.Client

// Queues is the shared per-guild queue manager used by every music command.
var Queues = NewManager()

// Init creates the disgolink client for the given bot user and connects it
// to a single Lavalink node. Lavalink is only set once the node is actually
// connected, so callers can keep treating a nil Lavalink as "music is
// unavailable".
func Init(ctx context.Context, botUserID snowflake.ID, host string, port int, password string) error {
	client := disgolink.New(botUserID, disgolink.WithListenerFunc(onTrackEnd))
	if _, err := client.AddNode(ctx, disgolink.NodeConfig{
		Name:     "main",
		Address:  fmt.Sprintf("%s:%d", host, port),
		Password: password,
	}); err != nil {
		return err
	}
	Lavalink = client
	return nil
}
