package sdbot

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	LoginURL = "https://play.pokemonshowdown.com/action.php"
)

var Log Logger

type Bot struct {
	Config       *Config
	Connection   *Connection    // The websocket connection
	UserList     map[*User]bool // List of all users the bot knows about
	RoomList     map[*Room]bool // List of all rooms the bot knows about
	Rooms        map[*Room]bool // List of all the rooms the bot is in
	Nick         string         // The bot's username
	Plugins      []*Plugin      // List of all registered plugins
	TimedPlugins []*TimedPlugin // List of all registered timed plugins
}

// Creates a new bot instance.
func NewBot() *Bot {
	b := &Bot{
		Config:       ReadConfig(),
		UserList:     make(map[*User]bool),
		RoomList:     make(map[*Room]bool),
		Rooms:        make(map[*Room]bool),
		Plugins:      []*Plugin{},
		TimedPlugins: []*TimedPlugin{},
	}
	b.Nick = b.Config.Nick
	b.Connection = &Connection{
		Bot:       b,
		Connected: false,
		inQueue:   make(chan string, 128),
		outQueue:  make(chan string, 128),
	}
	Log = &PrettyLogger{AnyLogger{Output: os.Stderr}}
	return b
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
	if err != nil {
		Error(&Log, err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Error(&Log, err)
	}

	if b.Config.Password == "" {
		b.Connection.QueueMessage(strings.Join([]string{"/trn ", b.Config.Nick, ",0,", string(body)}, ""))
	} else {
		type LoginDetails struct {
			Assertion string
		}
		data := LoginDetails{}
		err = json.Unmarshal(body[1:], &data)
		if err != nil {
			Error(&Log, err)
		}

		b.Connection.QueueMessage(strings.Join([]string{"|/trn ", b.Config.Nick, ",0,", data.Assertion}, ""))
	}
}

// Joins a room.
func (b *Bot) JoinRoom(room *Room) {
	b.Rooms[room] = true
	b.RoomList[room] = true
	b.Connection.QueueMessage("|/join " + room.Name)
}

// Leaves a room.
func (b *Bot) LeaveRoom(room *Room) {
	delete(b.Rooms, room)
	b.RoomList[room] = true
	b.Connection.QueueMessage("|/leave " + room.Name)
}

// Register a plugin
// Return false if the plugin has already been registered.
// Return true if the plugin is successfully registered.
func (b *Bot) RegisterPlugin(p *Plugin) bool {
	for _, plugin := range b.Plugins {
		if plugin == p {
			return false
		}
	}
	p.Listen()
	b.Plugins = append(b.Plugins, p)
	return true
}

// Register a timed plugin
// Return false if the plugin has already been registered.
// Return true if the plugin is successfully registered.
func (b *Bot) RegisterTimedPlugin(tp *TimedPlugin) bool {
	for _, plugin := range b.TimedPlugins {
		if plugin == tp {
			return false
		}
	}
	b.TimedPlugins = append(b.TimedPlugins, tp)
	return true
}

// Unregister a plugin.
// Returns true if the plugin was successfully unregistered.
func (b *Bot) UnregisterPlugin(p *Plugin) bool {
	for i, plugin := range b.Plugins {
		if plugin == p {
			b.Plugins = append(b.Plugins[:i], b.Plugins[i+1:]...)
			return true
		}
	}
	return false
}

// Unregister a timed plugin.
// Returns true if the plugins was successfully unregistered.
func (b *Bot) UnregisterTimedPlugin(tp *TimedPlugin) bool {
	for i, plugin := range b.TimedPlugins {
		if plugin == tp {
			b.TimedPlugins = append(b.TimedPlugins[:i], b.TimedPlugins[i+1:]...)
			return true
		}
	}
	return false
}

// Start all registered TimedPlugins
func (b *Bot) StartTimedPlugins() {
	for _, tp := range b.TimedPlugins {
		tp.TimedEventHandler.Start()
	}
}

// Stop all registered TimedPlugins
func (b *Bot) StopTimedPlugins() {
	for _, tp := range b.TimedPlugins {
		tp.TimedEventHandler.Stop()
	}
}
