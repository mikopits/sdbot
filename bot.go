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

// Bot represents the entrypoint to all the necessary behaviour of the bot.
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
	BattleFormats         []string
	RecentBattles         chan *RecentBattles
	pccMutex              sync.Mutex
	ppcMutex              sync.Mutex
	semMutex              sync.Mutex
	semaphores            map[string]*sync.Mutex
}

// NewBot creates a new instance of the Bot struct. In doing so it creates a
// new Connection as well as adds a PrettyLogger to the Loggers that logs to
// os.Stderr.
func NewBot() *Bot {
	b := &Bot{
		Config:                readConfig(),
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
		Bot:   b,
		queue: make(chan string, 64),
	}
	loggers = NewLoggerList(&PrettyLogger{AnyLogger{Output: os.Stderr}})
	return b
}

// login connects to the Pokemon Showdown server.
func (b *Bot) login(msg *Message) {
	var res *http.Response
	var err error

	loginURL := "https://play.pokemonshowdown.com/action.php"

	if b.Config.Password == "" {
		res, err = http.Get(strings.Join([]string{
			loginURL,
			"?act=getassertion&userid=", Sanitize(b.Config.Nick),
			"&challengekeyid=", msg.Params[0],
			"&challenge=", msg.Params[1],
		}, ""))
	} else {
		res, err = http.PostForm(loginURL, url.Values{
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

// JoinRoom makes the bot join a room.
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

// LeaveRoom makes the bot leave a room.
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

// ErrPluginNameAlreadyRegistered is returned whenever a plugin with a particular
// name is attempted to be registered, but the name has previously been
// registered.
var ErrPluginNameAlreadyRegistered = errors.New("sdbot: plugin name was already in use (register under another name)")

// ErrPluginAlreadyRegistered is returned whenever the same plugin is attempted
// to be registered twice.
var ErrPluginAlreadyRegistered = errors.New("sdbot: plugin was already registered")

// RegisterPlugin registers a plugin under a name and starts listening on its event handler.
func (b *Bot) RegisterPlugin(p *Plugin, name string) error {
	for _, plugin := range b.Plugins {
		if plugin == p {
			Error(ErrPluginAlreadyRegistered)
			return ErrPluginAlreadyRegistered
		}
	}

	if b.PluginChatChannels[name] != nil {
		Error(ErrPluginNameAlreadyRegistered)
		return ErrPluginNameAlreadyRegistered
	}

	p.Bot = b
	p.Name = name

	// Load prefix and suffix from the config if none were provided.
	if p.Prefix == nil {
		p.Prefix = b.Config.PluginPrefix
	}
	if p.Suffix == nil {
		p.Suffix = b.Config.PluginSuffix
	}

	p.formatPrefixAndSuffix()
	Debugf("[on bot] Registering plugin `%s` listening on prefix `%v` and suffix `%v`", name, p.Prefix, p.Suffix)

	chatChannel := make(chan *Message, 64)
	privateChannel := make(chan *Message, 64)
	b.PluginChatChannels[name] = &chatChannel
	b.PluginPrivateChannels[name] = &privateChannel

	b.Plugins = append(b.Plugins, p)
	p.listen()
	return nil
}

// RegisterPlugins registers a slice of plugins in one call.
// The map should be formatted with pairs of "plugin name"=>*Plugin.
func (b *Bot) RegisterPlugins(plugins map[string]*Plugin) error {
	for name, p := range plugins {
		err := b.RegisterPlugin(p, name)
		CheckErr(err)
	}
	return nil
}

// RegisterTimedPlugin registers a timed plugin under the provided name.
// Timed plugins are not started until the bot is logged in.
func (b *Bot) RegisterTimedPlugin(tp *TimedPlugin, name string) error {
	for _, plugin := range b.TimedPlugins {
		if plugin.Name == tp.Name {
			Error(ErrPluginNameAlreadyRegistered)
			return ErrPluginNameAlreadyRegistered
		}
	}
	Debugf("[on bot] Registering timed plugin `%s` with period `%f`", name, tp.Period)
	tp.Bot = b
	b.TimedPlugins = append(b.TimedPlugins, tp)
	return nil
}

// UnregisterPlugin unregisters a plugin.
// Returns true if the plugin was successfully unregistered.
func (b *Bot) UnregisterPlugin(p *Plugin) bool {
	for i, plugin := range b.Plugins {
		if plugin == p {
			Debugf("[on bot] Unregistering plugin `%s`", p.Name)
			p.stopListening()
			delete(b.PluginChatChannels, p.Name)
			delete(b.PluginPrivateChannels, p.Name)
			b.Plugins = append(b.Plugins[:i], b.Plugins[i+1:]...)
			return true
		}
	}
	return false
}

// UnregisterPlugins unregisters all plugins.
func (b *Bot) UnregisterPlugins() {
	for _, plugin := range b.Plugins {
		b.UnregisterPlugin(plugin)
	}
}

// UnregisterTimedPlugin unregisters a timed plugin.
// Returns true if the plugin was successfully unregistered.
func (b *Bot) UnregisterTimedPlugin(tp *TimedPlugin) bool {
	for i, plugin := range b.TimedPlugins {
		if plugin.Name == tp.Name {
			Debugf("[on bot] Unregistering timed plugin `%s`", tp.Name)
			tp.stop()
			b.TimedPlugins = append(b.TimedPlugins[:i], b.TimedPlugins[i+1:]...)
			return true
		}
	}
	return false
}

// StartTimedPlugins starts all registered TimedPlugins.
func (b *Bot) StartTimedPlugins() {
	for _, tp := range b.TimedPlugins {
		tp.start()
	}
}

// StopTimedPlugins stops all registered TimedPlugins.
func (b *Bot) StopTimedPlugins() {
	for _, tp := range b.TimedPlugins {
		tp.stop()
	}
}

func (b *Bot) pluginChatChannelsWrite(s string, m *Message) {
	b.pccMutex.Lock()
	*b.PluginChatChannels[s] <- m
	b.pccMutex.Unlock()
}

func (b *Bot) pluginPrivateChannelsWrite(s string, m *Message) {
	b.ppcMutex.Lock()
	*b.PluginPrivateChannels[s] <- m
	b.ppcMutex.Unlock()
}

func (b *Bot) pluginChatChannelsRead(s string) chan *Message {
	b.pccMutex.Lock()
	defer b.pccMutex.Unlock()
	return *b.PluginChatChannels[s]
}

func (b *Bot) pluginPrivateChannelsRead(s string) chan *Message {
	b.ppcMutex.Lock()
	defer b.ppcMutex.Unlock()
	return *b.PluginPrivateChannels[s]
}

// Synchronize provides an API to keep your code thread-safe and concurrent.
// For example, if a plugin is going to write to a file, it would be a bad idea
// to have multiple threads with different state to try to access the file at
// the same time. So you should run such commands under Bot.Synchronize.
// The name is an arbitrary name you can choose for your mutex. The lambda is
// Then run in the mutex defined by the name. Choose unique names!
//
// Example:
// var doUnsafeAction = func() interface{} {
//   someActionThatMustBePerformedSequentially()
//   return nil
// }
// bot.Synchronize("uniqueIdentifierForUnsafeAction", &doUnsafeAction)
func (b *Bot) Synchronize(name string, lambda *func() interface{}) interface{} {
	b.semMutex.Lock()
	_, exists := b.semaphores[name]
	if !exists {
		b.semaphores[name] = &sync.Mutex{}
	}
	semaphore := b.semaphores[name]
	b.semMutex.Unlock()

	semaphore.Lock()
	defer semaphore.Unlock()
	return (*lambda)()
}

// Connect starts the bot and connects to the websocket.
func (b *Bot) Connect() {
	b.Connection.connect()
}

// Send queues a string onto the outgoing message queue.
func (b *Bot) Send(s string) {
	b.Connection.QueueMessage(s)
}
