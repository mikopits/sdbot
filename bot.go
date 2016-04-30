package sdbot

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const (
	LoginURL = "https://play.pokemonshowdown.com/action.php"
)

// TODO Move this to Bot.Loggers
var Log Logger

// The Bot struct is the entrypoint to all the necessary behaviour of the bot.
// The bot runs its handlers in separate goroutines, so an API is provided to
// allow for thread-safe and concurrent access to the bot. See the Synchronize
// method for how this works.
type Bot struct {
	Config                *Config
	Connection            *Connection
	UserList              map[string]*User
	RoomList              map[string]*Room
	Rooms                 []string
	Nick                  string
	Plugins               []*Plugin
	TimedPlugins          []*TimedPlugin
	PluginChatChannels    map[string]*chan *Message
	PluginPrivateChannels map[string]*chan *Message
	Callback              *Callback
	Loggers               []Logger
	BattleFormats         []string
	RecentBattles         chan *RecentBattles
	mutex                 sync.Mutex
	semaphores            map[string]*sync.Mutex
}

// Creates a new bot instance.
func NewBot() *Bot {
	b := &Bot{
		Config:                ReadConfig(),
		UserList:              make(map[string]*User),
		RoomList:              make(map[string]*Room),
		Plugins:               []*Plugin{},
		TimedPlugins:          []*TimedPlugin{},
		PluginChatChannels:    make(map[string]*chan *Message, 64),
		PluginPrivateChannels: make(map[string]*chan *Message, 64),
		semaphores:            make(map[string]*sync.Mutex),
		RecentBattles:         make(chan *RecentBattles, 1),
	}
	b.Nick = b.Config.Nick
	b.Connection = &Connection{
		Bot:       b,
		Connected: false,
		outQueue:  make(chan string, 64),
	}
	b.Callback = &Callback{Bot: b}
	Log = &PrettyLogger{AnyLogger{Output: os.Stderr}}
	return b
}

// TODO Refactor the logging system to log to every logger in Bot.Loggers
// whenever a log function is called. Allows for custom loggers with file
// outputs.
func (b *Bot) AddLogger(lo Logger) {
	b.Loggers = append(b.Loggers, lo)
}

// Connects to the Pokemon Showdown server.
func (b *Bot) Login(msg *Message) {
	var res *http.Response
	var err error

	if b.Config.Password == "" {
		res, err = http.Get(strings.Join([]string{
			LoginURL,
			"?act=getassertion&userid=", Sanitize(b.Config.Nick),
			"&challengekeyid=", msg.Params[0],
			"&challenge=", msg.Params[1],
		}, ""))
	} else {
		res, err = http.PostForm(LoginURL, url.Values{
			"act":            {"login"},
			"name":           {b.Config.Nick},
			"pass":           {b.Config.Password},
			"challengekeyid": {msg.Params[0]},
			"challenge":      {msg.Params[1]},
		})
	}
	CheckErr(err)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	CheckErr(err)

	if b.Config.Password == "" {
		b.Connection.QueueMessage(strings.Join([]string{"/trn ", b.Config.Nick, ",0,", string(body)}, ""))
	} else {
		type LoginDetails struct {
			Assertion string
		}
		data := LoginDetails{}
		err = json.Unmarshal(body[1:], &data)
		CheckErr(err)

		b.Connection.QueueMessage(strings.Join([]string{"|/trn ", b.Config.Nick, ",0,", data.Assertion}, ""))
	}
}

// Joins a room.
func (b *Bot) JoinRoom(room *Room) {
	var joinRoom = func() interface{} {
		b.Rooms = append(b.Rooms, room.Name)
		if b.RoomList[room.Name] != nil {
			b.RoomList[room.Name] = room
		}
		return nil
	}

	b.Synchronize("room", &joinRoom)

	b.Connection.QueueMessage("|/join " + room.Name)
}

// Leaves a room.
func (b *Bot) LeaveRoom(room *Room) {
	var leaveRoom = func() interface{} {
		for i, r := range b.Rooms {
			if room.Name == r {
				b.Rooms = append(b.Rooms[:i], b.Rooms[i+1:]...)
			}
		}
		delete(b.RoomList, room.Name)
		return nil
	}

	b.Synchronize("room", &leaveRoom)

	b.Connection.QueueMessage("|/leave " + room.Name)
}

var ErrPluginNameAlreadyRegistered = errors.New("sdbot: plugin name was already in use (register under another name)")
var ErrPluginAlreadyRegistered = errors.New("sdbot: plugin was already registered")

