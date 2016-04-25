package sdbot

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TODO Add a Killable struct to add functionality to break out of the
// Plugin.Listen loop.
//
// TODO Automation of cooldown. Should not have to implement it in a plugin.
type Plugin struct {
	Bot          *Bot
	Name         string
	Prefix       *regexp.Regexp
	Suffix       *regexp.Regexp
	Command      string
	NumArgs      int
	Cooldown     int
	LastUsed     time.Time
	EventHandler EventHandler
}

// Define different convenient ways to make new Plugin structs. You must add
// the bot and event handlers after creation with these methods. If you want
// to add custom prefixes or suffixes, you can do it with the Plugin.SetPrefix
// and Plugin.SetSuffix methods. Otherwise the bot will load prefixes and
// suffixes from your Config.
func NewPlugin(b *Bot, cmd string) *Plugin {
	return &Plugin{
		Bot:     b,
		Command: cmd,
	}
}

func NewPluginWithArgs(b *Bot, cmd string, numArgs int) *Plugin {
	return &Plugin{
		Bot:     b,
		Command: cmd,
		NumArgs: numArgs,
	}
}

func NewPluginWithCooldown(b *Bot, cmd string, cooldown int) *Plugin {
	return &Plugin{
		Bot:      b,
		Command:  cmd,
		Cooldown: cooldown,
	}
}

func NewPluginWithArgsAndCooldown(b *Bot, cmd string, numArgs int, cooldown int) *Plugin {
	return &Plugin{
		Bot:      b,
		Command:  cmd,
		NumArgs:  numArgs,
		Cooldown: cooldown,
	}
}

// Call this if you want this plugin to match to prefixes other than the
// default prefixes defined in your Config.
func (p *Plugin) SetPrefix(prefixes []string) {
	if len(prefixes) == 0 {
		return
	}

	regStr := "^(" + strings.Join(prefixes, "|") + ")"
	reg, err := regexp.Compile(regStr)
	if err != nil {
		Error(&Log, err)
	}

	p.Prefix = reg
}

// Call this if you want this plugin to match to suffixes other than the
// default suffixes defined in your Config.
func (p *Plugin) SetSuffix(suffixes []string) {
	if len(suffixes) == 0 {
		return
	}

	regStr := "(" + strings.Join(suffixes, "|") + ")$"
	reg, err := regexp.Compile(regStr)
	if err != nil {
		Error(&Log, err)
	}

	p.Suffix = reg
}

func (p *Plugin) FormatPrefixAndSuffix() {
	ps := p.Prefix.String()
	ss := p.Suffix.String()
	var flags string

	if p.Bot.Config.CaseInsensitive {
		flags = "(?i)"
	}

	p.Prefix = regexp.MustCompile(fmt.Sprintf("^(%s%s%s)", flags, ps[1:], p.Command))
	p.Suffix = regexp.MustCompile(fmt.Sprintf("(%s%s)$", flags, ss[:len(ss)-1]))
}

// Allows you to use a custom event handler with any fields you want.
func (p *Plugin) SetEventHandler(eh EventHandler) {
	p.EventHandler = eh
}

type DefaultEventHandler struct {
	Plugin   *Plugin
	LastUsed time.Time
}

func NewDefaultEventHandler(p *Plugin) *DefaultEventHandler {
	return &DefaultEventHandler{Plugin: p}
}

type TimedPlugin struct {
	Bot               *Bot
	Timer             time.Timer    // Timer that events fire on
	Period            time.Duration // The period over which you want to fire events
	TimedEventHandler TimedEventHandler
}

type EventHandler interface {
	HandleChatEvents(*Message, string, []string)
	HandlePrivateEvents(*Message, string, []string)
}

type TimedEventHandler interface {
	Start()
	Stop()
}

// Find out if the message is a match for this plugin.
func (p *Plugin) match(m *Message) bool {
	return p.Prefix.MatchString(m.Message) && p.Suffix.MatchString(m.Message)
}

// Parse the message. Returns values:
// 0. The command input. (eg. !echo Hello World) for prefix "!", command "echo",
// numArgs 0 will give input "Hello World".
// 1. The arguments provided. (eg. !echoNTimes 5, "Hello World") for prefix "!",
// command "echoNTimes", numArgs 1 will give input "Hello World" and args ["5"]
func (p *Plugin) parse(m *Message) (string, []string) {
	Debug(&Log, fmt.Sprintf("[msg=%s] [fs1=%s] [fs2=%s] [pre=%+v] [suf=%+v]", m.Message, p.Prefix.FindString(m.Message), p.Suffix.FindString(m.Message), p.Prefix, p.Suffix))
	input := m.Message[len(p.Prefix.FindString(m.Message)) : len(m.Message)-len(p.Suffix.FindString(m.Message))]
	if p.NumArgs == 0 {
		return strings.TrimSpace(input), nil
	}

	var args []string
	var buffer bytes.Buffer

	for i, arg := range strings.Split(input, "|") {
		if i < p.NumArgs {
			args = append(args, strings.TrimSpace(arg))
		} else {
			buffer.WriteString(",")
			buffer.WriteString(arg)
		}
	}

	return strings.TrimSpace(buffer.String()), args
}

// Starts a loop in its own goroutine looking for events.
func (p *Plugin) Listen() {
	go func() {
		for {
			select {
			case m := <-*p.Bot.PluginChatChannels[p.Name]:
				if p.match(m) {
					input, args := p.parse(m)
					Debug(&Log, fmt.Sprintf("[on plugin] Starting goroutine for plugin `%s`", p.Name))
					go p.EventHandler.HandleChatEvents(m, input, args)
				}
			case m := <-*p.Bot.PluginPrivateChannels[p.Name]:
				if p.match(m) {
					input, args := p.parse(m)
					Debug(&Log, fmt.Sprintf("[on plugin] Starting goroutine for plugin `%s`", p.Name))
					go p.EventHandler.HandlePrivateEvents(m, input, args)
				}
			}
		}
	}()
}
