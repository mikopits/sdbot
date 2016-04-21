package plugins

import (
	"sdbot"
	"time"
)

var HelloWorldPlugin = &sdbot.Plugin{
	Command:      "hi",
	EventHandler: &HelloWorldEventHandler{},
}

type HelloWorldEventHandler struct {
	LastUsed time.Time
}

func (eh *HelloWorldEventHandler) HandleChatEvents(m *sdbot.Message, prefix string, args []string, rest string) {
	eh.LastUsed = time.Now() // Could keep track of cooldown with eh.LastUsed
	m.Reply("hi :)")
}

func (eh *HelloWorldEventHandler) HandlePrivateEvents(m *sdbot.Message, prefix string, args []string, rest string) {
	eh.HandleChatEvents(m, prefix, args, rest)
}
