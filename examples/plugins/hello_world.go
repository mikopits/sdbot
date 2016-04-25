package plugins

import (
	"github.com/mikopits/sdbot"
)

// We don't have the bot yet, so define a function literal that allows us to
// set the bot at any time. If you want to define the plugins in the main
// loop alongside the bot then there is no need to do this.
var HelloWorldPlugin = func(b *sdbot.Bot) *sdbot.Plugin {
	p := sdbot.NewPluginWithCooldown(b, "hi", 5)
	p.EventHandler = &HelloWorldEventHandler{Plugin: p}
	return p
}

// Set up an alias since we are using the default event handler.
type HelloWorldEventHandler sdbot.DefaultEventHandler

func (eh HelloWorldEventHandler) HandleChatEvents(m *sdbot.Message, input string, args []string) {
	if int(m.Time.Unix()-eh.LastUsed.Unix()) < eh.Plugin.Cooldown {
		m.Reply("cooldown not done")
	} else {
		m.Reply("hi :)")
	}
	eh.LastUsed = m.Time
}

func (eh HelloWorldEventHandler) HandlePrivateEvents(m *sdbot.Message, input string, args []string) {
	eh.HandleChatEvents(m, input, args)
}
