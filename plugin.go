package sdbot

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"
	"time"
)

// A plugin is a command that triggers on some regexp and performs a task
// based on the message that triggered it. It can trigger on several
// different formats (consider prefix "." and command "cmd"). Note that
// cooldowns will not affect command syntax, but will ignore the command
// should it have been sent less than Cooldown time from the last instance.
//
// NewPlugin: ".cmd"
// NewPluginWithArgs:
//   1 arg:  ".cmd arg0"
//   2 args: ".cmd arg1, arg0" (space is optional)
//   n args: ".cmd arg1, arg2, arg3, ..., arg n, arg0"
//
// See that the last argument is always considered to be the first, after
// which they are numbered in the order provided.
//
// It is possible to define a Command with a regexp. For example, a command
// of "say|echo" will trigger on either say or echo.
//
// Each fired event is run in its own separate goroutine, so for anything
// that must be run sequentially (ie. cannot read and write to the same file
// at once) use Bot.Synchronize.
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
	k            Killable
}

// A timed plugin will fire an event on a regular schedule defined by the
// time.Duration provided to the time.Ticker. Each event is run in its own
// goroutine, so every event will fire regardless of whether or not the last
// event ticked had completed. For anything that must be run sequentially
// (ie. cannot read and write to the same file at once) use Bot.Synchronize.
//
// Because each event fires in its own goroutine, you should take care to not
// have each event take longer than the duration of the ticker, or else you
// will be spawning goroutines faster than you can finish them, which is a
// recipe for disaster.
type TimedPlugin struct {
	Bot               *Bot
	Name              string
	Ticker            *time.Ticker
	Period            time.Duration
	TimedEventHandler TimedEventHandler
	k                 Killable
}

type DefaultEventHandler struct {
	Plugin *Plugin
}

type DefaultTimedEventHandler struct {
	TimedPlugin *TimedPlugin
}

// Define different convenient ways to make new Plugin structs. You must add
// the bot and event handlers after creation with these methods. If you want
// to add custom prefixes or suffixes, you can do it with the Plugin.SetPrefix
// and Plugin.SetSuffix methods. Otherwise the bot will load prefixes and
// suffixes from your Config.
func NewPlugin(cmd string) *Plugin {
	return &Plugin{
		Command: cmd,
	}
}

// Will be triggered on every chat and private event.
func NewPluginWithoutCommand() *Plugin {
	return &Plugin{}
}

func NewPluginWithArgs(cmd string, numArgs int) *Plugin {
	return &Plugin{
		Command: cmd,
		NumArgs: numArgs,
	}
}

func NewPluginWithCooldown(cmd string, cooldown time.Duration) *Plugin {
	return &Plugin{
		Command:  cmd,
		Cooldown: cooldown,
	}
}

func NewPluginWithArgsAndCooldown(cmd string, numArgs int, cooldown time.Duration) *Plugin {
	return &Plugin{
		Command:  cmd,
		NumArgs:  numArgs,
		Cooldown: cooldown,
	}
}

func NewTimedPlugin(period time.Duration) *TimedPlugin {
	return &TimedPlugin{
		Period: period,
	}
}

func NewDefaultEventHandler(p *Plugin) *DefaultEventHandler {
	return &DefaultEventHandler{Plugin: p}
}

// Allows you to use a custom event handler with any fields you want.
func (p *Plugin) SetEventHandler(eh EventHandler) {
	p.EventHandler = eh
}

func (tp *TimedPlugin) SetEventHandler(teh TimedEventHandler) {
	tp.TimedEventHandler = teh
}

// Call this if you want this plugin to match to prefixes other than the
// default prefixes defined in your Config.
func (p *Plugin) SetPrefix(prefixes []string) {
	if len(prefixes) == 0 {
		return
	}

	regStr := "^(" + strings.Join(prefixes, "|") + ")"
	reg, err := regexp.Compile(regStr)
	CheckErr(err)

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
	CheckErr(err)

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
// Recovers errors so that the bot doesn't crash on plugin errors.
// TODO Better output than simply debug.PrintStack()
func (p *Plugin) Listen() {
	go func() {
		for {
			select {
			case m := <-p.Bot.pluginChatChannelsRead(p.Name):
				if !p.Bot.Config.IgnoreChatMessages && p.match(m) {
					args := p.parse(m)
					Debugf(&Log, "[on plugin] Starting chat event handler goroutine for plugin `%s` with args `%+v`", p.Name, args)
					if m.Time.Sub(p.LastUsed) > p.Cooldown {
						p.LastUsed = m.Time
						go func() {
							defer func() {
								if r := recover(); r != nil {
									err, ok := r.(error)
									if !ok {
										Error(&Log, err)
										debug.PrintStack()
									}
								}
							}()

							p.EventHandler.HandleEvent(m, args)
						}()
					}
				}
			case m := <-p.Bot.pluginPrivateChannelsRead(p.Name):
				if !p.Bot.Config.IgnorePrivateMessages && p.match(m) {
					args := p.parse(m)
					Debugf(&Log, "[on plugin] Starting private event handler goroutine for plugin `%s` with args `%+v`", p.Name, args)
					if m.Time.Sub(p.LastUsed) > p.Cooldown {
						p.LastUsed = m.Time
						go func() {
							defer func() {
								if r := recover(); r != nil {
									err, ok := r.(error)
									if !ok {
										Error(&Log, err)
										debug.PrintStack()
									}
								}
							}()

							p.EventHandler.HandleEvent(m, args)
						}()
					}
				}
			case <-p.k.Dying():
				// Break out of the listen loop when we call p.k.Kill()
				return
			}
		}
	}()
}

// Request the termination of the Plugin.Listen loop.
func (p *Plugin) StopListening() {
	p.k.Kill()
	p.k.Wait()
}

// Starts a loop listening on the time.Ticker.
func (tp *TimedPlugin) Start() {
	tp.Ticker = time.NewTicker(tp.Period)
	go func() {
		for {
			select {
			case <-tp.Ticker.C:
				go tp.TimedEventHandler.HandleEvent()
			case <-tp.k.Dying():
				// Break out of the loop when we call tp.k.Kill()
				return
			}
		}
	}()
}

// Request the termination of the TimedPlugin.Start loop.
func (tp *TimedPlugin) Stop() {
	tp.Ticker.Stop()
	tp.k.Kill()
	tp.k.Wait()
}

// Defines the behaviour and action of any event on a Plugin. Use the
// DefaultEventHandler unless you want to add custom behaviour. For example,
// you could keep track of variables that are known globally to the plugin.
// (Every Plugin event goes through the same handler)
type EventHandler interface {
	HandleEvent(*Message, []string)
}

// Defines the behaviour and action of any event on a TimedPlugin. Use the
// DefaultTimedEventHandler unless you want to add custom behaviour.
type TimedEventHandler interface {
	HandleEvent()
}
