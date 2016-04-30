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
var Handlers = map[string]interface{}{
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

// Uses reflection to call a handler if it exists for the given command
func CallHandler(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
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

func onChallstr(msg *Message, events chan string) {
	Info(&Log, "Attempting to log in...")
	msg.Bot.Login(msg)
}

var AvatarSet bool = false
var once sync.Once

func onUpdateuser(msg *Message, events chan string) {
	switch msg.Params[1] {
	case "0":
		if msg.Bot.Config.Avatar > 0 && msg.Bot.Config.Avatar <= 294 {
			msg.Bot.Connection.QueueMessage("|/avatar " + strconv.Itoa(msg.Bot.Config.Avatar))
			AvatarSet = true
		}
	case "1":
		for _, r := range msg.Bot.Config.Rooms {
			room := FindRoomEnsured(r, msg.Bot)
			msg.Bot.JoinRoom(room)
		}
		// We have successfully logged in, start TimedPlugins.
		once.Do(func() { msg.Bot.StartTimedPlugins() })
	}
}

func onLeave(msg *Message, events chan string) {
	delete(msg.Bot.UserList, Sanitize(msg.User.Name))
}

func onJoin(msg *Message, events chan string) {
	if msg.User.Name == msg.Bot.Nick {
		onInit(msg, events)
	}

	FindUserEnsured(msg.User.Name, msg.Bot).AddAuth(msg.Room.Name, msg.Auth)
	FindRoomEnsured(msg.Room.Name, msg.Bot).AddUser(msg.User.Name)
}

func onNick(msg *Message, events chan string) {
	oldNick := Sanitize(msg.Params[0])
	if oldNick == Sanitize(msg.Bot.Nick) {
		msg.Bot.Nick = msg.User.Name
	} else {
		Rename(oldNick, msg.User.Name, msg.Bot)
	}
}

func onInit(msg *Message, events chan string) {
	// This may occur if the bot is redirected. Leave the room if it
	// is not a room it should be in, and try to rejoin any rooms that
	// it should be in.
	//
	// Note that a successful /join will trigger another init event.
	// So be careful to not cause an infinite loop.
	if !includes(msg.Bot.Config.Rooms, msg.Room.Name) {
		msg.Bot.LeaveRoom(FindRoomEnsured(msg.Room.Name, msg.Bot))
		// Try to join each of the config rooms, as you may have been redirected.
		// It is safe to call "/join [room]" if you are already in it, so there is
		// not really a need to check.
		for _, room := range msg.Bot.Config.Rooms {
			msg.Bot.JoinRoom(FindRoomEnsured(room, msg.Bot))
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

func onDeinit(msg *Message, events chan string) {
	// TODO Attempt to rejoin? Does the state need to be updated?
}

func onUsers(msg *Message, events chan string) {
	// Populate the room with its users and their auth levels.
	for _, user := range strings.Split(msg.Params[0], ",")[1:] {
		auth, nick := string(user[0]), user[1:]
		FindRoomEnsured(msg.Room.Name, msg.Bot).AddUser(nick)
		FindUserEnsured(nick, msg.Bot).AddAuth(msg.Room.Name, auth)
	}
}

func onPopup(msg *Message, events chan string) {
	if len(msg.Params) < 3 {
		return
	}

	// Handle bans
	if strings.Contains(msg.Params[2], "has banned you from the room") {
		reg, err := regexp.Compile("(?P<user>[^ ]+) has banned you from the room (?P<room>[^ ]*).</p><p>To appeal")
		if err != nil {
			Error(&Log, err)
		}
		match := reg.FindStringSubmatch(msg.Params[2])
		result := make(map[string]string)
		for i, name := range reg.SubexpNames() {
			if i != 0 {
				result[name] = match[i]
			}
		}
		user, room := result["user"], result["room"]
		Warn(&Log, "You have been banned from the room "+room+" by the user "+user)
		// TODO: This is a lazy solution, but the server no longer notifies you when
		// you are unbanned. Try to rejoin after a potential kick.
		time.Sleep(time.Second)
		msg.Bot.JoinRoom(&Room{Name: room})
	}
}

func onChat(msg *Message, events chan string) {
	if LoginTime == 0 {
		return
	}
	if msg.Message != "" && msg.Timestamp >= LoginTime {
		events <- "message"
		for _, channel := range msg.Bot.PluginChatChannels {
			*channel <- msg
		}
	}
}

func onPrivateMessage(msg *Message, events chan string) {
	events <- "private"
	for _, channel := range msg.Bot.PluginPrivateChannels {
		*channel <- msg
	}
}

func onTournament(msg *Message, events chan string) {
	// Tournaments are very noisy and send a lot of information we don't need,
	// only want to listen to tournament create, update, and start events. You
	// can still use the "tournament" event for other purposes.
	switch msg.Params[0] {
	case "create":
		fallthrough
	case "update":
		fallthrough
	case "start":
		events <- "tour"
	}
}

// Parse and store the current battle formats in Bot.BattleFormats
func onFormats(msg *Message, events chan string) {
	formatsStr := "|" + strings.Join(msg.Params, "|")
	formatsStr = regexp.MustCompile("[,#]").ReplaceAllString(formatsStr, "")
	formatsStr = regexp.MustCompile("\\|[0-9]+\\|[^|]+").ReplaceAllString(formatsStr, "")
	var formats []string
	for _, format := range strings.Split(formatsStr, "|") {
		sanitized := Sanitize(format)
		if len(sanitized) > 0 {
			formats = append(formats, sanitized[:len(sanitized)-1])
		}
	}
	//msg.Bot.BattleFormats = formats[1:]
	// TODO Verify that this does what I want it to.
	Debugf(&Log, "[on handlers] battle formats: %+v", msg.Bot.BattleFormats)
}

// Struct to hold the roomlist JSON unmarshalling.
// TODO Check the output of |/cmd roomlist and verify that this struct will
// indeed unmarshal the JSON.
type RecentBattles struct {
	Battles map[string]BattleInfo
}

type BattleInfo struct {
	FirstPlayer  string `json:"p1"`
	SecondPlayer string `json:"p2"`
	MinElo       string `json:"minElo"`
}

func onQueryResponse(msg *Message, events chan string) {
	// Populate the bot with "roomlist" information.
	if msg.Params[0] == "roomlist" {
		var recentBattles RecentBattles
		err := json.Unmarshal([]byte(msg.Params[1]), &recentBattles)
		CheckErr(err)
		msg.Bot.RecentBattles <- &recentBattles
	}
}
