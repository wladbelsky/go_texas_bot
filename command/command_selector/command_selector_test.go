package command_selector

import (
	"testing"

	"github.com/disgoorg/disgo/events"
)

func noop(_ *events.ApplicationCommandInteractionCreate) error {
	return nil
}

func TestAddAndGetCommand(t *testing.T) {
	cs := newCommandSelector()

	called := false
	cs.AddCommand("ping", func(e *events.ApplicationCommandInteractionCreate) error {
		called = true
		return noop(e)
	})

	fn, ok := cs.GetCommand("ping")
	if !ok {
		t.Fatal("GetCommand(\"ping\") not found")
	}
	if err := fn(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("registered command was not invoked")
	}
}

func TestGetCommand_Missing(t *testing.T) {
	cs := newCommandSelector()

	if _, ok := cs.GetCommand("missing"); ok {
		t.Fatal("expected GetCommand to report missing command")
	}
}

func TestAddCommand_PanicsOnDuplicate(t *testing.T) {
	cs := newCommandSelector()
	cs.AddCommand("dup", noop)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when registering a duplicate command name")
		}
	}()
	cs.AddCommand("dup", noop)
}
