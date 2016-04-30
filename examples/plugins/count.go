// Sample timed plugin for sdbot.
// +build ignore
package sdbot

import (
	"github.com/mikopits/sdbot"
	"strconv"
	"time"
)

var CountPlugin = func() *sdbot.TimedPlugin {
	tp := sdbot.NewTimedPlugin(time.Second * 5)
	tp.SetEventHandler(&CountEventHandler{TimedPlugin: tp})
}

type CountEventHandler struct {
	TimedPlugin *TimedPlugin
	Count       int
}

// Send the next int to yourself in pm every 5 seconds.
func (teh *CountEventHandler) HandleEvent() {
	b := teh.TimedPlugin.Bot
	b.FindUserEnsured(b.Nick).Reply(strconv.Itoa(teh.Count))
	teh.Count++
}
