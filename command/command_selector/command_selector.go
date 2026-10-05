package command_selector

import (
	"fmt"
	"log"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type CommandFunc func(event *events.ApplicationCommandInteractionCreate) error

type commandSelector struct {
	commandMap map[string]CommandFunc
	m          sync.RWMutex
}

var CommandSelector = newCommandSelector()

func newCommandSelector() *commandSelector {
	return &commandSelector{
		commandMap: make(map[string]CommandFunc),
	}
}

// Key builds the lookup key for a command, disambiguating same-named
// commands that exist under different discord.ApplicationCommandType(s)
// (e.g. a "/f" slash command and an "f" user-context-menu command can
// share the name "f" -- Discord treats them as separate namespaces).
func Key(name string, cmdType discord.ApplicationCommandType) string {
	return fmt.Sprintf("%s:%d", name, cmdType)
}

func (cs *commandSelector) AddCommand(key string, command CommandFunc) {
	cs.m.Lock()
	defer cs.m.Unlock()
	if _, ok := cs.commandMap[key]; ok {
		log.Panicln("command already exists")
	}
	cs.commandMap[key] = command
}

func (cs *commandSelector) GetCommand(key string) (CommandFunc, bool) {
	cs.m.RLock()
	defer cs.m.RUnlock()
	command, ok := cs.commandMap[key]
	return command, ok
}
