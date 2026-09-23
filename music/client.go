package music

import (
	"context"
	"fmt"
	"sync"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"
)

var (
	clientMu sync.RWMutex
	client   disgolink.Client
)

// Client returns the shared disgolink client used by every music command,
// or nil while Lavalink isn't connected yet (Connect is still retrying).
// Callers should treat nil as "music is unavailable right now".
func Client() disgolink.Client {
	clientMu.RLock()
	defer clientMu.RUnlock()
	return client
}

func setClient(c disgolink.Client) {
	clientMu.Lock()
	defer clientMu.Unlock()
	client = c
}

// Queues is the shared per-guild queue manager used by every music command.
var Queues = NewManager()

// Connect creates the disgolink client for the given bot user and connects
// it to a single Lavalink node, blocking until the node is up. disgolink
// retries a failed connection itself (backing off up to 30s between tries),
// so Connect only returns an error once ctx is done. The client is published
// through Client only after the node has connected; main runs Connect in the
// background so the rest of the bot works while Lavalink is still starting.
//
// Once connected, disgolink also reconnects on its own if the websocket to
// Lavalink drops later on.
func Connect(ctx context.Context, botUserID snowflake.ID, host string, port int, password string) error {
	c := disgolink.New(botUserID, disgolink.WithListenerFunc(onTrackEnd))
	if _, err := c.AddNode(ctx, disgolink.NodeConfig{
		Name:     "main",
		Address:  fmt.Sprintf("%s:%d", host, port),
		Password: password,
	}); err != nil {
		c.Close()
		return err
	}
	setClient(c)
	return nil
}
