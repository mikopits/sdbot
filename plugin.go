package sdbot

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Plugin is a command that triggers on some regexp and performs a task
// based on the message that triggered it as defined by its EventHandler.
// It can trigger on several different formats (consider prefix "." and
// command "cmd"). Note that cooldowns will not affect command syntax, but
// will ignore the commands hould it have been sent less than Cooldown time
// from the last instance.
//
// NewPlugin: ".cmd"
// NewPluginWithArgs:
//   1 arg:  ".cmd arg0"
//   2 args: ".cmd arg0, arg1" (space is optional)
//   n args: ".cmd arg0,arg1,arg2,arg3, ...,arg n" (space is optional)
//
// It is possible to define a Command with a regexp string. For example,
// a command of "y|n" will trigger on either y or n.
//
// Note that if a command is given as something like s(ay|peak) then the
// parsed args will be pushed back in the slice, and args[0] will contain
// the string of either "ay" or "peak", depending on which triggered the
// message.
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

// TimedPlugin structs will fire an event on a regular schedule defined by the
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

// DefaultEventHandler is the default event handler that you can make use of
// if you do not want to encapsulate your plugin with any custom behaviour.
// Refer to the example plugins to see how you can make use of this.
type DefaultEventHandler struct {
	Plugin *Plugin
}

// DefaultTimedEventHandler is the default event handler for a TimedPlugin.
// Refer to the example plugins to see how you can make use of this.
type DefaultTimedEventHandler struct {
	TimedPlugin *TimedPlugin
}

// NewPlugin (and its variants) define convenient ways to make new Plugin
// structs. You must add an event handler after creation. If you want
// to add custom prefixes or suffixes, you can do it with the Plugin.SetPrefix
// and Plugin.SetSuffix methods. Otherwise the bot will load prefixes and
// suffixes from your Config.
//
// NewPlugin in particular creates a Plugin that will trigger on a command.
func NewPlugin(cmd string) *Plugin {
	return &Plugin{
		Command: cmd,
	}
}

// NewPluginWithoutCommand creates a new Plugin that will trigger on every
// chat and private event.
func NewPluginWithoutCommand() *Plugin {
	return &Plugin{}
}

// NewPluginWithArgs creates a new Plugin that will trigger on a command, and
// will parse comma-separated arguments provided along with the command. It
// will not trigger unless an adequate amount of commas are present.
func NewPluginWithArgs(cmd string, numArgs int) *Plugin {
	return &Plugin{
		Command: cmd,
		NumArgs: numArgs,
	}
}

// NewPluginWithCooldown creates a new Plugin that will trigger on a command,
// but will not trigger if the last time it was used was not at least the
// provided time.Duration ago.
func NewPluginWithCooldown(cmd string, cooldown time.Duration) *Plugin {
	return &Plugin{
		Command:  cmd,
		Cooldown: cooldown,
	}
}

// NewPluginWithArgsAndCooldown creates a new Plugin with arguments and a
// cooldown as described by both NewPluginWithArgs and NewPluginWithCooldown.
func NewPluginWithArgsAndCooldown(cmd string, numArgs int, cooldown time.Duration) *Plugin {
	return &Plugin{
		Command:  cmd,
		NumArgs:  numArgs,
		Cooldown: cooldown,
	}
}

// NewTimedPlugin creates a new TimedPlugin that fires events on its
// TimedEventHandler every given period of time.Duration.
func NewTimedPlugin(period time.Duration) *TimedPlugin {
	return &TimedPlugin{
		Period: period,
	}
}

// NewDefaultEventHandler creates a new DefaultEventHandler.
// TODO Perhaps this isn't useful.
func NewDefaultEventHandler(p *Plugin) *DefaultEventHandler {
	return &DefaultEventHandler{Plugin: p}
}

// SetEventHandler sets the EventHandler of the Plugin.
// Allows you to use a custom EventHandler with any fields you want.
// The EventHandler of every Plugin MUST be set after the creation of a Plugin.
func (p *Plugin) SetEventHandler(eh EventHandler) {
	p.EventHandler = eh
}

// SetEventHandler sets the TimedEventHandler of the TimedPlugin.
// The TimedEventHandler of every TimedPlugin MUST be set after its creation.
func (tp *TimedPlugin) SetEventHandler(teh TimedEventHandler) {
	tp.TimedEventHandler = teh
}

// SetPrefix overrides the Plugin's default Prefix as read by the Config.
func (p *Plugin) SetPrefix(prefixes []string) {
	if len(prefixes) == 0 {
		return
	}

	regStr := "^(" + strings.Join(prefixes, "|") + ")"
	reg, err := regexp.Compile(regStr)
	CheckErr(err)

	p.Prefix = reg
}

// SetSuffix overrides the Plugin's default Suffix as read by the Config.
func (p *Plugin) SetSuffix(suffixes []string) {
	if len(suffixes) == 0 {
		return
	}

	regStr := "(" + strings.Join(suffixes, "|") + ")$"
	reg, err := regexp.Compile(regStr)
	CheckErr(err)

	p.Suffix = reg
}

// Formats the prefixes and suffixes into the regexp that will be used to match
// messages.
func (p *Plugin) formatPrefixAndSuffix() {
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

// Parse the message. Returns the arguments provided to the message.
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
func (p *Plugin) listen() {
	go func() {
		for {
			select {
			case m := <-p.Bot.pluginChatChannelsRead(p.Name):
				if !p.Bot.Config.IgnoreChatMessages && p.match(m) {
					args := p.parse(m)
					Debugf("[on plugin] Starting chat event handler goroutine for plugin `%s` with args `%+v`", p.Name, args)
					if m.Time.Sub(p.LastUsed) > p.Cooldown {
						p.LastUsed = m.Time
						go p.EventHandler.HandleEvent(m, args)
					}
				}
			case m := <-p.Bot.pluginPrivateChannelsRead(p.Name):
				if !p.Bot.Config.IgnorePrivateMessages && p.match(m) {
					args := p.parse(m)
					Debugf("[on plugin] Starting private event handler goroutine for plugin `%s` with args `%+v`", p.Name, args)
					if m.Time.Sub(p.LastUsed) > p.Cooldown {
						p.LastUsed = m.Time
						go p.EventHandler.HandleEvent(m, args)
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
func (p *Plugin) stopListening() {
	p.k.Kill()
	p.k.Wait()
}

// Starts a loop listening on the time.Ticker.
func (tp *TimedPlugin) start() {
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
func (tp *TimedPlugin) stop() {
	tp.Ticker.Stop()
	tp.k.Kill()
	tp.k.Wait()
}

// EventHandler defines the behaviour and action of any event on a Plugin. Use
// the DefaultEventHandler unless you want to add custom behaviour. For
// example, you could keep track of variables that are known globally to the
// plugin. (Every Plugin event goes through the same handler)
type EventHandler interface {
	HandleEvent(*Message, []string)
}

// TimedEventHandler defines the behaviour and action of any event on a
// TimedPlugin. Use the  DefaultTimedEventHandler unless you want to add custom
// behaviour.
type TimedEventHandler interface {
	HandleEvent()
}
