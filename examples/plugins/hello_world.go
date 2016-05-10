// Sample plugin for sdbot.
// +build ignore

package plugins

import (
	"github.com/mikopits/sdbot"
	"time"
)

var HelloWorldPlugin = func() *sdbot.Plugin {
	p := sdbot.NewPluginWithCooldown("hi", time.Second*5)
	p.SetEventHandler(HelloWorldEventHandler{Plugin: p})
	return p
}

// Set up an alias since we are using the default event handler.
type HelloWorldEventHandler sdbot.DefaultEventHandler

func (eh HelloWorldEventHandler) HandleEvent(m *sdbot.Message, args []string) {
	m.Reply("hi :)")
}
