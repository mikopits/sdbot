package sdbot

import (
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Define function handlers to call depending on the command we get.
var handlers = map[string]interface{}{
	"challstr":      onChallstr,
	"updateuser":    onUpdateuser,
	"l":             onLeave,
	"j":             onJoin,
	"n":             onNick,
	"init":          onInit,
	"deinit":        onDeinit,
	"users":         onUsers,
	"popup":         onPopup,
	"c:":            onChat,
	"pm":            onPrivateMessage,
	"tournament":    onTournament,
	"formats":       onFormats,
	"queryresponse": onQueryResponse,
}

// callHandler uses reflection to call a handler if it exists for the given command
func callHandler(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if !f.IsValid() {
		return
	}

	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func onChallstr(m *Message) {
	Info("Attempting to log in...")
	m.Bot.login(m)
}

// AvatarSet is true if we have set the avater, and false otherwise.
var AvatarSet bool = false
var once sync.Once

func onUpdateuser(m *Message) {
	switch m.Params[1] {
	case "0":
		if m.Bot.Config.Avatar > 0 && m.Bot.Config.Avatar <= 294 {
			m.Bot.Connection.QueueMessage("|/avatar " + strconv.Itoa(m.Bot.Config.Avatar))
			AvatarSet = true
		}
	case "1":
		for _, r := range m.Bot.Config.Rooms {
			room := FindRoomEnsured(r, m.Bot)
			m.Bot.JoinRoom(room)
		}
		// We have successfully logged in, start TimedPlugins.
		once.Do(func() { m.Bot.StartTimedPlugins() })
	}
}

func onLeave(msg *Message) {
	FindRoomEnsured(msg.Room.Name, msg.Bot).RemoveUser(msg.User.Name)
	delete(msg.Bot.UserList, Sanitize(msg.User.Name))
}

func onJoin(msg *Message) {
	if msg.User.Name == msg.Bot.Nick {
		onInit(msg)
	}

	FindUserEnsured(msg.User.Name, msg.Bot).AddAuth(msg.Room.Name, msg.Auth)
	FindRoomEnsured(msg.Room.Name, msg.Bot).AddUser(msg.User.Name)
}

func onNick(m *Message) {
	oldNick := Sanitize(m.Params[0])
	if oldNick == Sanitize(m.Bot.Nick) {
		m.Bot.Nick = m.User.Name
	} else {
		Rename(oldNick, m.User.Name, m.Bot)
	}
}

func onInit(m *Message) {
	// This may occur if the bot is redirected. Leave the room if it
	// is not a room it should be in, and try to rejoin any rooms that
	// it should be in.
	//
	// Note that a successful /join will trigger another init event.
	// So be careful to not cause an infinite loop.
	if !includes(m.Bot.Config.Rooms, m.Room.Name) {
		m.Bot.LeaveRoom(FindRoomEnsured(m.Room.Name, m.Bot))
		// Try to join each of the config rooms, as you may have been redirected.
		// It is safe to call "/join [room]" if you are already in it, so there is
		// not really a need to check.
		for _, room := range m.Bot.Config.Rooms {
			m.Bot.JoinRoom(FindRoomEnsured(room, m.Bot))
		}
	}
}

func includes(a []string, s string) bool {
	for _, e := range a {
		if e == s {
			return true
		}
	}
	return false
}

func onDeinit(m *Message) {
	// TODO Attempt to rejoin? Does the state need to be updated?
}

func onUsers(m *Message) {
	// Populate the room with its users and their auth levels.
	for _, user := range strings.Split(m.Params[0], ",")[1:] {
		auth, nick := string(user[0]), user[1:]
		FindRoomEnsured(m.Room.Name, m.Bot).AddUser(nick)
		FindUserEnsured(nick, m.Bot).AddAuth(m.Room.Name, auth)
	}
}

func onPopup(m *Message) {
	if len(m.Params) < 3 {
		return
	}

	// Handle bans
	if strings.Contains(m.Params[2], "has banned you from the room") {
		reg, err := regexp.Compile("<p>(?P<user>[^ ]+) has banned you from the room (?P<room>[^ ]*).</p><p>To appeal")
		CheckErr(err)
		match := reg.FindStringSubmatch(m.Params[2])
		result := make(map[string]string)
		for i, name := range reg.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		user, room := result["user"], result["room"]
		Warn("You have been banned from the room " + room + " by the user " + user)
		// This is a lazy solution, but the server no longer notifies you when
		// you are unbanned. Try to rejoin after a potential kick.
		// TODO Start a ticker that will attempt to join every some period of time that
		// stops itself once the room has been joined.
		time.Sleep(time.Second)
		m.Bot.JoinRoom(&Room{Name: room})
	}
}

func onChat(m *Message) {
	if LoginTime[m.Room.Name] == 0 {
		return
	}
	if m.Message != "" && m.Timestamp >= LoginTime[m.Room.Name] {
		for name, _ := range m.Bot.PluginChatChannels {
			m.Bot.pluginChatChannelsWrite(name, m)
		}
	}
}

func onPrivateMessage(m *Message) {
	for name, _ := range m.Bot.PluginPrivateChannels {
		m.Bot.pluginPrivateChannelsWrite(name, m)
	}
}

func onTournament(m *Message) {
	// TODO
}

// Parse and store the current battle formats in Bot.BattleFormats
func onFormats(m *Message) {
	formatsStr := "|" + strings.Join(m.Params, "|")
	formatsStr = regexp.MustCompile("[,#]").ReplaceAllString(formatsStr, "")
	formatsStr = regexp.MustCompile("\\|[0-9]+\\|[^|]+").ReplaceAllString(formatsStr, "")
	var formats []string
	var sanitized string
	for _, format := range strings.Split(formatsStr, "|") {
		sanitized = Sanitize(format)
		if len(sanitized) > 0 {
			formats = append(formats, sanitized[:len(sanitized)-1])
		}
	}
	m.Bot.BattleFormats = formats[1:]
	// FIXME This doesn't parse the battle formats correctly.
	Debugf("[on handlers] battle formats: %+v", m.Bot.BattleFormats)
}

// RecentBattles holds the roomlist JSON unmarshalling battle map.
type RecentBattles struct {
	Battles map[string]BattleInfo
}

// BattleInfo holds the roomlist JSON unmarshalling for each battle.
type BattleInfo struct {
	FirstPlayer  string `json:"p1"`
	SecondPlayer string `json:"p2"`
	MinElo       string `json:"minElo"`
}

func onQueryResponse(m *Message) {
	// Populate the bot with "roomlist" information.
	if m.Params[0] == "roomlist" {
		var recentBattles RecentBattles
		err := json.Unmarshal([]byte(m.Params[1]), &recentBattles)
		CheckErr(err)
		m.Bot.RecentBattles <- &recentBattles
	}
}
