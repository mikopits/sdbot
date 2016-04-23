package sdbot

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
)

// TODO Add a Killable struct to add functionality to break out of the
// Plugin.Listen loop.
type Plugin struct {
	Bot          *Bot
	Name         string
	Prefixes     []string
	Command      string
	NumArgs      int
	Cooldown     int
	LastUsed     time.Time
	EventHandler EventHandler
}

// Define different convenient ways to make new Plugin structs. You must add
// the bot and event handlers after creation with these methods.
func NewPlugin(b *Bot, cmd string, prefixes ...string) *Plugin {
	var p []string
	if prefixes != nil {
		p = prefixes
	} else {
		p = b.Config.PluginPrefixes
	}

	return &Plugin{
		Bot:      b,
		Command:  cmd,
		Prefixes: p,
	}
}

func NewPluginWithArgs(b *Bot, cmd string, numArgs int, prefixes ...string) *Plugin {
	var p []string
	if prefixes != nil {
		p = prefixes
	} else {
		p = b.Config.PluginPrefixes
	}

	return &Plugin{
		Bot:      b,
		Command:  cmd,
		NumArgs:  numArgs,
		Prefixes: p,
	}
}

func NewPluginWithCooldown(b *Bot, cmd string, cooldown int, prefixes ...string) *Plugin {
	var p []string
	if prefixes != nil {
		p = prefixes
	} else {
		p = b.Config.PluginPrefixes
	}

	return &Plugin{
		Bot:      b,
		Command:  cmd,
		Cooldown: cooldown,
		Prefixes: p,
	}
}

func NewPluginWithArgsAndCooldown(b *Bot, cmd string, numArgs int, cooldown int, prefixes ...string) *Plugin {
	var p []string
	if prefixes != nil {
		p = prefixes
	} else {
		p = b.Config.PluginPrefixes
	}

	return &Plugin{
		Bot:      b,
		Command:  cmd,
		NumArgs:  numArgs,
		Cooldown: cooldown,
		Prefixes: p,
	}
}

func (p *Plugin) SetEventHandler(eh EventHandler) {
	p.EventHandler = eh
}

type TimedPlugin struct {
	Bot               *Bot
	Timer             time.Timer    // Timer that events fire on
	Period            time.Duration // The period over which you want to fire events
	TimedEventHandler TimedEventHandler
}

type EventHandler interface {
	HandleChatEvents(*Message, string, []string, string)
	HandlePrivateEvents(*Message, string, []string, string)
}

type TimedEventHandler interface {
	Start()
	Stop()
}

// Checks if the message matches the plugin and returns:
// If it matched or not (bool)
// The prefix that matched (string)
// The arguments that were passed alongside the command ([]string)
// The rest of the command (string)
func (p *Plugin) Match(m *Message) (bool, string, []string, string) {
	msg := m.Message

	var match string

	for _, prefix := range p.Prefixes {
		if msg[0:int(math.Min(float64(len(prefix)), float64(len(msg))))] == prefix {
			match = prefix
		}
	}

	// No prefix matched
	if match == "" {
		return false, "", []string{}, ""
	}

	// Command not matched
	if msg[len(match):][:len(p.Command)] != p.Command {
		return false, "", []string{}, ""
	}

	// Comma separated message
	cs := strings.Split(msg[len(match)+len(p.Command):], ",")

	var args []string
	var rest bytes.Buffer

	for i, arg := range cs {
		if i < p.NumArgs {
			args = append(args, strings.TrimSpace(arg))
		} else {
			rest.WriteString(",")
			rest.WriteString(arg)
		}
	}

	// Wrong number of arguments
	if p.NumArgs != len(args) {
		return false, "", []string{}, ""
	}

	return true, match, args, strings.TrimSpace(rest.String())
}

// Starts a loop in its own goroutine looking for events.
func (p *Plugin) Listen() {
	go func() {
		for {
			select {
			case m := <-*p.Bot.PluginChatChannels[p.Name]:
				//Debug(&Log, fmt.Sprintf("[message=%+v]", m))
				match, prefix, args, rest := p.Match(m)
				//Debug(&Log, fmt.Sprintf("[match=%t] [prefix=%s] [args=%s] [rest=%s]", match, prefix, args, rest))
				if match {
					Debug(&Log, fmt.Sprintf("Matched on [prefix=%s] [args=%s] [rest=%s]"))
					go p.EventHandler.HandleChatEvents(m, prefix, args, rest)
				}
			case m := <-*p.Bot.PluginPrivateChannels[p.Name]:
				//Debug(&Log, fmt.Sprintf("[message=%+v]", m))
				match, prefix, args, rest := p.Match(m)
				//Debug(&Log, fmt.Sprintf("[match=%t] [prefix=%s] [args=%s] [rest=%s]", match, prefix, args, rest))
				if match {
					Debug(&Log, fmt.Sprintf("Matched on [prefix=%s] [args=%s] [rest=%s]"))
					go p.EventHandler.HandlePrivateEvents(m, prefix, args, rest)
				}
			}
		}
	}()
}