// Registers a plugin under a name and starts listening on its event handler.
func (b *Bot) RegisterPlugin(p *Plugin, name string) error {
	for _, plugin := range b.Plugins {
		if plugin == p {
			Error(&Log, ErrPluginAlreadyRegistered)
			return ErrPluginAlreadyRegistered
		}
	}

	if b.PluginChatChannels[name] != nil {
		Error(&Log, ErrPluginNameAlreadyRegistered)
		return ErrPluginNameAlreadyRegistered
	} else {
		p.Bot = b
		p.Name = name

		// Load prefix and suffix from the config if none were provided.
		if p.Prefix == nil {
			p.Prefix = b.Config.PluginPrefix
		}
		if p.Suffix == nil {
			p.Suffix = b.Config.PluginSuffix
		}

		p.FormatPrefixAndSuffix()
		Debugf(&Log, "[on bot] Registering plugin `%s` listening on prefix `%v` and suffix `%v`", name, p.Prefix, p.Suffix)

		chatChannel := make(chan *Message, 64)
		privateChannel := make(chan *Message, 64)
		b.PluginChatChannels[name] = &chatChannel
		b.PluginPrivateChannels[name] = &privateChannel
	}

	b.Plugins = append(b.Plugins, p)
	p.Listen()
	return nil
}

// Register a slice of plugins in one call.
// The map should be formatted with pairs of "plugin name"=>*Plugin.
func (b *Bot) RegisterPlugins(plugins map[string]*Plugin) error {
	for name, p := range plugins {
		err := b.RegisterPlugin(p, name)
		CheckErr(err)
	}
	return nil
}

// Registers a timed plugin under the provided name. Timed plugins are not
// Started until the bot is logged in.
func (b *Bot) RegisterTimedPlugin(tp *TimedPlugin, name string) error {
	for _, plugin := range b.TimedPlugins {
		if plugin.Name == tp.Name {
			Error(&Log, ErrPluginNameAlreadyRegistered)
			return ErrPluginNameAlreadyRegistered
		}
	}
	Debugf(&Log, "[on bot] Registering timed plugin `%s` with period `%f`", name, tp.Period)
	tp.Bot = b
	b.TimedPlugins = append(b.TimedPlugins, tp)
	return nil
}

// Unregister a plugin.
// Returns true if the plugin was successfully unregistered.
func (b *Bot) UnregisterPlugin(p *Plugin) bool {
	for i, plugin := range b.Plugins {
		if plugin == p {
			Debugf(&Log, "[on bot] Unregistering plugin `%s`", p.Name)
			p.StopListening()
			delete(b.PluginChatChannels, p.Name)
			delete(b.PluginPrivateChannels, p.Name)
			b.Plugins = append(b.Plugins[:i], b.Plugins[i+1:]...)
			return true
		}
	}
	return false
}

// Unregister all plugins.
func (b *Bot) UnregisterPlugins() {
	for _, plugin := range b.Plugins {
		b.UnregisterPlugin(plugin)
	}
}

// Unregister a timed plugin.
// Returns true if the plugin was successfully unregistered.
func (b *Bot) UnregisterTimedPlugin(tp *TimedPlugin) bool {
	for i, plugin := range b.TimedPlugins {
		if plugin.Name == tp.Name {
			Debugf(&Log, "[on bot] Unregistering timed plugin `%s`", tp.Name)
			tp.Stop()
			b.TimedPlugins = append(b.TimedPlugins[:i], b.TimedPlugins[i+1:]...)
			return true
		}
	}
	return false
}

// Start all registered TimedPlugins
func (b *Bot) StartTimedPlugins() {
	for _, tp := range b.TimedPlugins {
		tp.Start()
	}
}

// Stop all registered TimedPlugins
func (b *Bot) StopTimedPlugins() {
	for _, tp := range b.TimedPlugins {
		tp.Stop()
	}
}

// Provides an API to keep your code thread-safe and concurrent.
// For example, if a plugin is going to write to a file, it would be a bad idea
// to have multiple threads with different state to try to access the file at
// the same time. So you should run such commands under Bot.Synchronize.
// The name is an arbitrary name you can choose for your mutex. The lambda is
// Then run in the mutex defined by the name. Choose unique names!
//
// Example:
// var doUnsafeAction = func() interface{} {
//   // Some action that must be performed sequentially.
//   return nil
// }
// bot.Synchronize("uniqueIdentifierForUnsafeAction", &doUnsafeAction)
func (b *Bot) Synchronize(name string, lambda *func() interface{}) interface{} {
	b.mutex.Lock()
	_, exists := b.semaphores[name]
	if !exists {
		b.semaphores[name] = &sync.Mutex{}
	}
	semaphore := b.semaphores[name]
	b.mutex.Unlock()

	semaphore.Lock()
	defer semaphore.Unlock()
	return (*lambda)()
}
