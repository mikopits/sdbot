package plugins

import (
	"sdbot"
)

// We don't have the bot yet, so define a function literal that we can call at
// a later time.
var HelloWorldPlugin = func(b *sdbot.Bot) *sdbot.Plugin {
	p := &sdbot.Plugin{
		Bot:      b,
		Prefixes: []string{},
		Command:  "hi",
		Cooldown: 5,
		NumArgs:  0,
	}
	p.EventHandler = &HelloWorldEventHandler{Plugin: p}
	return p
}

type HelloWorldEventHandler struct {
	Plugin *sdbot.Plugin
}

func (eh *HelloWorldEventHandler) HandleChatEvents(m *sdbot.Message, prefix string, args []string, rest string) {
	if int(m.Time.Unix()-eh.Plugin.LastUsed.Unix()) < eh.Plugin.Cooldown {
		m.Reply("cooldown not done")
	} else {
		m.Reply("hi :)")
	}
	eh.Plugin.LastUsed = m.Time
}

func (eh *HelloWorldEventHandler) HandlePrivateEvents(m *sdbot.Message, prefix string, args []string, rest string) {
	eh.HandleChatEvents(m, prefix, args, rest)
}
