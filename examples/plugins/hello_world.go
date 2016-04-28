// Sample plugin for sdbot.
// +build ignore

package plugins

import (
	"github.com/mikopits/sdbot"
	"time"
)

// We don't have the bot yet, so define a function literal that allows us to
// set the bot at any time. If you want to define the plugins in the main
// loop alongside the bot then there is no need to do this.
var HelloWorldPlugin = func(b *sdbot.Bot) *sdbot.Plugin {
	p := sdbot.NewPluginWithCooldown(b, "hi", time.Second*5)
	p.SetEventHandler(HelloWorldEventHandler{Plugin: p})
	return p
}

// Set up an alias since we are using the default event handler.
type HelloWorldEventHandler sdbot.DefaultEventHandler

func (eh HelloWorldEventHandler) HandleEvent(m *sdbot.Message, args []string) {
	m.Reply("hi :)")
}
