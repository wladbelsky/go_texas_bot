package music

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// fakeLavalink is a minimal Lavalink v4 websocket endpoint: it rejects the
// first failFirst connection attempts with 503 (Lavalink still starting),
// then accepts and sends the "ready" op disgolink waits for.
func fakeLavalink(t *testing.T, failFirst int32) (host string, port int, attempts *atomic.Int32) {
	t.Helper()
	attempts = new(atomic.Int32)
	upgrader := websocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) <= failFirst {
			http.Error(w, "starting", http.StatusServiceUnavailable)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, []byte(`{"op":"ready","resumed":false,"sessionId":"test"}`))
		// Keep the connection open until the client hangs up.
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	t.Cleanup(srv.Close)

	h, p, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err = strconv.Atoi(p)
	if err != nil {
		t.Fatal(err)
	}
	return h, port, attempts
}

// resetClient clears the package-level client before and after a test. It
// deliberately doesn't Close the disgolink client: in disgolink v3.1.0,
// Node.Close writes its conn field without the lock its listen goroutine
// reads it under, so closing from the test trips the race detector. The
// leftover connection just lives until the test binary exits.
func resetClient(t *testing.T) {
	t.Helper()
	setClient(nil)
	t.Cleanup(func() { setClient(nil) })
}

func TestConnect_RetriesUntilLavalinkIsUp(t *testing.T) {
	host, port, attempts := fakeLavalink(t, 1)
	resetClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := Connect(ctx, 1, host, port, "pw"); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	if Client() == nil {
		t.Fatal("Client() is nil after Connect succeeded")
	}
	if got := attempts.Load(); got < 2 {
		t.Fatalf("expected a retry after the first rejected attempt, got %d attempt(s)", got)
	}
}

func TestConnect_GivesUpWhenContextIsDone(t *testing.T) {
	host, port, _ := fakeLavalink(t, 1<<30) // never comes up
	resetClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := Connect(ctx, 1, host, port, "pw")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Connect() error = %v, want context.DeadlineExceeded", err)
	}
	if Client() != nil {
		t.Fatal("Client() must stay nil while Lavalink is unreachable")
	}
}
