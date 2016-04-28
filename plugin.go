package sdbot

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// TODO Add a Killable struct to add functionality to break out of the
// Plugin.Listen loop.
type Plugin struct {
	Bot          *Bot
	Name         string
	Prefix       *regexp.Regexp
	Suffix       *regexp.Regexp
	Command      string
	NumArgs      int
	Cooldown     time.Duration
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

// Will be triggered on every chat and private event.
func NewPluginWithoutCommand(b *Bot) *Plugin {
	return &Plugin{Bot: b}
}

func NewPluginWithArgs(b *Bot, cmd string, numArgs int) *Plugin {
	return &Plugin{
		Bot:     b,
		Command: cmd,
		NumArgs: numArgs,
	}
}

func NewPluginWithCooldown(b *Bot, cmd string, cooldown time.Duration) *Plugin {
	return &Plugin{
		Bot:      b,
		Command:  cmd,
		Cooldown: cooldown,
	}
}

func NewPluginWithArgsAndCooldown(b *Bot, cmd string, numArgs int, cooldown time.Duration) *Plugin {
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
	var args string

	if p.Bot.Config.CaseInsensitive {
		flags = "(?i)"
	}

	if p.NumArgs > 0 {
		if p.NumArgs == 1 {
			args = " +(.+)"
		} else {
			args = " +([^,]+)"
		}
		for i := 0; i < p.NumArgs-1; i++ {
			if i == p.NumArgs-2 {
				args = strings.Join([]string{args, ", +(.+)"}, "")
			} else {
				args = strings.Join([]string{args, ", +([^,]+)"}, "")
			}
		}
	} else {
		p.Prefix = regexp.MustCompile(fmt.Sprintf("^(%s%s%s$)", flags, ps[1:], p.Command))
		p.Suffix = regexp.MustCompile(fmt.Sprintf("(%s%s)$", flags, ss[:len(ss)-1]))
		return
	}

	p.Prefix = regexp.MustCompile(fmt.Sprintf("^(%s%s%s%s)", flags, ps[1:], p.Command, args))
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
	Ticker            *time.Ticker
	TimedEventHandler TimedEventHandler
}

func NewTimedPlugin(b *Bot, period time.Duration) *TimedPlugin {
	return &TimedPlugin{
		Bot:    b,
		Ticker: time.NewTicker(period),
	}
}

func (tp *TimedPlugin) SetEventHandler(teh TimedEventHandler) {
	tp.TimedEventHandler = teh
}

type DefaultTimedEventHandler struct {
	TimedPlugin *TimedPlugin
}

type EventHandler interface {
	HandleEvent(*Message, []string)
}

type TimedEventHandler interface {
	Start()
	Stop()
}

// Find out if the message is a match for this plugin.
func (p *Plugin) match(m *Message) bool {
	return p.Prefix.MatchString(m.Message) && p.Suffix.MatchString(m.Message)
}

// Parse the message. Returns the arguments provided to the message:
func (p *Plugin) parse(m *Message) []string {
	submatches := p.Prefix.FindStringSubmatch(m.Message)

	switch p.NumArgs {
	case 0:
		return []string{}
	default:
		return submatches[3:]
	}
}

// Starts a loop in its own goroutine listening for events.
func (p *Plugin) Listen() {
	go func() {
		for {
			select {
			case m := <-*p.Bot.PluginChatChannels[p.Name]:
				if p.match(m) {
					args := p.parse(m)
					Debug(&Log, fmt.Sprintf("[on plugin] Starting chat event handler goroutine for plugin `%s` with args `%+v`", p.Name, args))
					if m.Time.Sub(p.LastUsed) > p.Cooldown {
						p.LastUsed = m.Time
						go p.EventHandler.HandleEvent(m, args)
					}
				}
			case m := <-*p.Bot.PluginPrivateChannels[p.Name]:
				if p.match(m) {
					args := p.parse(m)
					Debug(&Log, fmt.Sprintf("[on plugin] Starting private event handler goroutine for plugin `%s` with args `%+v`", p.Name, args))
					if m.Time.Sub(p.LastUsed) > p.Cooldown {
						p.LastUsed = m.Time
						go p.EventHandler.HandleEvent(m, args)
					}
				}
			}
		}
	}()
}
