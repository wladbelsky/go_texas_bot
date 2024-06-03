package command_selector

import (
	"github.com/disgoorg/disgo/events"
	"log"
	"sync"
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

func (cs *commandSelector) AddCommand(name string, command CommandFunc) {
	cs.m.Lock()
	defer cs.m.Unlock()
	if _, ok := cs.commandMap[name]; ok {
		log.Panicln("command already exists")
	}
	cs.commandMap[name] = command
}

func (cs *commandSelector) GetCommand(name string) (CommandFunc, bool) {
	cs.m.RLock()
	defer cs.m.RUnlock()
	command, ok := cs.commandMap[name]
	return command, ok
}
