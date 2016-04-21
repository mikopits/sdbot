package sdbot

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
)

var ChatEvents = make(chan *Message, 64)
var PrivateEvents = make(chan *Message, 64)

type Plugin struct {
	Bot          *Bot
	Prefixes     []string // The prefixes you want the plugin to fire on
	Command      string   // The command you want the plugin to fire on
	NumArgs      int      // Number of comma separated arguments you expect for your plugin.
	Cooldown     int      // Number of seconds before the command can be used again. No cooldown if <= 0.
	EventHandler EventHandler
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

func (p *Plugin) Listen() {
	go func() {
		for {
			select {
			case m := <-ChatEvents:
				Debug(&Log, "Got chat event")
				match, prefix, args, rest := p.Match(m)
				if match {
					Debug(&Log, fmt.Sprintf("Matched on [prefix=%s] [args=%s] [rest=%s]"))
					p.EventHandler.HandleChatEvents(m, prefix, args, rest)
				}
			case m := <-PrivateEvents:
				Debug(&Log, "Got pm event")
				match, prefix, args, rest := p.Match(m)
				if match {
					Debug(&Log, fmt.Sprintf("Matched on [prefix=%s] [args=%s] [rest=%s]"))
					p.EventHandler.HandlePrivateEvents(m, prefix, args, rest)
				}
			}
		}
	}()
}
