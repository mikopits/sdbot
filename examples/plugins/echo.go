// Sample plugin for sdbot.
// +build ignore
package plugins

import (
	"github.com/mikopits/sdbot"
)

// Repeat what was said after the echo command.
var EchoPlugin = func() *sdbot.Plugin {
	p := sdbot.NewPluginWithArgs("echo", 1)
	p.SetEventHandler(&EchoEventHandler{Plugin: p})
	return p
}

type EchoEventHandler sdbot.DefaultEventHandler

func (eh *EchoEventHandler) HandleEvent(m *sdbot.Message, args []string) {
	m.Reply(args[0])
}
