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
// to a single Lavalink node.
func Init(ctx context.Context, botUserID snowflake.ID, host string, port int, password string) error {
	Lavalink = disgolink.New(botUserID, disgolink.WithListenerFunc(onTrackEnd))
	_, err := Lavalink.AddNode(ctx, disgolink.NodeConfig{
		Name:     "main",
		Address:  fmt.Sprintf("%s:%d", host, port),
		Password: password,
	})
	return err
}
